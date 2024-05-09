package memory

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

type MongoDBChatMessageHistoryOption func(m *MongoDBChatMessageHistory)

func applyMongoDBChatOptions(options ...MongoDBChatMessageHistoryOption) (*MongoDBChatMessageHistory, error) {
	h := &MongoDBChatMessageHistory{
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
func WithConnectionURL(connectionURL string) MongoDBChatMessageHistoryOption {
	return func(p *MongoDBChatMessageHistory) {
		p.url = connectionURL
	}
}

// WithSessionID is an arbitrary key that is used to store the messages of a single chat session,
// like user name, email, chat id etc. Must be set.
func WithSessionID(sessionID string) MongoDBChatMessageHistoryOption {
	return func(p *MongoDBChatMessageHistory) {
		p.sessionID = sessionID
	}
}

// WithCollectionName is an option for specifying the collection name.
func WithCollectionName(name string) MongoDBChatMessageHistoryOption {
	return func(p *MongoDBChatMessageHistory) {
		p.collectionName = name
	}
}

// WithDataBaseName is an option for specifying the database name.
func WithDataBaseName(name string) MongoDBChatMessageHistoryOption {
	return func(p *MongoDBChatMessageHistory) {
		p.databaseName = name
	}
}
