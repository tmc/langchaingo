package local

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/local/internal/localclient"
	"github.com/tmc/langchaingo/util"
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

// Generate generates completions using the local LLM binary.
func (o *LLM) Generate(prompts []string, stopWords []string) ([]*llms.Generation, error) {
	_ = stopWords // TODO: use this
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

// New creates a new local LLM implementation.
func New() (*LLM, error) {
	// Require the user to set the path to the local LLM binary
	binPath := os.Getenv(localLLMBinVarName)

	// Ensure binary path is supplied
	if binPath == "" {
		return nil, ErrMissingBin
	}

	// Find the full path to the binary
	binPath, err := util.LookPath(binPath)
	if err != nil {
		return nil, err
	}

	// Allow the user to pass CLI arguments to the local LLM binary (optional)
	args := os.Getenv(localLLMArgsVarName)

	var parts []string
	var currentPart string
	var inDoubleQuote, inSingleQuote bool

	for _, char := range args {
		switch {
		case char == ' ' && !inDoubleQuote && !inSingleQuote:
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
		case char == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote
		case char == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote
		default:
			currentPart += string(char)
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}

	// Return the client
	c, err := localclient.New(binPath, parts)
	return &LLM{
		client: c,
	}, err
}
