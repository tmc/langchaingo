package openai

import (
	"context"
	"fmt"
	"reflect"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/schema"
)

type ChatMessage = openaiclient.ChatMessage

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *openaiclient.Client
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
)

var _ llms.ChatLLM = (*Chat)(nil)

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	opt, c, err := newClient(opts...)
	if err != nil {
		return nil, err
	}
	return &Chat{
		client:           c,
		CallbacksHandler: opt.callbackHandler,
	}, err
}

// GenerateContent implements the Model interface.
//
//nolint:goerr113
func (o *Chat) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop
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
	}

	return &llms.ContentResponse{Choices: choices}, nil
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
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, getPromptsFromMessageSets(messageSets))
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		req := &openaiclient.ChatRequest{
			Model:            opts.Model,
			StopWords:        opts.StopWords,
			Messages:         messagesToClientMessages(messageSet),
			StreamingFunc:    opts.StreamingFunc,
			Temperature:      opts.Temperature,
			MaxTokens:        opts.MaxTokens,
			N:                opts.N, // TODO: note, we are not returning multiple completions
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
			Content: fmt.Sprint(result.Choices[0].Message.Content),
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
			StopReason:     result.Choices[0].FinishReason,
		})
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
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

func getPromptsFromMessageSets(messageSets [][]schema.ChatMessage) []string {
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

func messagesToClientMessages(messages []schema.ChatMessage) []*openaiclient.ChatMessage {
	msgs := make([]*openaiclient.ChatMessage, len(messages))
	for i, m := range messages {
		msg := &openaiclient.ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			msg.Role = "system"
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
			if mm, ok := m.(schema.AIChatMessage); ok && mm.FunctionCall != nil {
				msg.FunctionCall = &openaiclient.FunctionCall{
					Name:      mm.FunctionCall.Name,
					Arguments: mm.FunctionCall.Arguments,
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
		if cl, ok := m.(schema.ContentList); ok {
			msg.Content = cl.GetContentList()
		}
		msgs[i] = msg
	}

	return msgs
}
