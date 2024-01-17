package qwen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	qwen_client "github.com/tmc/langchaingo/llms/qwen/internal/qwenclient"
	"github.com/tmc/langchaingo/schema"
)

var _ llms.Model = (*Chat)(nil)

var (
	ErrNotSupportImsgePart = errors.New("not support Image parts yet")
	ErrMultipleTextParts   = errors.New("found multiple text parts in message")
	ErrEmptyMessageContent = errors.New("TongyiMessageContent is empty")
)

type UnSupportedRoleError struct {
	Role schema.ChatMessageType
}

func (e *UnSupportedRoleError) Error() string {
	return fmt.Sprintf("qwen role %s not supported", e.Role)
}

// GenerateContent implements llms.Model.
// nolint:lll
func (q *Chat) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
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

	tongyiContents := messagesCntentToQwenMessages(messages)
	// qwenContents
	qwenTextMessages := make([]qwen_client.Message, len(tongyiContents))
	for i, tc := range tongyiContents {
		qwenTextMessages[i] = *tc.getTextMessage()
	}
	rsp, err := q.doTextGenRequest(ctx, model, qwenTextMessages, opts)
	if err != nil {
		return nil, err
	}

	choices := make([]*llms.ContentChoice, len(rsp.Output.Choices))
	for i, c := range rsp.Output.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content,
			StopReason: c.FinishReason,
			GenerationInfo: map[string]any{
				"PromptTokens":     rsp.Usage.InputTokens,
				"CompletionTokens": rsp.Usage.OutputTokens,
				"TotalTokens":      rsp.Usage.TotalTokens,
			},
		}
	}

	return &llms.ContentResponse{Choices: choices}, nil
}

func (q *Chat) doTextGenRequest(
	ctx context.Context,
	model string,
	qwenTextMessages []qwen_client.Message,
	opts llms.CallOptions,
) (*qwen_client.QwenOutputMessage, error) {
	input := qwen_client.Input{
		Messages: qwenTextMessages,
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
		return nil, ErrEmptyResponse
	}

	return rsp, nil
}

// qwen message Decorator in order to support llms.MessageContent.
type TongyiMessageContent struct {
	TextMessage *qwen_client.Message

	// TODO: intergrate tongyi-wanx and qwen-vl api to support image type
	// ImageMessage *qwen_client.ImageData
}

func (q *TongyiMessageContent) getTextMessage() *qwen_client.Message {
	return q.TextMessage
}

func (q *TongyiMessageContent) MarshalJSON() ([]byte, error) {
	if q.TextMessage != nil {
		return json.Marshal(q.TextMessage)
	}
	return nil, ErrEmptyMessageContent
}

func messagesCntentToQwenMessages(messagesContent []llms.MessageContent) []TongyiMessageContent {
	qwenMessages := make([]TongyiMessageContent, len(messagesContent))

	for i, mc := range messagesContent {
		foundText := false
		qmsg := TongyiMessageContent{}
		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				qmsg.TextMessage = &qwen_client.Message{
					Role:    typeToQwenRole(mc.Role),
					Content: pt.Text,
				}
				if foundText {
					panic(ErrMultipleTextParts)
				}
				foundText = true
			case llms.BinaryContent:
				// imageData := qwen_client.ImageData{Data: pt.Data}
				// qmsg.ImageMessage = &imageData
				panic(ErrNotSupportImsgePart)
			default:
				panic("only support Text parts right now")
			}
		}

		qwenMessages[i] = qmsg
	}
	return qwenMessages
}

func typeToQwenRole(typ schema.ChatMessageType) string {
	switch typ {
	case schema.ChatMessageTypeSystem:
		return "system"
	case schema.ChatMessageTypeAI:
		return "assistant"
	case schema.ChatMessageTypeHuman:
		return "user"
	case schema.ChatMessageTypeGeneric:
		return "user"
	case schema.ChatMessageTypeFunction:
		fallthrough
	default:
		panic(&UnSupportedRoleError{Role: typ})
	}
}
