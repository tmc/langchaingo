package qwen_text

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/qwen"
	"github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
)

var (
	ErrCodeResponse = errors.New("has error code")
)

type LLM struct {
	client           *qwenclient.Client
	model            qwen.ModelName
	CallbacksHandler callbacks.Handler
}

var _ llms.Model = (*LLM)(nil)

func (o *LLM) Call(ctx context.Context, prompt string, Options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, Options...)
}

func New(opts ...qwen.Option) (*LLM, error) {
	Options := &qwen.Options{
		ApiKey: os.Getenv(qwen.QwenAPIKey),
	}

	for _, opt := range opts {
		opt(Options)
	}

	c, err := newClient(Options)

	return &LLM{
		client:           c,
		model:            Options.ModelName,
		CallbacksHandler: Options.CallbacksHandler,
	}, err
}

func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, Options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range Options {
		opt(opts)
	}

	msg0 := messages[0]
	part := msg0.Parts[0]
	input := struct {
		Messages []qwenclient.Message `json:"messages"`
	}{
		Messages: []qwenclient.Message{{Role: "user", Content: part.(llms.TextContent).Text}},
	}

	parameters := qwenclient.Parameters{
		Temperature:  opts.Temperature,
		ResultFormat: "message",
	}

	if opts.StreamingFunc != nil {
		parameters.IncrementalOutput = true
	}

	result, err := o.client.CreateCompletion(ctx, &qwenclient.CompletionRequest{
		Model:         string(o.getModelPath(*opts)),
		Input:         input,
		Parameters:    parameters,
		Stream:        opts.StreamingFunc != nil,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}
	if result.Code != "" {
		err = fmt.Errorf("%w, error_code:%v, erro_msg:%v, id:%v",
			ErrCodeResponse, result.Code, result.Message, result.RequestID)
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	content := result.Output.Choices[0].Message.Content
	if opts.StreamingFunc != nil {
		content = result.Output.Text
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: content,
			},
		},
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, resp)
	}

	return resp, nil
}

func newClient(opts *qwen.Options) (*qwenclient.Client, error) {
	if opts.ApiKey == "" {
		return nil, fmt.Errorf(`%w
You can pass auth info by use qwen_text.New(qwen.WithAK("{api Key}")) ,
or
export Qwen_API_KEY={API Key} `, qwenclient.ErrNotSetAuth)
	}

	return qwenclient.New(qwenclient.WithAK(opts.ApiKey))
}

func (o *LLM) getModelPath(opts llms.CallOptions) qwenclient.ModelPath {
	model := o.model

	if model == "" {
		model = qwen.ModelName(opts.Model)
	}

	return modelToPath(model)
}

func modelToPath(model qwen.ModelName) qwenclient.ModelPath {
	switch model {
	case qwen.ModelNameQwen_Turbo:
		return "qwen-turbo"
	case qwen.ModelNameQwen_Plus:
		return "qwen-plus"
	case qwen.ModelNameQwen_Max:
		return "qwen-max"
	case qwen.ModelNameQwen_QwQ_32B_Preview:
		return "qwq-32b-preview"
	default:
		return qwenclient.DefaultTextModelPath
	}
}
