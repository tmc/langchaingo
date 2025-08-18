package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ollama/ollama/api"
	"github.com/ollama/ollama/envconfig"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
	ErrPullError           = errors.New("ollama model pull error")
	ErrPullTimeout         = errors.New("ollama model pull deadline exceeded")
)

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

	client := api.NewClient(o.ollamaServerURL, o.httpClient)

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

	// Override LLM model if set as llms.CallOption
	model := o.options.model
	if opts.Model != "" {
		model = opts.Model
	}

	// Pull model if enabled
	if o.options.pullModel {
		if err := o.pullModelIfNeeded(ctx, model); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrPullError, err)
		}
	}

	// Our input is a sequence of MessageContent, each of which potentially has
	// a sequence of Part that could be text, images etc.
	// We have to convert it to a format Ollama undestands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := make([]api.Message, 0, len(messages))
	for _, mc := range messages {
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
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text
			case llms.BinaryContent:
				images = append(images, pt.Data)
			case llms.ToolCall:
				tc := api.ToolCall{
					Function: api.ToolCallFunction{
						Name: pt.FunctionCall.Name,
					},
				}

				var err error
				tc.Function.Index, err = strconv.Atoi(pt.ID)
				if err != nil {
					return nil, fmt.Errorf("error converting tool call ID to int: %w", err)
				}

				err = json.Unmarshal([]byte(pt.FunctionCall.Arguments), &tc.Function.Arguments)
				if err != nil {
					return nil, fmt.Errorf("error unmarshalling tool call arguments: %w", err)
				}

				toolCalls = append(toolCalls, tc)
			case llms.ToolCallResponse:
				text = pt.Content
			default:
				return nil, errors.New("only support Text and BinaryContent parts right now")
			}
		}

		msg.Content = text
		msg.Images = images
		msg.ToolCalls = toolCalls
		chatMsgs = append(chatMsgs, msg)
	}

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
		Messages: chatMsgs,
		Options:  ollamaOptions,
		Stream:   &stream,
		Think:    o.options.thinking,
		Tools:    make(api.Tools, len(opts.Tools)),
	}

	for i := range opts.Tools {
		jt, err := json.Marshal(opts.Tools[i])
		if err != nil {
			return nil, fmt.Errorf("error marshalling tool: %w", err)
		}

		var tool api.Tool
		err = json.Unmarshal(jt, &tool)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling tool: %w", err)
		}

		req.Tools[i] = tool
	}

	keepAlive := o.options.keepAlive
	if keepAlive > 0 {
		req.KeepAlive = &api.Duration{Duration: keepAlive}
	}

	var fn api.ChatResponseFunc
	streamedResponse := ""
	var streamedToolCalls []api.ToolCall
	var resp api.ChatResponse

	fn = func(response api.ChatResponse) error {
		if opts.StreamingFunc != nil && response.Message.Content != "" {
			if err := opts.StreamingFunc(ctx, []byte(response.Message.Content)); err != nil {
				return err
			}
		}
		if response.Message.Content != "" {
			streamedResponse += response.Message.Content
		}
		streamedToolCalls = append(streamedToolCalls, response.Message.ToolCalls...)

		rs := req.Stream != nil && *req.Stream
		if !rs || response.Done {
			resp = response
			resp.Message = api.Message{
				Role:      "assistant",
				Content:   streamedResponse,
				ToolCalls: streamedToolCalls,
			}
		}
		return nil
	}

	err = o.client.Chat(ctx, req, fn)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	choices := []*llms.ContentChoice{
		{
			Content:    resp.Message.Content,
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
			ID:   fmt.Sprintf("%d", tc.Function.Index),
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: tc.Function.Arguments.String(),
			},
		})
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
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
		if o.options.keepAlive > 0 {
			req.KeepAlive = &api.Duration{Duration: o.options.keepAlive}
		}

		eResp, err := o.client.Embeddings(ctx, req)
		if err != nil {
			return nil, err
		}

		if len(eResp.Embedding) == 0 {
			return nil, ErrEmptyResponse
		}

		embedding := make([]float32, len(eResp.Embedding))
		for i := range eResp.Embedding {
			embedding = append(embedding, float32(eResp.Embedding[i]))
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
	}
	return ""
}

func makeOllamaOptionsFromOptions(ollamaOptions api.Options, opts llms.CallOptions) (map[string]any, error) {
	// Load back CallOptions as ollamaOptions
	ollamaOptions.NumPredict = opts.MaxTokens
	ollamaOptions.Temperature = float32(opts.Temperature)
	ollamaOptions.Stop = opts.StopWords
	ollamaOptions.TopK = opts.TopK
	ollamaOptions.TopP = float32(opts.TopP)
	ollamaOptions.Seed = opts.Seed
	ollamaOptions.RepeatPenalty = float32(opts.RepetitionPenalty)
	ollamaOptions.FrequencyPenalty = float32(opts.FrequencyPenalty)
	ollamaOptions.PresencePenalty = float32(opts.PresencePenalty)

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
	f := false
	req := &api.PullRequest{
		Model:  model,
		Stream: &f,
	}

	observeProgress := o.options.pullModelObserver
	if observeProgress == nil {
		observeProgress = func(_ api.ProgressResponse) error {
			return nil
		}
	}

	err := o.client.Pull(pullCtx, req, observeProgress)
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
