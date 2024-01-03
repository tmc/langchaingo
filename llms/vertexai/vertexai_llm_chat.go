package vertexai

import (
	"context"
	"github.com/tmc/langchaingo/llms/vertexai/internal/schema"

	"github.com/tmc/langchaingo/llms"
	lcgschema "github.com/tmc/langchaingo/schema"
)

const (
	userAuthor = "user"
	botAuthor  = "bot"
)

type ChatMessage = schema.ChatMessage

type Chat struct {
	*baseLLM
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// Call requests a chat response for the given messages.
func (o *Chat) Call(ctx context.Context, messages []lcgschema.ChatMessage, options ...llms.CallOption) (*lcgschema.AIChatMessage, error) { // nolint: lll
	r, err := o.Generate(ctx, [][]lcgschema.ChatMessage{messages}, options...)
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, ErrEmptyResponse
	}
	return r[0].Message, nil
}

// Generate requests a chat response for each of the sets of messages.
func (o *Chat) Generate(ctx context.Context, messageSets [][]lcgschema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint: lll
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	o.setDefaultCallOptions(&opts)

	if opts.StreamingFunc != nil {
		return nil, ErrNotImplemented
	}

	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messages := range messageSets {
		chatContext := parseContext(messages)
		if len(chatContext) > 0 {
			// remove system context from messages
			messages = messages[1:]
		}
		msgs := toClientChatMessage(messages)
		result, err := o.client.CreateChat(ctx, opts.Model, o.Publisher, &schema.ChatRequest{
			Temperature: float32(opts.Temperature),
			Messages:    msgs,
			Context:     chatContext,
			TopK:        float32(opts.TopK),
			TopP:        float32(int(opts.TopP)),
		})
		if err != nil {
			return nil, err
		}
		if len(result.Candidates) == 0 {
			return nil, ErrEmptyResponse
		}
		generations = append(generations, &llms.Generation{
			Message: &lcgschema.AIChatMessage{
				Content: result.Candidates[0].Content,
			},
			Text: result.Candidates[0].Content,
		})
	}

	return generations, nil
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []lcgschema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

func toClientChatMessage(messages []lcgschema.ChatMessage) []*schema.ChatMessage {
	msgs := make([]*schema.ChatMessage, len(messages))

	for i, m := range messages {
		msg := &schema.ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()

		switch typ {
		case lcgschema.ChatMessageTypeSystem:
			msg.Author = botAuthor
		case lcgschema.ChatMessageTypeAI:
			msg.Author = botAuthor
		case lcgschema.ChatMessageTypeHuman:
			msg.Author = userAuthor
		case lcgschema.ChatMessageTypeGeneric:
			msg.Author = userAuthor
		case lcgschema.ChatMessageTypeFunction:
			msg.Author = userAuthor
		}
		if n, ok := m.(lcgschema.Named); ok {
			msg.Author = n.GetName()
		}
		msgs[i] = msg
	}
	return msgs
}

func parseContext(messages []lcgschema.ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	// check if 1st message type is system. use it as context.
	if messages[0].GetType() == lcgschema.ChatMessageTypeSystem {
		return messages[0].GetContent()
	}
	return ""
}

// NewChat returns a new VertexAI PaLM Chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	// The context should be provided by the caller but that would be a big change, so we just do this
	ctx := context.Background()

	// Ensure options are initialized only once.
	initOptions.Do(initOpts)
	options := &options{}
	*options = *defaultOptions // Copy default options.

	for _, opt := range opts {
		opt(options)
	}

	// The Chat struct uses the chatModel provided in the options, so we configure the base with that
	base, err := newBase(ctx, options.chatModel, *options)

	return &Chat{baseLLM: base}, err
}
