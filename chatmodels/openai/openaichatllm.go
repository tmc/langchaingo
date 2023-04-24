package openai

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/chatmodels/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/llms"
)

const tokenEnvVarName = "OPENAI_API_KEY"

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")
)

type LLM struct {
	client *openaiclient.Client
}

var _ llms.LLM = (*LLM)(nil)

// Call calls the OpenAI API
// The prompt parameter is a schema.ChatMessage, which is a message from a user.
// The stopWords parameter is a list of words that should not be included in the response.
// The response is a string, which is the response from the OpenAI API.
// func (o *LLM) Call(prompt schema.ChatMessage, stopWords []string) (string, error) {
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

// Generate calls the OpenAI API
func (o *LLM) Generate(prompts []string, stopWords []string) ([]*llms.Generation, error) {
	result, err := o.client.Chat(context.TODO(), &openaiclient.ChatRequest{
		Messages: prompts,
		Stop:     stopWords,
	})
	if err != nil {
		return nil, err
	}

	return []*llms.Generation{
		{
			Text:           result.Text,
			GenerationInfo: map[string]any{},
		},
	}, nil
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
