package cloudflare

import (
	"context"
	"errors"
	"net/http"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cloudflare/internal/cloudflareclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
)

// LLM is a cloudflare LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *cloudflareclient.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new cloudflare LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(&o)
	}

	client := cloudflareclient.NewClient(
		o.httpClient,
		o.cloudflareAccountID,
		o.cloudflareServerURL.String(),
		o.cloudflareToken,
		o.model,
		o.embeddingModel,
	)

	return &LLM{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen, goerr113
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Our input is a sequence of Message, each of which potentially has
	// a sequence of Part that is text.
	// We have to convert it to a format Cloudflare understands: []Message, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := []cloudflareclient.Message{}

	if o.options.system != "" {
		chatMsgs = append(chatMsgs, cloudflareclient.Message{
			Role:    cloudflareclient.RoleSystem,
			Content: o.options.system,
		})
	}

	for i := range messages {
		mc := messages[i]

		msg := cloudflareclient.Message{
			Role: typeToRole(mc.Role),
		}

		// Look at all the parts in mc; expect to find a single Text part and
		// any number of binary parts.
		var text string
		var foundText bool

		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				if foundText {
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text
			case llms.BinaryContent:
				return nil, errors.New("only supports Text right now")
			default:
				return nil, errors.New("only supports Text right now")
			}
		}

		msg.Content = text
		chatMsgs = append(chatMsgs, msg)
	}

	stream := func(b bool) *bool { return &b }(opts.StreamingFunc != nil)

	res, err := o.client.GenerateContent(ctx, &cloudflareclient.GenerateContentRequest{
		Messages:      chatMsgs,
		Stream:        *stream,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		return nil, err
	}

	for i := range res.Errors {
		return nil, errors.Join(errors.New(res.Errors[i].Message))
	}

	choices := []*llms.ContentChoice{
		{
			Content: res.Result.Response,
		},
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	res, err := o.client.CreateEmbedding(ctx, &cloudflareclient.CreateEmbeddingRequest{
		Text: inputTexts,
	})
	if err != nil {
		return nil, err
	}

	if len(res.Result.Data) == 0 {
		return nil, ErrEmptyResponse
	}

	if len(inputTexts) != len(res.Result.Data) {
		return res.Result.Data, llms.ErrIncompleteEmbedding
	}

	return res.Result.Data, nil
}

func typeToRole(typ llms.ChatMessageType) cloudflareclient.Role {
	switch typ {
	case llms.ChatMessageTypeSystem:
		return cloudflareclient.RoleSystem
	case llms.ChatMessageTypeAI:
		return cloudflareclient.RoleAssistant
	case llms.ChatMessageTypeHuman:
		fallthrough
	case llms.ChatMessageTypeGeneric:
		return cloudflareclient.RoleTypeUser
	case llms.ChatMessageTypeFunction:
		return "function"
	case llms.ChatMessageTypeTool:
		return "tool"
	}
	return ""
}
