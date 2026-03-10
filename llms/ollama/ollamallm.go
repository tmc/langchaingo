package ollama

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
)

const (
	// CloudURL is the default Ollama Cloud API endpoint.
	// Use this URL when connecting to Ollama Cloud service with an API key.
	CloudURL = "https://ollama.com"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
	ErrPullError           = errors.New("ollama model pull error")
	ErrPullTimeout         = errors.New("ollama model pull deadline exceeded")
)

// authRoundTripper is a proxy HTTP RoundTripper that adds Authorization header to requests.
// It wraps an existing RoundTripper and adds the API key as a Bearer token.
type authRoundTripper struct {
	apiKey    string
	transport http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface.
// It adds the Authorization header and delegates to the wrapped transport.
func (a *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	clonedReq := req.Clone(req.Context())

	// Add Authorization header with Bearer token
	clonedReq.Header.Set("Authorization", "Bearer "+a.apiKey)

	// Use the wrapped transport to perform the actual request
	return a.transport.RoundTrip(clonedReq)
}

// LLM is a ollama LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *api.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new ollama LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{
		ollamaServerURL: envconfig.Host(),
		httpClient:      http.DefaultClient,
	}
	for _, opt := range opts {
		opt(&o)
	}

	// Check for API key in environment variable if not set explicitly
	if o.apiKey == "" {
		if envKey := os.Getenv("OLLAMA_API_KEY"); envKey != "" {
			o.apiKey = envKey
		}
	}

	// If API key is provided, wrap the HTTP client with auth transport
	httpClient := o.httpClient
	if o.apiKey != "" {
		// Get the underlying transport from the HTTP client
		transport := httpClient.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}

		// Create a new HTTP client with the auth transport
		httpClient = &http.Client{
			Transport: &authRoundTripper{
				apiKey:    o.apiKey,
				transport: transport,
			},
			CheckRedirect: httpClient.CheckRedirect,
			Jar:           httpClient.Jar,
			Timeout:       httpClient.Timeout,
		}
	}

	client := api.NewClient(o.ollamaServerURL, httpClient)

	return &LLM{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// override LLM model if set as llms.CallOption
	model := o.getModel(opts)

	// convert messages to Ollama format
	chatMsgs, err := o.prepareMessages(messages)
	if err != nil {
		return nil, err
	}

	// pull model if enabled
	if o.options.pullModel {
		if err := o.pullModelIfNeeded(ctx, model); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPullError, err)
		}
	}

	req, err := o.createChatRequest(model, chatMsgs, opts)
	if err != nil {
		return nil, err
	}

	if err := o.processTools(req, opts.Tools); err != nil {
		return nil, err
	}

	resp, err := o.handleChat(ctx, req, opts)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	response := o.createContentResponse(resp)

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// getModel determines which model to use based on options and defaults.
func (o *LLM) getModel(opts llms.CallOptions) string {
	if model := opts.GetModel(); model != "" {
		return model
	}
	return o.options.model
}

// prepareMessages converts LangChain message format to Ollama format.
func (o *LLM) prepareMessages(messages []llms.MessageContent) ([]api.Message, error) {
	chatMsgs := make([]api.Message, 0, len(messages))
	for _, mc := range messages {
		msg, err := o.convertMessageContent(mc)
		if err != nil {
			return nil, err
		}
		chatMsgs = append(chatMsgs, msg)
	}
	return chatMsgs, nil
}

// convertMessageContent converts a single message content to Ollama format.
func (o *LLM) convertMessageContent(mc llms.MessageContent) (api.Message, error) {
	// Our input is a sequence of MessageContent, each of which potentially has
	// a sequence of Part that could be text, images etc.
	// We have to convert it to a format Ollama undestands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	msg := api.Message{Role: typeToRole(mc.Role)}

	// Look at all the parts in mc; expect to find a single Text part and
	// any number of binary parts.
	var text string
	foundText := false
	var images []api.ImageData
	var toolCalls []api.ToolCall

	for _, p := range mc.Parts {
		switch pt := p.(type) {
		case llms.TextContent:
			if foundText {
				text += "\n\nnext part of text\n\n" + pt.Text
			} else {
				foundText = true
				text = pt.Text
			}
		case llms.BinaryContent:
			images = append(images, pt.Data)
		case llms.ToolCall:
			tc, err := o.convertToolCall(pt)
			if err != nil {
				return api.Message{}, err
			}
			toolCalls = append(toolCalls, tc)
		case llms.ToolCallResponse:
			text = pt.Content
		default:
			return api.Message{}, errors.New("only support Text and BinaryContent parts right now")
		}
	}

	msg.Content = text
	msg.Images = images
	msg.ToolCalls = toolCalls
	return msg, nil
}

// convertToolCall converts LangChain tool call to Ollama format.
func (o *LLM) convertToolCall(toolCall llms.ToolCall) (api.ToolCall, error) {
	tc := api.ToolCall{
		Function: api.ToolCallFunction{
			Name:  toolCall.FunctionCall.Name,
			Index: parseToolCallID(toolCall.ID),
		},
	}

	err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &tc.Function.Arguments)
	if err != nil {
		return api.ToolCall{}, fmt.Errorf("error unmarshalling tool call arguments: %w", err)
	}

	return tc, nil
}

// createChatRequest creates a chat request with the given parameters.
func (o *LLM) createChatRequest(model string, messages []api.Message, opts llms.CallOptions) (*api.ChatRequest, error) {
	format := o.options.format
	if opts.JSONMode {
		format = "json"
	}

	// Get our ollamaOptions from llms.CallOptions
	ollamaOptions, err := makeOllamaOptionsFromOptions(o.options.ollamaOptions, opts)
	if err != nil {
		return nil, fmt.Errorf("error creating ollama options: %w", err)
	}

	stream := opts.StreamingFunc != nil

	req := &api.ChatRequest{
		Model:    model,
		Format:   json.RawMessage(fmt.Sprintf(`"%s"`, format)),
		Messages: messages,
		Options:  ollamaOptions,
		Stream:   &stream,
		Tools:    make(api.Tools, len(opts.Tools)),
	}

	keepAlive := o.options.keepAlive
	if keepAlive != nil {
		req.KeepAlive = &api.Duration{Duration: *keepAlive}
	}

	return req, nil
}

// processTools adds tools to the chat request.
func (o *LLM) processTools(req *api.ChatRequest, tools []llms.Tool) error {
	for i := range tools {
		jt, err := json.Marshal(tools[i])
		if err != nil {
			return fmt.Errorf("error marshalling tool: %w", err)
		}

		var tool api.Tool
		err = json.Unmarshal(jt, &tool)
		if err != nil {
			return fmt.Errorf("error unmarshalling tool: %w", err)
		}

		req.Tools[i] = tool
	}

	return nil
}

// handleChat sends the chat request and processes the streaming response.
func (o *LLM) handleChat(ctx context.Context, req *api.ChatRequest, opts llms.CallOptions) (api.ChatResponse, error) {
	defer streaming.CallWithDone(ctx, opts.StreamingFunc) //nolint:errcheck

	var (
		resp              api.ChatResponse
		streamedResponse  string
		streamedToolCalls []api.ToolCall
	)

	splitter := reasoning.NewChunkContentSplitter()
	fn := func(response api.ChatResponse) error {
		textContent, reasoningContent := splitter.Split(response.Message.Content)
		if opts.StreamingFunc != nil {
			reasoning := &reasoning.ContentReasoning{Content: reasoningContent}
			if err := streaming.CallWithReasoning(ctx, opts.StreamingFunc, reasoning); err != nil {
				return fmt.Errorf("error calling streaming reasoning: %w", err)
			}
			if err := streaming.CallWithText(ctx, opts.StreamingFunc, textContent); err != nil {
				return fmt.Errorf("error calling streaming text: %w", err)
			}
		}

		for _, tc := range response.Message.ToolCalls {
			toolCallID := makeToolCallID(tc.Function.Index, tc.Function.Name)
			toolCallArgs, err := json.Marshal(tc.Function.Arguments)
			if err != nil {
				return fmt.Errorf("error marshalling tool call '%s' arguments: %w", toolCallID, err)
			}
			toolCall := streaming.NewToolCall(toolCallID, tc.Function.Name, string(toolCallArgs))
			if err := streaming.CallWithToolCall(ctx, opts.StreamingFunc, toolCall); err != nil {
				return fmt.Errorf("error calling streaming tool call '%s': %w", toolCallID, err)
			}
		}

		if response.Message.Content != "" {
			streamedResponse += response.Message.Content
		}
		if len(response.Message.ToolCalls) > 0 {
			streamedToolCalls = append(streamedToolCalls, response.Message.ToolCalls...)
		}

		if req.Stream == nil || !*req.Stream || response.Done {
			resp = response
			resp.Message = api.Message{
				Role:      "assistant",
				Content:   streamedResponse,
				ToolCalls: streamedToolCalls,
			}
		}
		return nil
	}

	err := o.client.Chat(ctx, req, fn)
	return resp, err
}

// createContentResponse creates a LangChain content response from Ollama response.
func (o *LLM) createContentResponse(resp api.ChatResponse) *llms.ContentResponse {
	reasoning, content := reasoning.SplitContentWithReasoning(resp.Message.Content)
	choices := []*llms.ContentChoice{
		{
			Content:    content,
			Reasoning:  reasoning,
			StopReason: resp.DoneReason,
			GenerationInfo: map[string]any{
				"CompletionTokens": resp.EvalCount,
				"PromptTokens":     resp.PromptEvalCount,
				"TotalTokens":      resp.EvalCount + resp.PromptEvalCount,
			},
		},
	}

	for _, tc := range resp.Message.ToolCalls {
		choices[0].ToolCalls = append(choices[0].ToolCalls, llms.ToolCall{
			ID:   makeToolCallID(tc.Function.Index, tc.Function.Name),
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments.String(),
			},
		})
	}

	return &llms.ContentResponse{Choices: choices}
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	// Pull model if enabled
	if o.options.pullModel {
		if err := o.pullModelIfNeeded(ctx, o.options.model); err != nil {
			return nil, err
		}
	}

	embeddings := [][]float32{}

	for _, input := range inputTexts {
		req := &api.EmbeddingRequest{
			Prompt: input,
			Model:  o.options.model,
		}
		if o.options.keepAlive != nil {
			req.KeepAlive = &api.Duration{Duration: *o.options.keepAlive}
		}

		eResp, err := o.client.Embeddings(ctx, req)
		if err != nil {
			return nil, err
		}

		if len(eResp.Embedding) == 0 {
			return nil, ErrEmptyResponse
		}

		embedding := make([]float32, len(eResp.Embedding))
		for i, val := range eResp.Embedding {
			embedding[i] = float32(val)
		}

		embeddings = append(embeddings, embedding)
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrIncompleteEmbedding
	}

	return embeddings, nil
}

func typeToRole(typ llms.ChatMessageType) string {
	switch typ {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeHuman:
		fallthrough
	case llms.ChatMessageTypeGeneric:
		return "user"
	case llms.ChatMessageTypeFunction:
		return "function"
	case llms.ChatMessageTypeTool:
		return "tool"
	default:
		return ""
	}
}

func makeOllamaOptionsFromOptions(ollamaOptions api.Options, opts llms.CallOptions) (map[string]any, error) {
	// Load back CallOptions as ollamaOptions
	ollamaOptions.NumPredict = opts.GetMaxTokens()
	ollamaOptions.Temperature = float32(opts.GetTemperature())
	ollamaOptions.Stop = opts.StopWords
	ollamaOptions.TopK = opts.GetTopK()
	ollamaOptions.TopP = float32(opts.GetTopP())
	ollamaOptions.Seed = opts.GetSeed()
	ollamaOptions.RepeatPenalty = float32(opts.GetRepetitionPenalty())
	ollamaOptions.FrequencyPenalty = float32(opts.GetFrequencyPenalty())
	ollamaOptions.PresencePenalty = float32(opts.GetPresencePenalty())

	os, err := json.Marshal(ollamaOptions)
	if err != nil {
		return nil, fmt.Errorf("error marshalling ollama options: %w", err)
	}

	var result map[string]any
	err = json.Unmarshal(os, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ollama options: %w", err)
	}

	return result, nil
}

// pullModelIfNeeded pulls the model if it's not already available.
func (o *LLM) pullModelIfNeeded(ctx context.Context, model string) error {
	// Try to use the model first. If it fails with a model not found error,
	// then pull the model.
	// This is a simple implementation. In production, you might want to
	// implement a more sophisticated check (e.g., using a list endpoint).

	// Apply timeout if configured
	pullCtx := ctx
	if o.options.pullTimeout > 0 {
		var cancel context.CancelFunc
		pullCtx, cancel = context.WithTimeoutCause(ctx, o.options.pullTimeout, ErrPullTimeout)
		defer func() {
			if cancel != nil {
				cancel()
			}
		}()
	}

	// For now, we'll just pull the model without checking.
	// This ensures the model is available but may result in unnecessary pulls.
	stream := false
	req := &api.PullRequest{
		Model:  model,
		Stream: &stream,
	}
	progress := func(_ api.ProgressResponse) error {
		return nil
	}
	if o.options.pullProgressFunc != nil {
		progress = o.options.pullProgressFunc
	}

	err := o.client.Pull(pullCtx, req, progress)
	if err != nil {
		// Check if the error is due to context timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		// Check if the context has a cause
		if cause := context.Cause(pullCtx); cause != nil {
			return fmt.Errorf("%w: %w", cause, err)
		}
	}
	return err
}

func makeToolCallID(index int, name string) string { //nolint:gosec
	hash := crc32.NewIEEE().Sum([]byte(name))
	encHash := hex.EncodeToString(hash)
	return fmt.Sprintf("ollama-%s-%d", encHash, index)
}

func parseToolCallID(id string) int { //nolint:gosec
	fallback := func() int {
		hash := crc32.NewIEEE()
		hash.Write([]byte(id))
		return int(hash.Sum32())
	}

	parts := strings.Split(id, "-")
	if len(parts) != 3 {
		return fallback()
	}

	index, err := strconv.Atoi(parts[2])
	if err != nil {
		return fallback()
	}

	return index
}
