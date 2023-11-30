package ernie

import (
	"context"
	"os"
	"reflect"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie/internal/ernieclient"
	"github.com/tmc/langchaingo/schema"
)

type ChatMessage = ernieclient.ChatMessage

type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *ernieclient.Client
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

func NewChat(opts ...Option) (*Chat, error) {
	options := &options{
		apiKey:    os.Getenv(ernieAPIKey),
		secretKey: os.Getenv(ernieSecretKey),
	}

	for _, opt := range opts {
		opt(options)
	}

	c, err := newClient(options)
	if err != nil {
		return nil, err
	}
	c.ModelPath = modelToPath(ModelName(c.Model))

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
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, getPromptsFromMessageSets(messageSets))
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		req := &ernieclient.ChatRequest{
			Model:            opts.Model,
			StopWords:        opts.StopWords,
			Messages:         messagesToClientMessages(messageSet),
			StreamingFunc:    opts.StreamingFunc,
			Temperature:      opts.Temperature,
			MaxTokens:        opts.MaxTokens,
			N:                opts.N, // TODO: note, we are not returning multiple completions
			FrequencyPenalty: opts.FrequencyPenalty,
			PresencePenalty:  opts.PresencePenalty,
			System:           getSystem(messageSet),

			FunctionCallBehavior: ernieclient.FunctionCallBehavior(opts.FunctionCallBehavior),
		}
		for _, fn := range opts.Functions {
			req.Functions = append(req.Functions, ernieclient.FunctionDefinition{
				Name:        fn.Name,
				Description: fn.Description,
				Parameters:  fn.Parameters,
			})
		}
		result, err := o.client.CreateChat(ctx, req)
		if err != nil {
			return nil, err
		}

		if result.Result == "" && result.FunctionCall == nil {
			return nil, ErrEmptyResponse
		}

		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		msg := &schema.AIChatMessage{
			Content: result.Result,
		}

		if result.FunctionCall != nil {
			msg.FunctionCall = &schema.FunctionCall{
				Name:      result.FunctionCall.Name,
				Arguments: result.FunctionCall.Arguments,
			}
		}
		generations = append(generations, &llms.Generation{
			Message:        msg,
			Text:           msg.Content,
			GenerationInfo: generationInfo,
		})
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return llms.CountTokens(o.client.Model, text)
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
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

func messagesToClientMessages(messages []schema.ChatMessage) []*ernieclient.ChatMessage {
	msgs := make([]*ernieclient.ChatMessage, 0)
	for _, m := range messages {
		msg := &ernieclient.ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem: // In Ernie's 'messages' parameter, there is no 'system' role.
			continue
		case schema.ChatMessageTypeAI:
			msg.Role = "assistant"
		case schema.ChatMessageTypeHuman:
			msg.Role = "user"
		case schema.ChatMessageTypeGeneric:
			msg.Role = "user"
		case schema.ChatMessageTypeFunction:
			msg.Role = "function"
		}

		if n, ok := m.(FunctionCalled); ok {
			msg.FunctionCall = n.GetFunctionCall()
		}

		if n, ok := m.(schema.Named); ok {
			msg.Name = n.GetName()
		}
		msgs = append(msgs, msg)
	}

	return msgs
}

// getSystem Retrieve system parameter from messages.
func getSystem(messages []schema.ChatMessage) string {
	for _, message := range messages {
		if message.GetType() == schema.ChatMessageTypeSystem {
			return message.GetContent()
		}
	}
	return ""
}

// FunctionCalled is an interface for objects that have a function call info.
type FunctionCalled interface {
	GetFunctionCall() *schema.FunctionCall
}
