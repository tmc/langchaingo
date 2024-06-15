package mongo

import (
	"errors"
)

const (
	mongoDefaultDBName         = "chat_history"
	mongoDefaultCollectionName = "message_store"
)

var (
	errMongoInvalidURL       = errors.New("invalid mongo url option")
	errMongoInvalidSessionID = errors.New("invalid mongo session id option")
)

type ChatMessageHistoryOption func(m *ChatMessageHistory)

func applyMongoDBChatOptions(options ...ChatMessageHistoryOption) (*ChatMessageHistory, error) {
	h := &ChatMessageHistory{
		databaseName:   mongoDefaultDBName,
		collectionName: mongoDefaultCollectionName,
	}

	for _, option := range options {
		option(h)
	}

	if h.url == "" {
		return nil, errMongoInvalidURL
	}
	if h.sessionID == "" {
		return nil, errMongoInvalidSessionID
	}

	return h, nil
}

// WithConnectionURL is an option for specifying the MongoDB connection URL. Must be set.
func WithConnectionURL(connectionURL string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.url = connectionURL
	}
}

// WithSessionID is an arbitrary key that is used to store the messages of a single chat session,
// like user name, email, chat id etc. Must be set.
func WithSessionID(sessionID string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.sessionID = sessionID
	}
}

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.collectionName = name
	}
}

// WithDataBaseName is an option for specifying the database name.
func WithDataBaseName(name string) ChatMessageHistoryOption {
	return func(p *ChatMessageHistory) {
		p.databaseName = name
	}
}
