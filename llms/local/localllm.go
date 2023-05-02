package local

import (
	"context"
	"errors"
	"os"
	"os/exec"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/local/internal/localclient"
	"github.com/tmc/langchaingo/logger"
)

const (
	// The name of the environment variable that contains the path to the local LLM binary.
	localLLMBinVarName = "LOCAL_LLM_BIN"
	// The name of the environment variable that contains the CLI arguments to pass to the local LLM binary.
	localLLMArgsVarName = "LOCAL_LLM_ARGS"
)

var (
	// ErrEmptyResponse is returned when the local LLM binary returns an empty response.
	ErrEmptyResponse = errors.New("no response")
	// ErrMissingBin is returned when the LOCAL_LLM_BIN environment variable is not set.
	ErrMissingBin = errors.New("missing the local LLM binary path, set the LOCAL_LLM_BIN environment variable")
)

// LLM is a local LLM implementation.
type LLM struct {
	client *localclient.Client
}

// _ ensures that LLM implements the llms.LLM interface.
var _ llms.LLM = (*LLM)(nil)

// Call calls the local LLM binary with the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, stopWords []string) (string, error) {
	// Generate a completion
	logger.LLMRequest(prompt)
	r, err := o.Generate(ctx, []string{prompt}, stopWords)
	if err != nil {
		return "", err
	}

	// Validate
	if len(r) == 0 {
		logger.LLMResponse(ErrEmptyResponse.Error())
		return "", ErrEmptyResponse
	}

	// Return the first completion
	logger.LLMResponse(r[0].Text)
	return r[0].Text, nil
}

// Generate generates completions using the local LLM binary.
func (o *LLM) Generate(ctx context.Context, prompts []string, stopWords []string) ([]*llms.Generation, error) {
	logger.LLMRequest(prompts[0])
	_ = stopWords // TODO: use this
	result, err := o.client.CreateCompletion(ctx, &localclient.CompletionRequest{
		Prompt: prompts[0],
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

// New creates a new local LLM implementation.
func New() (*LLM, error) {
	// Require the user to set the path to the local LLM binary
	bin := os.Getenv(localLLMBinVarName)
	path, err := exec.LookPath(bin)
	if err != nil {
		return nil, errors.Join(ErrMissingBin, err)
	}

	// Allow the user to pass CLI arguments to the local LLM binary (optional)
	args := os.Getenv(localLLMArgsVarName)
	c, err := localclient.New(path, args)
	return &LLM{
		client: c,
	}, err
}
