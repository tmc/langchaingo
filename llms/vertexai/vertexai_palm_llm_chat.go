package vertexai

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/vertexai/internal/vertexaiclient"
	"github.com/tmc/langchaingo/schema"
)

const (
	userAuthor = "user"
	botAuthor  = "bot"
)

type ChatMessage = vertexaiclient.ChatMessage

type Chat struct {
	client *vertexaiclient.PaLMClient
}

var _ llms.ChatLLM = (*Chat)(nil)

// Chat requests a chat response for the given messages.
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

// Generate requests a chat response for each of the sets of messages.
func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint: lll
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}
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
		result, err := o.client.CreateChat(ctx, &vertexaiclient.ChatRequest{
			Temperature: opts.Temperature,
			Messages:    msgs,
			Context:     chatContext,
		})
		if err != nil {
			return nil, err
		}
		if len(result.Candidates) == 0 {
			return nil, ErrEmptyResponse
		}
		generations = append(generations, &llms.Generation{
			Message: &schema.AIChatMessage{
				Content: result.Candidates[0].Content,
			},
			Text: result.Candidates[0].Content,
		})
	}

	return generations, nil
}

func toClientChatMessage(messages []schema.ChatMessage) []*vertexaiclient.ChatMessage {
	msgs := make([]*vertexaiclient.ChatMessage, len(messages))

	for i, m := range messages {
		msg := &vertexaiclient.ChatMessage{
			Content: m.GetContent(),
		}
		typ := m.GetType()

		switch typ {
		case schema.ChatMessageTypeSystem:
			msg.Author = botAuthor
		case schema.ChatMessageTypeAI:
			msg.Author = botAuthor
		case schema.ChatMessageTypeHuman:
			msg.Author = userAuthor
		case schema.ChatMessageTypeGeneric:
			msg.Author = userAuthor
		case schema.ChatMessageTypeFunction:
			msg.Author = userAuthor
		}
		if n, ok := m.(schema.Named); ok {
			msg.Author = n.GetName()
		}
		msgs[i] = msg
	}
	return msgs
}

func parseContext(messages []schema.ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	// check if 1st message type is system. use it as context.
	if messages[0].GetType() == schema.ChatMessageTypeSystem {
		return messages[0].GetContent()
	}
	return ""
}

// NewChat returns a new VertexAI PaLM Chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	client, err := newClient(opts...)
	return &Chat{client: client}, err
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &vertexaiclient.EmbeddingRequest{
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
