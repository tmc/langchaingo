package openai

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/schema"
)

type ChatMessage = openaiclient.ChatMessage

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *openaiclient.Client
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
)

var _ llms.Model = (*LLM)(nil)

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	opt, c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:           c,
		CallbacksHandler: opt.callbackHandler,
	}, err
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
//
//nolint:goerr113
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	chatMsgs := make([]*ChatMessage, 0, len(messages))
	for _, mc := range messages {
		msg := &ChatMessage{MultiContent: mc.Parts}
		switch mc.Role {
		case schema.ChatMessageTypeSystem:
			msg.Role = RoleSystem
		case schema.ChatMessageTypeAI:
			msg.Role = RoleAssistant
		case schema.ChatMessageTypeHuman:
			msg.Role = RoleUser
		case schema.ChatMessageTypeGeneric:
			msg.Role = RoleUser
		case schema.ChatMessageTypeFunction:
			fallthrough
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		chatMsgs = append(chatMsgs, msg)
	}

	req := &openaiclient.ChatRequest{
		Model:                opts.Model,
		StopWords:            opts.StopWords,
		Messages:             chatMsgs,
		StreamingFunc:        opts.StreamingFunc,
		Temperature:          opts.Temperature,
		MaxTokens:            opts.MaxTokens,
		N:                    opts.N,
		FrequencyPenalty:     opts.FrequencyPenalty,
		PresencePenalty:      opts.PresencePenalty,
		FunctionCallBehavior: openaiclient.FunctionCallBehavior(opts.FunctionCallBehavior),
	}

	if opts.JSONMode {
		req.ResponseFormat = openaiclient.ResponseFormatJSON
	}

	for _, fn := range opts.Functions {
		req.Functions = append(req.Functions, openaiclient.FunctionDefinition{
			Name:        fn.Name,
			Description: fn.Description,
			Parameters:  fn.Parameters,
		})
	}
	result, err := o.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(result.Choices))
	for i, c := range result.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content,
			StopReason: c.FinishReason,
			GenerationInfo: map[string]any{
				"CompletionTokens": result.Usage.CompletionTokens,
				"PromptTokens":     result.Usage.PromptTokens,
				"TotalTokens":      result.Usage.TotalTokens,
			},
		}

		if c.FinishReason == "function_call" {
			choices[i].FuncCall = &schema.FunctionCall{
				Name:      c.Message.FunctionCall.Name,
				Arguments: c.Message.FunctionCall.Arguments,
			}
		}
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
		Model: o.client.Model,
	})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}
