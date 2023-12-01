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

// nolint:cyclop
func (r RedisChatMessageHistory) CoverMessageList(message string) (schemaMessage schema.ChatMessage, err error) {
	messageType, err := r.getMessageType(message)
	if err != nil {
		return schemaMessage, err
	}
	return r.unmarshalMessage(messageType, message)
}

func (r RedisChatMessageHistory) unmarshalMessage(messageType schema.ChatMessageType, message string) (schema.ChatMessage, error) {
	var chatMessage schema.ChatMessage

	switch messageType {
	case schema.ChatMessageTypeAI:
		return r.unmarshalSpecificMessage(&chatMessage, message, &schema.AIChatMessage{})
	case schema.ChatMessageTypeHuman:
		return r.unmarshalSpecificMessage(&chatMessage, message, &schema.HumanChatMessage{})
	case schema.ChatMessageTypeSystem:
		return r.unmarshalSpecificMessage(&chatMessage, message, &schema.SystemChatMessage{})
	case schema.ChatMessageTypeFunction:
		return r.unmarshalSpecificMessage(&chatMessage, message, &schema.FunctionChatMessage{})
	case schema.ChatMessageTypeGeneric:
		return r.unmarshalSpecificMessage(&chatMessage, message, &schema.GenericChatMessage{})
	}

	return chatMessage, nil
}

func (r RedisChatMessageHistory) unmarshalSpecificMessage(chatMessage *schema.ChatMessage, message string, specificMessage interface{}) (schema.ChatMessage, error) {
	if unmarshalErr := json.Unmarshal([]byte(message), specificMessage); unmarshalErr != nil {
		return *chatMessage, unmarshalErr
	}
	return *chatMessage, nil
}

func (r RedisChatMessageHistory) getMessageType(message string) (schemaType schema.ChatMessageType, err error) {
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
