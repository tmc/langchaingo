// Package kronk is a thin abstraction over the ardanlabs/kronk SDK. It
// encapsulates the full lifecycle — installing llama.cpp libraries, downloading
// a GGUF model, initializing the backend, and exposing a simple chat API — so
// every kronk example can share the same surface without duplicating setup
// boilerplate.
//
// Client implements the [llms.Model] interface so it can be used anywhere a
// langchaingo model is expected, including [llms.GenerateFromSinglePrompt].
// It also implements [embeddings.EmbedderClient] so it can be used with
// [embeddings.NewEmbedder] for vector store and similarity search examples.
package kronk

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	krn *kronksdk.Kronk
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

	client := Client{krn: krn}

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

// GenerateContent implements the [llms.Model] interface. It converts
// langchaingo [llms.MessageContent] messages into kronk model documents,
// calls the kronk backend, and returns a [llms.ContentResponse].
//
// When [llms.WithStreamingFunc] is provided in options, the response is
// streamed token-by-token to the streaming function.
func (c *Client) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	docs, err := messageContentToDocs(messages)
	if err != nil {
		return nil, err
	}

	d := c.applyOptions(model.D{
		"messages": docs,
	}, opts)

	// Streaming path: forward each chunk to the caller's StreamingFunc.
	if opts.StreamingFunc != nil {
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
	ch, err := c.krn.ChatStreaming(ctx, d)
	if err != nil {
		return nil, fmt.Errorf("kronk: generate content stream: %w", err)
	}

	var content strings.Builder
	for resp := range ch {
		if len(resp.Choices) == 0 {
			continue
		}

		choice := resp.Choices[0]

		if choice.FinishReason() == model.FinishReasonError {
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
			if err := opts.StreamingFunc(ctx, []byte(choice.Delta.Content)); err != nil {
				return &llms.ContentResponse{
					Choices: []*llms.ContentChoice{{Content: content.String()}},
				}, err
			}
		}
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: content.String(), StopReason: model.FinishReasonStop},
		},
	}, nil
}

// applyOptions maps relevant [llms.CallOptions] fields onto the kronk
// [model.D] request map.
func (c *Client) applyOptions(d model.D, opts llms.CallOptions) model.D {
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

	return d
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

		// Concatenate all text parts; ignore non-text parts for this
		// text-only abstraction.
		var text strings.Builder
		for _, part := range mc.Parts {
			if tc, ok := part.(llms.TextContent); ok {
				text.WriteString(tc.Text)
			}
		}
		docs = append(docs, model.TextMessage(role, text.String()))
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
	}

	if resp.Usage != nil {
		cc.GenerationInfo = map[string]any{
			"prompt_tokens":     resp.Usage.PromptTokens,
			"completion_tokens": resp.Usage.CompletionTokens,
			"total_tokens":      resp.Usage.TotalTokens,
		}
	}

	cr.Choices = []*llms.ContentChoice{cc}
	return cr
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
