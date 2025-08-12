package mongo

import (
	"context"

	"github.com/tmc/langchaingo/internal/mongodb"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// mongoSessionIDKey a unique identifier of the session, like user name, email, chat id etc.
	// same as langchain.
	mongoSessionIDKey = "SessionId"
)

type ChatMessageHistory struct {
	url            string
	sessionID      string
	databaseName   string
	collectionName string
	client         *mongo.Client
	collection     *mongo.Collection
}

type chatMessageModel struct {
	SessionID string                  `bson:"SessionId"`
	History   chatMessageModelHistory `bson:"History"`
}

type chatMessageModelHistory struct {
	Type    string `bson:"type"`
	Content string `bson:"content"`
}

// Statically assert that MongoDBChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &ChatMessageHistory{}

// NewMongoDBChatMessageHistory creates a new MongoDBChatMessageHistory using chat message options.
func NewMongoDBChatMessageHistory(ctx context.Context, options ...ChatMessageHistoryOption) (*ChatMessageHistory, error) {
	h, err := applyMongoDBChatOptions(options...)
	if err != nil {
		return nil, err
	}

	client, err := mongodb.NewClient(ctx, h.url)
	if err != nil {
		return nil, err
	}

	h.client = client

	h.collection = client.Database(h.databaseName).Collection(h.collectionName)
	// create session id index
	if _, err := h.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: mongoSessionIDKey, Value: 1}}}); err != nil {
		return nil, err
	}

	return h, nil
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	messages := []llms.ChatMessage{}
	filter := bson.M{mongoSessionIDKey: h.sessionID}
	cursor, err := h.collection.Find(ctx, filter)
	if err != nil {
		return messages, err
	}

	_messages := []chatMessageModel{}
	if err := cursor.All(ctx, &_messages); err != nil {
		return messages, err
	}
	for _, message := range _messages {
		m := llms.ChatMessageModel{
			Type: message.History.Type,
			Data: llms.ChatMessageModelData{
				Content: message.History.Content,
			},
		}
		messages = append(messages, m.ToChatMessage())
	}

	return messages, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(ctx context.Context, text string) error {
	return h.AddMessage(ctx, llms.AIChatMessage{Content: text})
}

// AddUserMessage adds a user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(ctx context.Context, text string) error {
	return h.AddMessage(ctx, llms.HumanChatMessage{Content: text})
}

// Clear clear session memory from MongoDB.
func (h *ChatMessageHistory) Clear(ctx context.Context) error {
	filter := bson.M{mongoSessionIDKey: h.sessionID}
	_, err := h.collection.DeleteMany(ctx, filter)
	return err
}

// AddMessage adds a message to the store.
func (h *ChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	_, err := h.collection.InsertOne(ctx, chatMessageModel{
		SessionID: h.sessionID,
		History: chatMessageModelHistory{
			Type:    string(message.GetType()),
			Content: string(message.GetContent()),
		},
	})

	return err
}

// SetMessages replaces existing messages in the store.
func (h *ChatMessageHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	_messages := []interface{}{}
	for _, message := range messages {
		_messages = append(_messages, chatMessageModel{
			SessionID: h.sessionID,
			History: chatMessageModelHistory{
				Type:    string(message.GetType()),
				Content: string(message.GetContent()),
			},
		})
	}

	if err := h.Clear(ctx); err != nil {
		return err
	}

	_, err := h.collection.InsertMany(ctx, _messages)
	return err
}
