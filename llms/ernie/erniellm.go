package ernie

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie/internal/ernieclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrCodeResponse  = errors.New("has error code")
)

type LLM struct {
	client *ernieclient.Client
	model  ModelName
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)

	return &LLM{
		client: c,
	}, err
}

func newClient(opts ...Option) (*ernieclient.Client, error) {
	options := &options{
		apiKey:    os.Getenv(ernieAPIKey),
		secretKey: os.Getenv(ernieSecretKey),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.accessToken == "" && (options.apiKey == "" || options.secretKey == "") {
		return nil, fmt.Errorf(`%w
You can pass auth info by use ernie.New(ernie.WithAKSK("{api Key}","{serect Key}")) ,
or
export ERNIE_API_KEY={API Key} 
export ERNIE_SECRET_KEY={Secret Key}
doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2`, ernieclient.ErrNotSetAuth)
	}

	return ernieclient.New(
		ernieclient.WithAccessToken(options.accessToken),
		ernieclient.WithAKSK(options.apiKey, options.secretKey))
}

// GeneratePrompt implements llms.LanguageModel.
func (l *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue,
	options ...llms.CallOption,
) (llms.LLMResult, error) {
	return llms.GeneratePrompt(ctx, l, promptValues, options...)
}

// GetNumTokens implements llms.LanguageModel.
func (l *LLM) GetNumTokens(_ string) int {
	// todo: not provided yet
	// see: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Nlks5zkzu
	return -1
}

// Call implements llms.LLM.
func (l *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := l.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}

	if len(r) == 0 {
		return "", ErrEmptyResponse
	}

	return r[0].Text, nil
}

// Generate implements llms.LLM.
func (l *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		result, err := l.client.CreateCompletion(ctx, l.getModelPath(opts), &ernieclient.CompletionRequest{
			Messages:      []ernieclient.Message{{Role: "user", Content: prompt}},
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			PenaltyScore:  opts.RepetitionPenalty,
			StreamingFunc: opts.StreamingFunc,
			Stream:        opts.StreamingFunc != nil,
		})
		if err != nil {
			return nil, err
		}
		generations = append(generations, &llms.Generation{
			Text: result.Result,
		})
	}

	return generations, nil
}

// CreateEmbedding use ernie Embedding-V1.
// 1. texts counts less than 16
// 2. text runes counts less than 384
// doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
func (l *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float64, error) {
	resp, e := l.client.CreateEmbedding(ctx, texts)
	if e != nil {
		return nil, e
	}

	if resp.ErrorCode > 0 {
		return nil, fmt.Errorf("%w, error_code:%v, erro_msg:%v",
			ErrCodeResponse, resp.ErrorCode, resp.ErrorMsg)
	}

	emb := make([][]float64, 0, len(texts))
	for i := range resp.Data {
		emb = append(emb, resp.Data[i].Embedding)
	}

	return emb, nil
}

func (l *LLM) getModelPath(opts llms.CallOptions) ernieclient.ModelPath {
	model := l.model

	if model == "" {
		model = ModelName(opts.Model)
	}

	switch model {
	case ModelNameERNIEBot:
		return "completions"
	case ModelNameERNIEBotTurbo:
		return "eb-instant"
	case ModelNameBloomz7B:
		return "bloomz_7b1"
	case ModelNameLlama2_7BChat:
		return "llama_2_7b"
	case ModelNameLlama2_13BChat:
		return "llama_2_13b"
	case ModelNameLlama2_70BChat:
		return "llama_2_70b"
	default:
		return ernieclient.DefaultCompletionModelPath
	}
}
