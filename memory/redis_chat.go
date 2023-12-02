package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt" // nolint:goimports,gofumpt,gci
	"github.com/go-redis/redis"
	"github.com/tmc/langchaingo/schema"
	"time" // nolint:goimports,gofumpt,gci
)

var ErrInvalidMessageType = errors.New("invalid message type")

// RedisChatMessageHistory is a struct that stores chat messages.
type RedisChatMessageHistory struct {
	messages         []schema.ChatMessage
	redisConfOptions RedisConfOptions
}

type RedisChatMessage struct {
	Content string                 `json:"Content"`
	Type    schema.ChatMessageType `json:"type"`
}

// Statically assert that RedisChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &RedisChatMessageHistory{}

func NewRedisChatMessageHistory(options ...RedisChatMessageHistoryOption) *RedisChatMessageHistory {
	return applyRedisChatOptions(options...)
}

// AddUserMessage adds an user to the chat message history.
func (r RedisChatMessageHistory) AddUserMessage(_ context.Context, message string) error {
	chatMessage := RedisChatMessage{Content: message, Type: schema.ChatMessageTypeHuman}
	redisClient, err := r.getRedisClientIns()
	if err != nil {
		return err
	}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.TTL > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTL))
	}
	return err
}

// AddAIMessage adds an AIMessage to the chat message history.
func (r RedisChatMessageHistory) AddAIMessage(_ context.Context, message string) error {
	chatMessage := RedisChatMessage{Content: message, Type: schema.ChatMessageTypeAI}
	redisClient, err := r.getRedisClientIns()
	if err != nil {
		return err
	}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.TTL > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTL))
	}
	return err
}

func (r RedisChatMessageHistory) AddMessage(_ context.Context, message schema.ChatMessage) error {
	redisClient, err := r.getRedisClientIns()
	if err != nil {
		return err
	}
	chatMessage := RedisChatMessage{Content: message.GetContent(), Type: message.GetType()}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.TTL > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTL))
	}
	return err
}

func (r RedisChatMessageHistory) Clear(_ context.Context) error {
	redisClient, err := r.getRedisClientIns()
	if err != nil {
		return err
	}
	_, err = redisClient.Del(r.GetRedisKsy()).Result()
	return err
}

// Messages returns all messages stored.
func (r RedisChatMessageHistory) Messages(_ context.Context) ([]schema.ChatMessage, error) {
	redisClient, err := r.getRedisClientIns()
	var messageList []schema.ChatMessage // nolint:prealloc
	if err != nil {
		return messageList, err
	}
	messages, err := redisClient.LRange(r.GetRedisKsy(), 0, -1).Result()
	if err != nil {
		return messageList, err
	}
	if len(messages) == 0 {
		return messageList, nil
	}
	for _, message := range messages {
		chatMessage, coverErr := r.CoverMessageList(message)
		if coverErr != nil {
			return messageList, coverErr
		}
		messageList = append(messageList, chatMessage)
	}
	return messageList, nil
}

func (r RedisChatMessageHistory) CoverMessageList(message string) (
	schema.ChatMessage, error) { // nolint:gofumpt
	messageType, err := r.getMessageType(message)
	if err != nil {
		var emptyMessage schema.ChatMessage
		return emptyMessage, err
	}
	return r.unmarshalMessage(messageType, message)
}

// nolint:cyclop
func (r RedisChatMessageHistory) unmarshalMessage(messageType schema.ChatMessageType, message string) (
	schema.ChatMessage, error) { // nolint:gofumpt
	var empty schema.ChatMessage
	switch messageType {
	case schema.ChatMessageTypeAI:
		var aiMsg schema.AIChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &aiMsg); unmarshalErr != nil {
			return empty, unmarshalErr
		}
		return aiMsg, nil
	case schema.ChatMessageTypeHuman:
		var humanMsg schema.HumanChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &humanMsg); unmarshalErr != nil {
			return empty, unmarshalErr
		}
		return humanMsg, nil
	case schema.ChatMessageTypeSystem:
		var systemMsg schema.SystemChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &systemMsg); unmarshalErr != nil {
			return empty, unmarshalErr
		}
		return systemMsg, nil
	case schema.ChatMessageTypeFunction:
		var funMsg schema.FunctionChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &funMsg); unmarshalErr != nil {
			return empty, unmarshalErr
		}
		return funMsg, nil
	case schema.ChatMessageTypeGeneric:
		var genMsg schema.GenericChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &genMsg); unmarshalErr != nil {
			return empty, unmarshalErr
		}
		return genMsg, nil
	}
	return empty, nil
}

func (r RedisChatMessageHistory) getMessageType(message string) ( // nolint:nonamedreturns
	schemaType schema.ChatMessageType, err error) { // nolint:gofumpt
	chatMessageMap := make(map[string]interface{})
	if unmarshalErr := json.Unmarshal([]byte(message), &chatMessageMap); unmarshalErr != nil {
		return schemaType, unmarshalErr
	}
	messageType, ok := chatMessageMap["type"].(string)
	if !ok {
		return schemaType, ErrInvalidMessageType
	}
	schemaType = schema.ChatMessageType(messageType)
	return schemaType, nil
}

func (r RedisChatMessageHistory) SetMessages(context context.Context, messages []schema.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}
	for _, message := range messages {
		addMessageErr := r.AddMessage(context, message)
		if addMessageErr != nil {
			return addMessageErr
		}
	}
	return nil
}

func (r RedisChatMessageHistory) GetRedisKsy() string {
	return fmt.Sprintf("%s:%s", r.redisConfOptions.KeyPrefix, r.redisConfOptions.SessionID)
}

func (r RedisChatMessageHistory) getRedisClientIns() (*redis.Client, error) {
	return redisClientIns.GetClient(r.redisConfOptions)
}
