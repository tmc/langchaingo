// Package kronk is a thin abstraction over the ardanlabs/kronk SDK. It
// encapsulates the full lifecycle — installing llama.cpp libraries, downloading
// a GGUF model, initializing the backend, and exposing a simple chat API — so
// every kronk example can share the same surface without duplicating setup
// boilerplate.
//
// Client implements the [llms.Model] interface, including text and media
// messages, native tool calls, streaming, reasoning content, and JSON output.
// It also implements [embeddings.EmbedderClient] so it can be used with
// [embeddings.NewEmbedder] for vector store and similarity search examples.
package kronk

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"

	kronksdk "github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/ardanlabs/kronk/sdk/tools/defaults"
	"github.com/ardanlabs/kronk/sdk/tools/libs"
	"github.com/ardanlabs/kronk/sdk/tools/models"
)

// Client is the kronk abstraction. It wraps a kronk SDK instance and exposes
// a small, consistent chat API shared by all kronk examples. It also
// implements the [llms.Model] interface so it can be passed to langchaingo
// utilities like [llms.GenerateFromSinglePrompt].
type Client struct {
	krn       *kronksdk.Kronk
	callbacks callbacks.Handler
}

// New installs the required libraries and model, initializes the Kronk
// backend, loads the model, and returns a ready-to-use [Client]. ModelSource
// may be a HuggingFace URL, a canonical "provider/modelID", or a bare model ID.
func New(ctx context.Context, modelSource string, opts ...Option) (*Client, error) {
	cfg := options{
		installTimeout: 25 * time.Minute,
		logger:         kronksdk.FmtLogger,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if modelSource == "" {
		return nil, fmt.Errorf("kronk: model is required")
	}

	mp, err := installSystem(ctx, modelSource, cfg)
	if err != nil {
		return nil, fmt.Errorf("kronk: install system: %w", err)
	}

	if err := kronksdk.Init(); err != nil {
		return nil, fmt.Errorf("kronk: init backend: %w", err)
	}

	modelOpts := append([]model.Option{}, cfg.modelOptions...)
	modelOpts = append(modelOpts, model.WithModelFiles(mp.ModelFiles))

	krn, err := kronksdk.NewWithContext(ctx, modelOpts...)
	if err != nil {
		return nil, fmt.Errorf("kronk: create model: %w", err)
	}

	client := Client{
		krn:       krn,
		callbacks: cfg.callbacks,
	}

	return &client, nil
}

// Unload closes the loaded model and frees its resources. Call this only when
// you are completely done using the [Client].
func (c *Client) Unload(ctx context.Context) error {
	return c.krn.Unload(ctx)
}

// CreateEmbedding implements the [embeddings.EmbedderClient] interface. It
// embeds each of the provided texts using the loaded kronk embedding model and
// returns the vectors in the same order as the input texts.
func (c *Client) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	if _, exists := ctx.Deadline(); !exists {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
	}

	resp, err := c.krn.Embeddings(ctx, model.D{
		"input":              texts,
		"truncate":           true,
		"truncate_direction": "right",
	})
	if err != nil {
		return nil, fmt.Errorf("kronk: create embedding: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("kronk: expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	// Preserve input ordering using each response item's Index field.
	result := make([][]float32, len(texts))
	for _, data := range resp.Data {
		if data.Index < 0 || data.Index >= len(texts) {
			return nil, fmt.Errorf("kronk: embedding index %d out of range [0, %d)", data.Index, len(texts))
		}
		if len(data.Embedding) == 0 {
			return nil, fmt.Errorf("kronk: empty embedding for input index %d", data.Index)
		}
		result[data.Index] = data.Embedding
	}

	for i, vec := range result {
		if vec == nil {
			return nil, fmt.Errorf("kronk: missing embedding for input index %d", i)
		}
	}

	return result, nil
}

// -----------------------------------------------------------------------------
// llms.Model interface implementation
// -----------------------------------------------------------------------------

// Compile-time check that Client implements llms.Model.
var _ llms.Model = (*Client)(nil)

// Compile-time check that Client implements embeddings.EmbedderClient.
var _ embeddings.EmbedderClient = (*Client)(nil)

// Compile-time check that Client exposes its callback handler.
var _ callbacks.HandlerHaver = (*Client)(nil)

// GetCallbackHandler returns the configured LangChainGo callback handler.
func (c *Client) GetCallbackHandler() callbacks.Handler {
	return c.callbacks
}

// GenerateContent implements the [llms.Model] interface. It converts
// langchaingo [llms.MessageContent] messages into kronk model documents,
// calls the kronk backend, and returns a [llms.ContentResponse].
//
// When [llms.WithStreamingFunc] is provided in options, the response is
// streamed token-by-token to the streaming function.
func (c *Client) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (response *llms.ContentResponse, err error) {
	if c.callbacks != nil {
		c.callbacks.HandleLLMGenerateContentStart(ctx, messages)
		defer func() {
			if err != nil {
				c.callbacks.HandleLLMError(ctx, err)
				return
			}
			c.callbacks.HandleLLMGenerateContentEnd(ctx, response)
		}()
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	docs, err := messageContentToDocs(messages)
	if err != nil {
		return nil, err
	}

	d, err := applyOptions(model.D{
		"messages": docs,
	}, opts)
	if err != nil {
		return nil, err
	}

	// Streaming path: forward each chunk to the caller's StreamingFunc.
	if opts.StreamingFunc != nil || opts.StreamingReasoningFunc != nil {
		return c.generateContentStream(ctx, d, opts)
	}

	// Non-streaming path.
	resp, err := c.krn.Chat(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("kronk: generate content: %w", err)
	}

	return chatResponseToContentResponse(resp), nil
}

// Call implements the [llms.Model] interface. It is a convenience wrapper
// around [GenerateContent] for simple text-only prompts.
func (c *Client) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	resp, err := c.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("kronk: empty response from model")
	}
	return resp.Choices[0].Content, nil
}

func (c *Client) generateContentStream(ctx context.Context, d model.D, opts llms.CallOptions) (*llms.ContentResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	d["stream_options"] = model.D{"include_usage": true}
	ch, err := c.krn.ChatStreaming(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("kronk: generate content stream: %w", err)
	}

	return c.processContentStream(ctx, ch, opts)
}

func (c *Client) processContentStream(ctx context.Context, ch <-chan model.ChatResponse, opts llms.CallOptions) (*llms.ContentResponse, error) {
	var content strings.Builder
	var reasoning strings.Builder
	var finishReason string
	var usage *model.Usage
	toolCalls := map[int]*llms.ToolCall{}
	for resp := range ch {
		if resp.Usage != nil {
			usage = resp.Usage
		}
		if len(resp.Choices) == 0 {
			continue
		}

		choice := resp.Choices[0]
		if reason := choice.FinishReason(); reason != "" {
			finishReason = reason
		}

		if finishReason == model.FinishReasonError {
			errMsg := "unknown error"
			if choice.Delta != nil && choice.Delta.Content != "" {
				errMsg = choice.Delta.Content
			}
			return &llms.ContentResponse{
				Choices: []*llms.ContentChoice{{Content: content.String(), StopReason: model.FinishReasonError}},
			}, fmt.Errorf("kronk: model error: %s", errMsg)
		}

		if choice.Delta != nil && choice.Delta.Content != "" {
			content.WriteString(choice.Delta.Content)
			if c.callbacks != nil {
				c.callbacks.HandleStreamingFunc(ctx, []byte(choice.Delta.Content))
			}
			if opts.StreamingFunc != nil {
				if err := opts.StreamingFunc(ctx, []byte(choice.Delta.Content)); err != nil {
					return &llms.ContentResponse{
						Choices: []*llms.ContentChoice{{Content: content.String(), ReasoningContent: reasoning.String()}},
					}, err
				}
			}
		}
		if choice.Delta != nil && choice.Delta.Reasoning != "" {
			reasoning.WriteString(choice.Delta.Reasoning)
			if opts.StreamingReasoningFunc != nil {
				if err := opts.StreamingReasoningFunc(ctx, []byte(choice.Delta.Reasoning), nil); err != nil {
					return &llms.ContentResponse{
						Choices: []*llms.ContentChoice{{Content: content.String(), ReasoningContent: reasoning.String()}},
					}, err
				}
			}
		}
		if choice.Delta != nil {
			for _, delta := range choice.Delta.ToolCallDeltas {
				call := toolCalls[delta.Index]
				if call == nil {
					call = &llms.ToolCall{FunctionCall: &llms.FunctionCall{}}
					toolCalls[delta.Index] = call
				}
				if delta.ID != "" {
					call.ID = delta.ID
				}
				if delta.Type != "" {
					call.Type = delta.Type
				}
				if delta.Function.Name != "" {
					call.FunctionCall.Name = delta.Function.Name
				}
				call.FunctionCall.Arguments += delta.Function.Arguments
			}
		}
	}
	if finishReason == "" {
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{{Content: content.String(), ReasoningContent: reasoning.String()}},
		}, errors.New("kronk: response stream ended without a finish reason")
	}

	choice := &llms.ContentChoice{
		Content:          content.String(),
		ReasoningContent: reasoning.String(),
		StopReason:       finishReason,
		GenerationInfo:   usageGenerationInfo(usage),
		ToolCalls:        orderedToolCalls(toolCalls),
	}
	if len(choice.ToolCalls) > 0 {
		choice.FuncCall = choice.ToolCalls[0].FunctionCall
	}

	return &llms.ContentResponse{Choices: []*llms.ContentChoice{choice}}, nil
}

// applyOptions maps relevant [llms.CallOptions] fields onto the kronk
// [model.D] request map.
func applyOptions(d model.D, opts llms.CallOptions) (model.D, error) {
	if opts.MaxTokens != 0 {
		d["max_tokens"] = opts.MaxTokens
	}

	if opts.Temperature != 0 {
		d["temperature"] = opts.Temperature
	}
	if opts.TopK != 0 {
		d["top_k"] = opts.TopK
	}
	if opts.TopP != 0 {
		d["top_p"] = opts.TopP
	}
	if opts.Seed != 0 {
		d["seed"] = opts.Seed
	}
	if len(opts.StopWords) > 0 {
		d["stop"] = opts.StopWords
	}
	if opts.RepetitionPenalty != 0 {
		d["repeat_penalty"] = opts.RepetitionPenalty
	}
	if opts.FrequencyPenalty != 0 {
		d["frequency_penalty"] = opts.FrequencyPenalty
	}
	if opts.PresencePenalty != 0 {
		d["presence_penalty"] = opts.PresencePenalty
	}
	if opts.N != 0 {
		d["n"] = opts.N
	} else if opts.CandidateCount != 0 {
		d["n"] = opts.CandidateCount
	}
	if opts.JSONMode || opts.ResponseMIMEType == "application/json" {
		d["response_format"] = model.D{"type": "json_object"}
	} else if opts.ResponseMIMEType != "" && opts.ResponseMIMEType != "text/plain" {
		return nil, fmt.Errorf("kronk: response MIME type %q is not supported", opts.ResponseMIMEType)
	}

	tools := opts.Tools
	if len(tools) == 0 && len(opts.Functions) > 0 {
		tools = make([]llms.Tool, len(opts.Functions))
		for i := range opts.Functions {
			tools[i] = llms.Tool{Type: "function", Function: &opts.Functions[i]}
		}
	}
	if len(tools) > 0 {
		toolDocs, err := toolDocuments(tools)
		if err != nil {
			return nil, err
		}
		d["tools"] = toolDocs
	}
	if opts.ToolChoice != nil {
		choice, err := modelValue(opts.ToolChoice)
		if err != nil {
			return nil, fmt.Errorf("kronk: tool choice: %w", err)
		}
		d["tool_choice"] = choice
	} else if opts.FunctionCallBehavior != "" {
		d["tool_choice"] = string(opts.FunctionCallBehavior)
	}

	return d, nil
}

// messageContentToDocs converts langchaingo [llms.MessageContent] messages
// into kronk [model.D] documents.
func messageContentToDocs(messages []llms.MessageContent) ([]model.D, error) {
	docs := model.DocumentArray()

	for _, mc := range messages {
		role, err := chatMessageTypeToRole(mc.Role)
		if err != nil {
			return nil, err
		}

		doc := model.D{"role": role}
		var content []model.D
		var toolCalls []model.D
		for _, part := range mc.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				content = append(content, model.D{"type": "text", "text": p.Text})
			case llms.ImageURLContent:
				if strings.HasPrefix(p.URL, "http://") || strings.HasPrefix(p.URL, "https://") {
					return nil, errors.New("kronk: remote media URLs are not supported; use a data URL or BinaryPart")
				}
				content = append(content, model.D{
					"type":      "image_url",
					"image_url": model.D{"url": p.URL},
				})
			case llms.BinaryContent:
				mediaPart, err := binaryContentPart(p)
				if err != nil {
					return nil, err
				}
				content = append(content, mediaPart)
			case llms.ToolCall:
				if p.FunctionCall == nil {
					return nil, errors.New("kronk: tool call is missing its function")
				}
				toolCalls = append(toolCalls, model.D{
					"id":   p.ID,
					"type": p.Type,
					"function": model.D{
						"name":      p.FunctionCall.Name,
						"arguments": p.FunctionCall.Arguments,
					},
				})
			case llms.ToolCallResponse:
				if len(mc.Parts) != 1 {
					return nil, errors.New("kronk: tool response must be the only message part")
				}
				doc["tool_call_id"] = p.ToolCallID
				doc["name"] = p.Name
				doc["content"] = p.Content
			default:
				return nil, fmt.Errorf("kronk: content part type %T is not supported", part)
			}
		}

		if _, exists := doc["content"]; !exists && len(content) > 0 {
			if len(content) == 1 && content[0]["type"] == "text" {
				doc["content"] = content[0]["text"]
			} else {
				doc["content"] = content
			}
		}
		if len(toolCalls) > 0 {
			doc["tool_calls"] = toolCalls
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func chatMessageTypeToRole(mt llms.ChatMessageType) (string, error) {
	switch mt {
	case llms.ChatMessageTypeSystem:
		return model.RoleSystem, nil
	case llms.ChatMessageTypeHuman, llms.ChatMessageTypeGeneric:
		return model.RoleUser, nil
	case llms.ChatMessageTypeAI:
		return model.RoleAssistant, nil
	case llms.ChatMessageTypeTool, llms.ChatMessageTypeFunction:
		return model.RoleTool, nil
	default:
		return "", fmt.Errorf("kronk: unsupported chat message type %q", mt)
	}
}

func chatResponseToContentResponse(resp model.ChatResponse) *llms.ContentResponse {
	cr := &llms.ContentResponse{}

	if len(resp.Choices) == 0 {
		return cr
	}

	choice := resp.Choices[0]
	cc := &llms.ContentChoice{
		StopReason: choice.FinishReason(),
	}
	if choice.Message != nil {
		cc.Content = choice.Message.Content
		cc.ReasoningContent = choice.Message.Reasoning
		cc.ToolCalls = toolCallsToLangChain(choice.Message.ToolCalls)
		if len(cc.ToolCalls) > 0 {
			cc.FuncCall = cc.ToolCalls[0].FunctionCall
		}
	}

	cc.GenerationInfo = usageGenerationInfo(resp.Usage)

	cr.Choices = []*llms.ContentChoice{cc}
	return cr
}

func binaryContentPart(content llms.BinaryContent) (model.D, error) {
	mediaType, _, ok := strings.Cut(content.MIMEType, "/")
	if !ok {
		return nil, fmt.Errorf("kronk: invalid media MIME type %q", content.MIMEType)
	}

	data := "data:" + content.MIMEType + ";base64," + base64.StdEncoding.EncodeToString(content.Data)
	switch mediaType {
	case "image":
		return model.D{"type": "image_url", "image_url": model.D{"url": data}}, nil
	case "video":
		return model.D{"type": "video_url", "video_url": model.D{"url": data}}, nil
	case "audio":
		return model.D{"type": "input_audio", "input_audio": model.D{"data": data}}, nil
	default:
		return nil, fmt.Errorf("kronk: unsupported media MIME type %q", content.MIMEType)
	}
}

func toolDocuments(tools []llms.Tool) ([]model.D, error) {
	docs := make([]model.D, len(tools))
	for i, tool := range tools {
		if tool.Function == nil {
			return nil, fmt.Errorf("kronk: tool %d is missing its function", i)
		}

		value, err := modelValue(tool.Function.Parameters)
		if err != nil {
			return nil, fmt.Errorf("kronk: tool %q parameters: %w", tool.Function.Name, err)
		}

		toolType := tool.Type
		if toolType == "" {
			toolType = "function"
		}
		function := model.D{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
			"parameters":  value,
		}
		if tool.Function.Strict {
			function["strict"] = true
		}
		docs[i] = model.D{
			"type":     toolType,
			"function": function,
		}
	}

	return docs, nil
}

func modelValue(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, err
	}

	return modelValueFromJSON(decoded), nil
}

func modelValueFromJSON(value any) any {
	switch value := value.(type) {
	case map[string]any:
		d := make(model.D, len(value))
		for key, item := range value {
			d[key] = modelValueFromJSON(item)
		}
		return d
	case []any:
		items := make([]any, len(value))
		for i, item := range value {
			items[i] = modelValueFromJSON(item)
		}
		return items
	default:
		return value
	}
}

func toolCallsToLangChain(toolCalls []model.ResponseToolCall) []llms.ToolCall {
	result := make([]llms.ToolCall, len(toolCalls))
	for i, toolCall := range toolCalls {
		arguments, _ := json.Marshal(map[string]any(toolCall.Function.Arguments))
		result[i] = llms.ToolCall{
			ID:   toolCall.ID,
			Type: toolCall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolCall.Function.Name,
				Arguments: string(arguments),
			},
		}
	}
	return result
}

func orderedToolCalls(toolCalls map[int]*llms.ToolCall) []llms.ToolCall {
	result := make([]llms.ToolCall, 0, len(toolCalls))
	for _, index := range slices.Sorted(maps.Keys(toolCalls)) {
		result = append(result, *toolCalls[index])
	}
	return result
}

func usageGenerationInfo(usage *model.Usage) map[string]any {
	if usage == nil {
		return nil
	}

	return map[string]any{
		"prompt_tokens":          usage.PromptTokens,
		"completion_tokens":      usage.CompletionTokens,
		"total_tokens":           usage.TotalTokens,
		"reasoning_tokens":       usage.CompletionTokensDetails.ReasoningTokens,
		"cached_tokens":          usage.PromptTokensDetails.CachedTokens,
		"tokens_per_second":      usage.TokensPerSecond,
		"time_to_first_token_ms": usage.TimeToFirstTokenMS,
	}
}

// -----------------------------------------------------------------------------

// installSystem downloads and installs the llama.cpp libraries and the
// configured model, returning the on-disk model path.
func installSystem(ctx context.Context, modelSource string, cfg options) (models.Path, error) {
	ctx, cancel := context.WithTimeout(ctx, cfg.installTimeout)
	defer cancel()

	libVersion := cfg.libVersion
	if libVersion == "" {
		libVersion = defaults.LibVersion("")
	}

	lb, err := libs.New(libs.WithVersion(libVersion))
	if err != nil {
		return models.Path{}, err
	}

	if _, err := lb.Download(ctx, cfg.logger); err != nil {
		return models.Path{}, fmt.Errorf("unable to install llama.cpp: %w", err)
	}

	mdls, err := models.New()
	if err != nil {
		return models.Path{}, fmt.Errorf("unable to init models: %w", err)
	}

	mp, err := mdls.Download(ctx, cfg.logger, modelSource)
	if err != nil {
		return models.Path{}, fmt.Errorf("unable to install model: %w", err)
	}

	return mp, nil
}
