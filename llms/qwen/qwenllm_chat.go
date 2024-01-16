package qwen

import (
	"context"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	qwen_client "github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
	"github.com/tmc/langchaingo/schema"
)

type Chat struct {
	CallbackHandler callbacks.Handler
	client          *qwen_client.QwenClient
	options         options
}

var _ llms.ChatLLM = (*Chat)(nil)

func NewChat(opts ...Option) (*Chat, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client := qwen_client.NewQwenClient(o.model, qwen_client.NewHttpClient())

	return &Chat{client: client, options: o}, nil
}

// Call implements llms.ChatLLM.
func (q *Chat) Call(ctx context.Context, messageSets []schema.ChatMessage, options ...llms.CallOption) (*schema.AIChatMessage, error) { //nolint:lll
	r, err := q.Generate(ctx, [][]schema.ChatMessage{messageSets}, options...)
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, ErrEmptyResponse
	}
	return r[0].Message, nil
}

// Generate implements llms.ChatLLM.
func (q *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { //nolint:lll
	if q.CallbackHandler != nil {
		q.CallbackHandler.HandleLLMStart(ctx, q.getPromptsFromMessageSets(messageSets))
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	var model string
	if opts.Model != "" {
		model = string(qwen_client.ChoseQwenModel(opts.Model))
	} else {
		model = string(q.client.Model)
	}

	generations := make([]*llms.Generation, 0, len(messageSets))

	for _, message := range messageSets {
		qwenMessages := messagesToQwenMessages(message)

		imput := qwen_client.Input{
			Messages: qwenMessages,
		}

		params := qwen_client.DefaultParameters()
		params.
			SetMaxTokens(opts.MaxTokens).
			SetTemperature(opts.Temperature).
			SetTopP(opts.TopP).
			SetTopK(opts.TopK).
			SetSeed(opts.Seed)

		req := &qwen_client.QwenRequest{}
		req.
			SetModel(model).
			SetInput(imput).
			SetParameters(params).
			SetStreamingFunc(opts.StreamingFunc)

		rsp, err := q.client.CreateCompletion(ctx, req)
		if err != nil {
			if q.CallbackHandler != nil {
				q.CallbackHandler.HandleLLMError(ctx, err)
			}
			return nil, err
		}

		gen := makeGenerationFromQwenResponse(rsp)
		generations = append(generations, gen)
	}

	if q.CallbackHandler != nil {
		q.CallbackHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}
	return generations, nil
}

func (q *Chat) getPromptsFromMessageSets(messageSets [][]schema.ChatMessage) []string {
	prompts := make([]string, 0, len(messageSets))
	for i := 0; i < len(messageSets); i++ {
		curPrompt := ""
		for j := 0; j < len(messageSets[i]); j++ {
			curPrompt += messageSets[i][j].GetContent()
		}
		prompts = append(prompts, curPrompt)
	}

	return prompts
}

func messagesToQwenMessages(messages []schema.ChatMessage) []qwen_client.Message {
	qwenMessages := make([]qwen_client.Message, len(messages))

	for i, m := range messages {
		qmsg := qwen_client.Message{}
		mtype := m.GetType()

		// nolint:exhaustive
		switch mtype {
		case schema.ChatMessageTypeSystem:
			qmsg.Role = "system"
		case schema.ChatMessageTypeAI:
			qmsg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			qmsg.Role = "user"
		case schema.ChatMessageTypeGeneric:
			qmsg.Role = "user"
		}

		qmsg.Content = m.GetContent()

		qwenMessages[i] = qmsg
	}
	return qwenMessages
}

func makeGenerationFromQwenResponse(resp *qwen_client.QwenOutputMessage) *llms.Generation {
	if len(resp.Output.Choices) == 0 {
		return nil
	}

	text := resp.Output.Choices[0].Message.Content

	gen := &llms.Generation{
		Text: text,
		Message: &schema.AIChatMessage{
			Content: text,
		},
		GenerationInfo: make(map[string]interface{}),
	}

	gen.GenerationInfo["CompletionTokens"] = resp.Usage.OutputTokens
	gen.GenerationInfo["PromptTokens"] = resp.Usage.InputTokens
	gen.GenerationInfo["TotalTokens"] = resp.Usage.OutputTokens + resp.Usage.InputTokens

	return gen
}
