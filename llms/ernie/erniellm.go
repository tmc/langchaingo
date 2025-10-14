package ernie

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/ernie/internal/ernieclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrCodeResponse  = errors.New("has error code")
)

type LLM struct {
	client           *ernieclient.Client
	model            ModelName
	CallbacksHandler callbacks.Handler
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	options := &options{
		apiKey:    os.Getenv(ernieAPIKey),
		secretKey: os.Getenv(ernieSecretKey),
	}

	for _, opt := range opts {
		opt(options)
	}

	c, err := newClient(options)

	return &LLM{
		client:           c,
		model:            options.modelName,
		CallbacksHandler: options.callbacksHandler,
	}, err
}

func newClient(opts *options) (*ernieclient.Client, error) {
	if opts.accessToken == "" && (opts.apiKey == "" || opts.secretKey == "") {
		return nil, fmt.Errorf(`%w
You can pass auth info by use ernie.New(ernie.WithAKSK("{api Key}","{serect Key}")) ,
or
export ERNIE_API_KEY={API Key} 
export ERNIE_SECRET_KEY={Secret Key}
doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/flfmc9do2`, ernieclient.ErrNotSetAuth)
	}

	clientOpts := []ernieclient.Option{
		ernieclient.WithAccessToken(opts.accessToken),
		ernieclient.WithAKSK(opts.apiKey, opts.secretKey),
	}

	if opts.httpClient != nil {
		clientOpts = append(clientOpts, ernieclient.WithHTTPClient(opts.httpClient))
	}

	return ernieclient.New(clientOpts...)
}

func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
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

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	result, err := o.client.CreateCompletion(ctx, o.getModelPath(*opts), &ernieclient.CompletionRequest{
		Messages:      []ernieclient.Message{{Role: "user", Content: part.(llms.TextContent).Text}},
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		PenaltyScore:  opts.RepetitionPenalty,
		StreamingFunc: opts.StreamingFunc,
		Stream:        opts.StreamingFunc != nil,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}
	if result.ErrorCode > 0 {
		err = fmt.Errorf("%w, error_code:%v, erro_msg:%v, id:%v",
			ErrCodeResponse, result.ErrorCode, result.ErrorMsg, result.ID)
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Result,
			},
		},
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

// CreateEmbedding use ernie Embedding-V1.
// 1. texts counts less than 16
// 2. text runes counts less than 384
// doc: https://cloud.baidu.com/doc/WENXINWORKSHOP/s/alj562vvu
func (o *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	resp, e := o.client.CreateEmbedding(ctx, texts)
	if e != nil {
		return nil, e
	}

	if resp.ErrorCode > 0 {
		return nil, fmt.Errorf("%w, error_code:%v, erro_msg:%v, id:%v",
			ErrCodeResponse, resp.ErrorCode, resp.ErrorMsg, resp.ID)
	}

	emb := make([][]float32, 0, len(texts))
	for i := range resp.Data {
		emb = append(emb, resp.Data[i].Embedding)
	}

	return emb, nil
}

func (o *LLM) getModelPath(opts llms.CallOptions) ernieclient.ModelPath {
	model := o.model

	if model == "" {
		model = ModelName(opts.Model)
	}

	return modelToPath(model)
}

func modelToPath(model ModelName) ernieclient.ModelPath {
	switch model {
	case ModelNameERNIEBot:
		return "completions"
	case ModelNameERNIEBotTurbo:
		return "eb-instant"
	case ModelNameERNIEBotPro:
		return "completions_pro"
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
