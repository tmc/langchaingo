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

type Perplexity struct {
	llm *openai.LLM
}

func NewPerplexity(model Model) (*Perplexity, error) {
	perplexity := &Perplexity{}
	var err error

	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY not set")
	}

	perplexity.llm, err = openai.New(
		openai.WithModel(string(model)),
		openai.WithBaseURL("https://api.perplexity.ai"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		return nil, err
	}

	return perplexity, nil
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
