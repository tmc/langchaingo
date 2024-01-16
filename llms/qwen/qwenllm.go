package qwen

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	qwen_client "github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Dashscope API key, set it in the DASHSCOPE_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")

	ErrIncompleteEmbedding = errors.New("no all input got emmbedded")
)

type LLM struct {
	CallbackHandler callbacks.Handler
	client          *qwen_client.QwenClient
	options         options
}

var _ llms.LLM = (*LLM)(nil)

func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client := qwen_client.NewQwenClient(o.model, qwen_client.NewHttpClient())

	return &LLM{client: client, options: o}, nil
}

// Call implements llms.LLM.
func (q *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := q.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

// Generate implements llms.LLM.
func (q *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	if q.CallbackHandler != nil {
		q.CallbackHandler.HandleLLMStart(ctx, prompts)
	}
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))

	var model string
	if opts.Model != "" {
		model = string(qwen_client.ChoseQwenModel(opts.Model))
	} else {
		model = string(q.client.Model)
	}

	for _, prompt := range prompts {
		req := &qwen_client.QwenRequest{}

		input := qwen_client.Input{
			Messages: []qwen_client.Message{
				{
					Role:    "user",
					Content: prompt,
				},
			},
		}
		params := qwen_client.DefaultParameters()

		params.
			SetMaxTokens(opts.MaxTokens).
			SetTemperature(opts.Temperature).
			SetTopP(opts.TopP).
			SetTopK(opts.TopK).
			SetSeed(opts.Seed)

		req.
			SetModel(model).
			SetInput(input).
			SetParameters(params).
			SetStreamingFunc(opts.StreamingFunc)

		rsp, err := q.client.CreateCompletion(ctx, req)
		if err != nil {
			if q.CallbackHandler != nil {
				q.CallbackHandler.HandleLLMError(ctx, err)
			}
			return nil, err
		}

		if len(rsp.Output.Choices) == 0 {
			return generations, nil
		}

		generations = append(generations, &llms.Generation{
			Text: rsp.Output.Choices[0].Message.Content,
		})

		if q.CallbackHandler != nil {
			q.CallbackHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
		}
	}

	return generations, nil
}

func (q *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	input := struct {
		Texts []string `json:"texts"`
	}{
		Texts: inputTexts,
	}
	embeddings, err := q.client.CreateEmbedding(ctx,
		&qwen_client.EmbeddingRequest{
			Input: input,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(embeddings) != len(inputTexts) {
		return nil, ErrIncompleteEmbedding
	}

	return embeddings, nil
}
