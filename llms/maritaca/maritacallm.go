package maritaca

import (
	"context"
	"errors"
	"net/http"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/maritaca/internal/maritacaclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("no all input got emmbedded")
)

// LLM is a maritaca LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *maritacaclient.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new maritaca LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	if o.httpClient == nil {
		o.httpClient = http.DefaultClient
	}

	client, err := maritacaclient.NewClient(o.httpClient)
	if err != nil {
		return nil, err
	}

	return &LLM{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
// nolint: goerr113
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

	// Our input is a sequence of MessageContent, each of which potentially has
	// a sequence of Part that could be text, images etc.
	// We have to convert it to a format maritaca undestands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := make([]*maritacaclient.Message, 0, len(messages))
	for _, mc := range messages {
		msg := &maritacaclient.Message{Role: typeToRole(mc.Role)}

		// Look at all the parts in mc; expect to find a single Text part and
		// any number of binary parts.
		var text string
		foundText := false

		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				if foundText {
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text

			default:
				return nil, errors.New("only support Text and BinaryContent parts right now")
			}
		}

		msg.Content = text

		chatMsgs = append(chatMsgs, msg)
	}

	format := o.options.format
	if opts.JSONMode {
		format = "json"
	}

	// Get our maritacaOptions from llms.CallOptions
	maritacaOptions := makemaritacaOptionsFromOptions(o.options.maritacaOptions, opts)
	req := &maritacaclient.ChatRequest{
		Model:    model,
		Format:   format,
		Messages: chatMsgs,
		Options:  maritacaOptions,
		Stream:   func(b bool) *bool { return &b }(opts.StreamingFunc != nil),
	}

	var fn maritacaclient.ChatResponseFunc
	streamedResponse := ""
	var resp maritacaclient.ChatResponse

	fn = func(response maritacaclient.ChatResponse) error {
		if opts.StreamingFunc != nil && response.Text != "" {
			if err := opts.StreamingFunc(ctx, []byte(response.Text)); err != nil {
				return err
			}
		}
		if response.Text != "" {
			streamedResponse += response.Text
		}
		if response.Event == "end" {
			resp.Answer = streamedResponse
		}
		if response.Model != "" && response.Text == "" {
			resp = response
		}
		return nil
	}
	o.client.Token = o.options.maritacaOptions.Token
	err := o.client.Generate(ctx, req, fn)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	choices := []*llms.ContentChoice{
		{
			Content: resp.Answer,
			GenerationInfo: map[string]any{
				"CompletionTokens": resp.Metrics.Usage.CompletionTokens,
				"PromptTokens":     resp.Metrics.Usage.PromptTokens,
				"TotalTokens":      resp.Metrics.Usage.TotalTokens,
			},
		},
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func typeToRole(typ schema.ChatMessageType) string {
	switch typ {
	case schema.ChatMessageTypeSystem:
		return "system"
	case schema.ChatMessageTypeAI:
		return "assistant"
	case schema.ChatMessageTypeHuman:
		fallthrough
	case schema.ChatMessageTypeGeneric:
		return "user"
	case schema.ChatMessageTypeFunction:
		return "function"
	case schema.ChatMessageTypeTool:
		return "tool"
	}
	return ""
}

func makemaritacaOptionsFromOptions(maritacaOptions maritacaclient.Options, opts llms.CallOptions) maritacaclient.Options {
	// Load back CallOptions as maritacaOptions
	maritacaOptions.MaxTokens = opts.MaxTokens
	maritacaOptions.Model = opts.Model
	maritacaOptions.TopP = opts.TopP
	maritacaOptions.RepetitionPenalty = opts.RepetitionPenalty
	maritacaOptions.StoppingTokens = opts.StopWords
	maritacaOptions.Stream = opts.StreamingFunc != nil

	return maritacaOptions
}
