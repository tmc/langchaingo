package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt" //nolint:goimports,gofumpt,gci
	"github.com/go-redis/redis"
	"github.com/tmc/langchaingo/schema"
	"time" //nolint:goimports,gofumpt,gci
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
	if err == nil && r.redisConfOptions.TTl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTl))
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
	if err == nil && r.redisConfOptions.TTl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTl))
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
	if err == nil && r.redisConfOptions.TTl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.TTl))
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
	var messageList []schema.ChatMessage
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
			messageList = append(messageList, chatMessage)
		}
	}
	return messageList, nil
}

func (r RedisChatMessageHistory) CoverMessageList(message string) (schemaMessage schema.ChatMessage, err error) {
	chatMessageMap := make(map[string]interface{})
	if unmarshalErr := json.Unmarshal([]byte(message), &chatMessageMap); unmarshalErr != nil {
		return schemaMessage, unmarshalErr
	}
	messageType, ok := chatMessageMap["type"].(string)
	if !ok {
		return schemaMessage, ErrInvalidMessageType
	}
	chatMessageType := schema.ChatMessageType(messageType)
	switch chatMessageType {
	case schema.ChatMessageTypeAI:
		var aiChatMessage schema.AIChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &aiChatMessage); unmarshalErr != nil {
			return aiChatMessage, unmarshalErr
		}
		return aiChatMessage, nil
	case schema.ChatMessageTypeHuman:
		var humanChatMessage schema.HumanChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &humanChatMessage); unmarshalErr != nil {
			return humanChatMessage, unmarshalErr
		}
		return humanChatMessage, nil
	case schema.ChatMessageTypeSystem:
		var systemChatMessage schema.SystemChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &systemChatMessage); unmarshalErr != nil {
			return systemChatMessage, unmarshalErr
		}
		return systemChatMessage, nil
	case schema.ChatMessageTypeFunction:
		var functionChatMessage schema.FunctionChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &functionChatMessage); unmarshalErr != nil {
			return functionChatMessage, unmarshalErr
		}
		return functionChatMessage, nil
	case schema.ChatMessageTypeGeneric:
		var genericChatMessage schema.GenericChatMessage
		if unmarshalErr := json.Unmarshal([]byte(message), &genericChatMessage); unmarshalErr != nil {
			return genericChatMessage, unmarshalErr
		}
		return genericChatMessage, nil
	}
	return schemaMessage, nil
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
