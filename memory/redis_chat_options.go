package memory

import (
	"context"
	"github.com/tmc/langchaingo/schema"
)

const DefaultKeyExpire = 900

// RedisChatMessageHistoryOption is a function for creating new chat message history
// with other then the default values.
type RedisChatMessageHistoryOption func(r *RedisChatMessageHistory)

// WithRedisPreviousMessages is an option for NewRedisChatMessageHistory for adding
// previous messages to the history.
func WithRedisPreviousMessages(previousMessages []schema.ChatMessage) RedisChatMessageHistoryOption {
	return func(r *RedisChatMessageHistory) {
		r.SetMessages(context.Background(), previousMessages)
	}
}

// WithRedisConfOptions is an option for NewRedisChatMessageHistory for adding
// options to the redisConfOptions.
func WithRedisConfOptions(options RedisConfOptions) RedisChatMessageHistoryOption {
	return func(r *RedisChatMessageHistory) {
		r.redisConfOptions = options
	}
}

func applyRedisChatOptions(options ...RedisChatMessageHistoryOption) *RedisChatMessageHistory {
	h := &RedisChatMessageHistory{
		messages: make([]schema.ChatMessage, 0),
		redisConfOptions: RedisConfOptions{
			Address:   "localhost:6379",
			Password:  "",
			Db:        0,
			Ttl:       DefaultKeyExpire,
			KeyPrefix: "message_store:",
		},
	}

	for _, option := range options {
		option(h)
	}

	return h
}
