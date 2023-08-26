package local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/local/internal/localclient"
	"github.com/tmc/langchaingo/schema"
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

// _ ensures that LLM implements the llms.LLM and language model interface.
var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// Call calls the local LLM binary with the given prompt.
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

func (o *LLM) appendGlobalsToArgs(opts llms.CallOptions) []string {
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

	return o.client.Args
}

// Generate generates completions using the local LLM binary.
func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, prompts)
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

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		result, err := o.client.CreateCompletion(ctx, &localclient.CompletionRequest{
			Prompt: prompt,
		})
		if err != nil {
			return nil, err
		}

		generations = append(generations, &llms.Generation{Text: result.Text})
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func (o *LLM) GeneratePrompt(
	ctx context.Context,
	prompts []schema.PromptValue,
	options ...llms.CallOption,
) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, prompts, options...)
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens("gpt2", text)
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
