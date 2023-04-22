package openai

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

const tokenEnvVarName = "OPENAI_API_KEY" //nolint:gosec

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	client *openaiclient.Client
}

var _ llms.LLM = (*LLM)(nil)

func (o *LLM) Call(prompt string, stopWords []string) (string, error) {
	r, err := o.Generate([]string{prompt}, stopWords)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(prompts []string, stopWords []string) ([]*llms.Generation, error) {
	result, err := o.client.CreateCompletion(context.TODO(), &openaiclient.CompletionRequest{
		Prompt:    prompts[0],
		StopWords: stopWords,
	})
	if err != nil {
		return nil, err
	}
	return []*llms.Generation{
		{Text: result.Text},
	}, nil
}

// CreateEmbedding creates an embedding for the given input text.
func (o *LLM) CreateEmbedding(inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(context.TODO(), &openaiclient.EmbeddingRequest{
		Input: inputTexts,
	})

	if len(embeddings) == 0 {
		return [][]float64{}, ErrEmptyResponse
	}

	if err != nil {
		return [][]float64{}, err
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}

	return embeddings, nil
}

func New() (*LLM, error) {
	token := os.Getenv(tokenEnvVarName)
	if token == "" {
		return nil, ErrMissingToken
	}
	c, err := openaiclient.New(token)
	return &LLM{
		client: c,
	}, err
}
