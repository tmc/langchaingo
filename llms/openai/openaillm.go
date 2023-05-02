package openai

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/logger"
)

const (
	tokenEnvVarName = "OPENAI_API_KEY" //nolint:gosec
	modelEnvVarName = "OPENAI_MODEL"   //nolint:gosec
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	client *openaiclient.Client
}

var _ llms.LLM = (*LLM)(nil)

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, stopWords []string) (string, error) {
	// Generate a completion
	r, err := o.Generate(ctx, []string{prompt}, stopWords)
	if err != nil {
		return "", err
	}

	// Validate
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}

	// Return the first completion
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, stopWords []string) ([]*llms.Generation, error) {
	logger.LLMRequest(prompts[0])
	result, err := o.client.CreateCompletion(ctx, &openaiclient.CompletionRequest{
		Prompt:    prompts[0],
		StopWords: stopWords,
	})
	if err != nil {
		logger.LLMResponse(err.Error())
		return nil, err
	}

	logger.LLMResponse(result.Text)
	return []*llms.Generation{
		{Text: result.Text},
	}, nil
}

// Chat requests a chat response for the given prompt.
func (o *LLM) Chat(prompt string) (string, error) {
	logger.LLMRequest(prompt)
	result, err := o.client.CreateChat(context.TODO(), &openaiclient.ChatRequest{Prompt: prompt})
	if err != nil {
		logger.LLMResponse(err.Error())
		return "", err
	}

	logger.LLMResponse(result.Text)
	return result.Text, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
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

// New returns a new OpenAI client.
func New() (*LLM, error) {
	// Require the OpenAI API key to be set.
	token := os.Getenv(tokenEnvVarName)
	if token == "" {
		return nil, ErrMissingToken
	}

	// Allow model selection.
	model := os.Getenv(modelEnvVarName)

	// Create the client.
	c, err := openaiclient.New(token, model)
	return &LLM{
		client: c,
	}, err
}
