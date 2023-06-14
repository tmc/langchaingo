package vertexai

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexaiclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingProjectID         = errors.New("missing the GCP Project ID, set it in the GOOGLE_CLOUD_PROJECT environment variable") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrNotImplemented           = errors.New("not implemented")
)

type LLM struct {
	client *vertexaiclient.PaLMClient
}

var _ llms.LLM = (*LLM)(nil)

// New returns a new VertexAI PaLM LLM.
func New(opts ...Option) (*LLM, error) {
	// Ensure options are initialized only once.
	initOptions.Do(initOpts)

	options := &options{}
	*options = *defaultOptions // Copy default options.

	for _, opt := range opts {
		opt(options)
	}

	if len(options.projectID) == 0 {
		return nil, ErrMissingProjectID
	}

	client, err := vertexaiclient.New(options.projectID, options.clientOptions...)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client: client,
	}, nil
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	results, err := o.client.CreateCompletion(ctx, &vertexaiclient.CompletionRequest{
		Prompts:     prompts,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	})
	if err != nil {
		return nil, err
	}

	generations := []*llms.Generation{}
	for _, r := range results {
		generations = append(generations, &llms.Generation{
			Text: r.Text,
		})
	}
	return generations, nil
}

type ChatMessage = vertexaiclient.ChatMessage

// Chat requests a chat response for the given prompt.
func (o *LLM) Chat(ctx context.Context, messages []schema.ChatMessage, options ...llms.CallOption) (*llms.ChatGeneration, error) { // nolint: lll
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	if opts.StreamingFunc != nil {
		return nil, ErrNotImplemented
	}
	msgs := make([]*vertexaiclient.ChatMessage, len(messages))
	for i, m := range messages {
		msg := &vertexaiclient.ChatMessage{
			Content: m.GetText(),
		}
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			msg.Author = "bot"
		case schema.ChatMessageTypeAI:
			msg.Author = "bot"
		case schema.ChatMessageTypeHuman:
			msg.Author = "user"
		case schema.ChatMessageTypeGeneric:
			msg.Author = "user"
		}
		msgs[i] = msg
	}

	result, err := o.client.CreateChat(ctx, &vertexaiclient.ChatRequest{
		Temperature: opts.Temperature,
		Messages:    msgs,
	})
	if err != nil {
		return nil, err
	}
	if len(result.Candidates) == 0 {
		return nil, ErrEmptyResponse
	}
	return &llms.ChatGeneration{
		Message: &schema.AIChatMessage{
			Text: result.Candidates[0].Content,
		},
	}, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &vertexaiclient.EmbeddingRequest{
		Input: inputTexts,
	})
	if err != nil {
		return [][]float64{}, err
	}

	if len(embeddings) == 0 {
		return [][]float64{}, ErrEmptyResponse
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}
