package openai

import (
	"context"
	"reflect"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/schema"
)

type ChatMessage = openaiclient.ChatMessage

type Chat struct {
	client *openaiclient.Client
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		client: c,
	}, err
}

// Call requests a chat response for the given messages.
func (o *Chat) Call(ctx context.Context, messages []schema.ChatMessage, options ...llms.CallOption) (*schema.AIChatMessage, error) { // nolint: lll
	r, err := o.Generate(ctx, [][]schema.ChatMessage{messages}, options...)
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, ErrEmptyResponse
	}
	return r[0].Message, nil
}

//nolint:funlen
func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint:lll,cyclop
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		msgs := make([]*openaiclient.ChatMessage, len(messageSet))
		for i, m := range messageSet {
			msg := &openaiclient.ChatMessage{
				Content: m.GetContent(),
			}
			typ := m.GetType()
			switch typ {
			case schema.ChatMessageTypeSystem:
				msg.Role = "system"
			case schema.ChatMessageTypeAI:
				msg.Role = "assistant"
				if aiChatMsg, ok := m.(schema.AIChatMessage); ok && aiChatMsg.FunctionCall != nil {
					msg.FunctionCall = &openaiclient.FunctionCall{
						Name:      aiChatMsg.FunctionCall.Name,
						Arguments: aiChatMsg.FunctionCall.Arguments,
					}
				}
			case schema.ChatMessageTypeHuman:
				msg.Role = "user"
			case schema.ChatMessageTypeGeneric:
				msg.Role = "user"
			case schema.ChatMessageTypeFunction:
				msg.Role = "function"
			}
			if n, ok := m.(schema.Named); ok {
				msg.Name = n.GetName()
			}
			msgs[i] = msg
		}
		req := &openaiclient.ChatRequest{
			Model:            opts.Model,
			StopWords:        opts.StopWords,
			Messages:         msgs,
			StreamingFunc:    opts.StreamingFunc,
			Temperature:      opts.Temperature,
			MaxTokens:        opts.MaxTokens,
			N:                opts.N,
			FrequencyPenalty: opts.FrequencyPenalty,
			PresencePenalty:  opts.PresencePenalty,

			FunctionCallBehavior: openaiclient.FunctionCallBehavior(opts.FunctionCallBehavior),
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
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Choices[0].Message.Content,
		}
		if result.Choices[0].FinishReason == "function_call" {
			msg.FunctionCall = &schema.FunctionCall{
				Name:      result.Choices[0].Message.FunctionCall.Name,
				Arguments: result.Choices[0].Message.FunctionCall.Arguments,
			}
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
	}

	return generations, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return llms.CountTokens(o.client.Model, text)
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
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
