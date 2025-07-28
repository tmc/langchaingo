package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/0xDezzy/langchaingo/callbacks"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/local/internal/localclient"
)

var (
	// ErrEmptyResponse is returned when the local LLM binary returns an empty response.
	ErrEmptyResponse = errors.New("no response")
	// ErrMissingBin is returned when the LOCAL_LLM_BIN environment variable is not set.
	ErrMissingBin = errors.New("missing the local LLM binary path, set the LOCAL_LLM_BIN environment variable")
)

// LLM is a local LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *localclient.Client
}

var _ llms.Model = (*LLM)(nil)

// Call calls the local LLM binary with the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

func (o *LLM) appendGlobalsToArgs(opts llms.CallOptions) {
	if opts.Temperature != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--temperature=%f", opts.Temperature))
	}
	if opts.TopP != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--top_p=%f", opts.TopP))
	}
	if opts.TopK != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--top_k=%d", opts.TopK))
	}
	if opts.MinLength != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--min_length=%d", opts.MinLength))
	}
	if opts.MaxLength != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--max_length=%d", opts.MaxLength))
	}
	if opts.RepetitionPenalty != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--repetition_penalty=%f", opts.RepetitionPenalty))
	}
	if opts.Seed != 0 {
		o.client.Args = append(o.client.Args, fmt.Sprintf("--seed=%d", opts.Seed))
	}
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	// If o.client.GlobalAsArgs is true
	if o.client.GlobalAsArgs {
		// Then add the option to the args in --key=value format
		o.appendGlobalsToArgs(*opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	result, err := o.client.CreateCompletion(ctx, &localclient.CompletionRequest{
		Prompt: part.(llms.TextContent).Text,
	})
	if err != nil {
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

// New creates a new local LLM implementation.
func New(opts ...Option) (*LLM, error) {
	options := &options{
		bin:  os.Getenv(localLLMBinVarName),
		args: os.Getenv(localLLMArgsVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	path, err := exec.LookPath(options.bin)
	if err != nil {
		return nil, errors.Join(ErrMissingBin, err)
	}

	c, err := localclient.New(path, options.globalAsArgs, strings.Split(options.args, " ")...)
	return &LLM{
		client: c,
	}, err
}
