package perplexity

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Model string

// Model pricing overview: https://docs.perplexity.ai/guides/pricing
const (
	ModelLlamaSonarSmall Model = "llama-3.1-sonar-small-128k-online"
	ModelLlamaSonarLarge Model = "llama-3.1-sonar-large-128k-online"
	ModelLlamaSonarHuge  Model = "llama-3.1-sonar-huge-128k-online"
)

type Option func(*options)

type options struct {
	apiKey string
	model  Model
}

func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

func WithModel(model Model) Option {
	return func(o *options) {
		o.model = model
	}
}

type Perplexity struct {
	llm *openai.LLM
}

func NewPerplexity(opts ...Option) (*Perplexity, error) {
	options := &options{
		apiKey: os.Getenv("PERPLEXITY_API_KEY"),
		model:  ModelLlamaSonarSmall, // Default model
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, fmt.Errorf("Perplexity API key not set")
	}

	llm, err := openai.New(
		openai.WithModel(string(options.model)),
		openai.WithBaseURL("https://api.perplexity.ai"),
		openai.WithToken(options.apiKey),
	)
	if err != nil {
		return nil, err
	}

	return &Perplexity{
		llm: llm,
	}, nil
}

func (p *Perplexity) Name() string {
	return "PerplexityAI"
}

func (p *Perplexity) Description() string {
	return "Perplexity AI has access to a wide range of information, as it functions as an AI-powered search engine that indexes, analyzes, and summarizes content from across the internet."
}

func (p *Perplexity) Call(ctx context.Context, input string) (string, error) {
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, input),
	}

	var generatedText string
	_, err := p.llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			generatedText += string(chunk)
			return nil
		}))
	if err != nil {
		return "", err
	}

	return generatedText, nil
}
