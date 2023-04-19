package local

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/local/internal/localclient"
)

const localLLMBinVarName = "LOCAL_LLM_BIN"
const localLLMArgsVarName = "LOCAL_LLM_ARGS"

var ErrEmptyResponse = errors.New("no response")
var ErrMissingBin = errors.New("missing the local LLM binary path, set the LOCAL_LLM_BIN environment variable")

type LLM struct {
	client *localclient.Client
}

var _ llms.LLM = (*LLM)(nil)

func (o *LLM) Call(prompt string) (string, error) {
	r, err := o.Generate([]string{prompt})
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(prompts []string) ([]*llms.Generation, error) {
	result, err := o.client.CreateCompletion(context.TODO(), &localclient.CompletionRequest{
		Prompt: prompts[0],
	})
	if err != nil {
		return nil, err
	}
	return []*llms.Generation{
		{Text: result.Text},
	}, nil
}

func New() (*LLM, error) {
	// Require the user to set the path to the local LLM binary
	binPath := os.Getenv(localLLMBinVarName)
	if binPath == "" {
		return nil, ErrMissingBin
	}

	// Allow the user to pass CLI arguments to the local LLM binary (optional)
	args := os.Getenv(localLLMArgsVarName)

	c, err := localclient.New(binPath, args)
	return &LLM{
		client: c,
	}, err
}
