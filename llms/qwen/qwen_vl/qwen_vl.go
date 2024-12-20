package qwen_vl

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

var (
	imageURLKey = "image_url"
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

	contents := make([]qwenclient.VLContent, 0)
	contents = append(contents, qwenclient.VLContent{
		Text: part.(llms.TextContent).Text,
	})
	imageURL := ctx.Value(imageURLKey).([]string)
	for _, url := range imageURL {
		contents = append(contents, qwenclient.VLContent{
			Image: url,
		})
	}

	message := qwenclient.VLMessage{
		Role:    "user",
		Content: contents,
	}

	input := struct {
		Messages []qwenclient.VLMessage `json:"messages"`
	}{
		Messages: []qwenclient.VLMessage{message},
	}

	parameters := qwenclient.Parameters{
		Temperature: opts.Temperature,
	}

	result, err := o.client.CreateVLChat(ctx, &qwenclient.VLRequest{
		Model:      string(o.getModelPath(*opts)),
		Input:      input,
		Parameters: parameters,
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

	content := result.Output.Choices[0].Message.Content[0].Text
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
You can pass auth info by use qwen_vl.New(qwen.WithAK("{api Key}")) ,
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
	case qwen.ModelNameQwen_VL_Plus:
		return "qwen-vl-plus"
	case qwen.ModelNameQwen_VL_Max:
		return "qwen-vl-max"
	default:
		return qwenclient.DefaultVLModelPath
	}
}
