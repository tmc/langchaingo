package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/schema"
	"time"
)

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
	redisClient, err := redisClientIns.GetClient(r.redisConfOptions)
	if err != nil {
		return err
	}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.Ttl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.Ttl))
	}
	return err
}

// AddAIMessage adds an AIMessage to the chat message history.
func (r RedisChatMessageHistory) AddAIMessage(_ context.Context, message string) error {
	chatMessage := RedisChatMessage{Content: message, Type: schema.ChatMessageTypeAI}
	redisClient, err := redisClientIns.GetClient(r.redisConfOptions)
	if err != nil {
		return err
	}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.Ttl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.Ttl))
	}
	return err
}

func (r RedisChatMessageHistory) AddMessage(_ context.Context, message schema.ChatMessage) error {
	redisClient, err := redisClientIns.GetClient(r.redisConfOptions)
	if err != nil {
		return err
	}
	chatMessage := RedisChatMessage{Content: message.GetContent(), Type: message.GetType()}
	messageByte, err := json.Marshal(chatMessage)
	if err != nil {
		return err
	}
	_, err = redisClient.RPush(r.GetRedisKsy(), messageByte).Result()
	if err == nil && r.redisConfOptions.Ttl > 0 {
		redisClient.Expire(r.GetRedisKsy(), time.Second*time.Duration(r.redisConfOptions.Ttl))
	}
	return err
}

func (r RedisChatMessageHistory) Clear(_ context.Context) error {
	redisClient, err := redisClientIns.GetClient(r.redisConfOptions)
	if err != nil {
		return err
	}
	_, err = redisClient.Del(r.GetRedisKsy()).Result()
	return err
}

// Messages returns all messages stored.
func (r RedisChatMessageHistory) Messages(_ context.Context) ([]schema.ChatMessage, error) {
	redisClient, err := redisClientIns.GetClient(r.redisConfOptions)
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
		chatMessageMap := make(map[string]interface{})
		if unmarshalErr := json.Unmarshal([]byte(message), &chatMessageMap); unmarshalErr != nil {
			return messageList, unmarshalErr
		}
		messageType, ok := chatMessageMap["type"].(string)
		if !ok {
			return messageList, fmt.Errorf("invalid message type")
		}
		chatMessageType := schema.ChatMessageType(messageType)
		if chatMessageType == schema.ChatMessageTypeAI {
			var aiChatMessage schema.AIChatMessage
			if unmarshalErr := json.Unmarshal([]byte(message), &aiChatMessage); unmarshalErr != nil {
				return messageList, unmarshalErr
			}
			messageList = append(messageList, aiChatMessage)
		} else if chatMessageType == schema.ChatMessageTypeHuman {
			var humanChatMessage schema.HumanChatMessage
			if unmarshalErr := json.Unmarshal([]byte(message), &humanChatMessage); unmarshalErr != nil {
				return messageList, unmarshalErr
			}
			messageList = append(messageList, humanChatMessage)
		} else if chatMessageType == schema.ChatMessageTypeSystem {
			var systemChatMessage schema.SystemChatMessage
			if unmarshalErr := json.Unmarshal([]byte(message), &systemChatMessage); unmarshalErr != nil {
				return messageList, unmarshalErr
			}
			messageList = append(messageList, systemChatMessage)
		} else if chatMessageType == schema.ChatMessageTypeFunction {
			var functionChatMessage schema.FunctionChatMessage
			if unmarshalErr := json.Unmarshal([]byte(message), &functionChatMessage); unmarshalErr != nil {
				return messageList, unmarshalErr
			}
			messageList = append(messageList, functionChatMessage)
		} else if chatMessageType == schema.ChatMessageTypeGeneric {
			var genericChatMessage schema.GenericChatMessage
			if unmarshalErr := json.Unmarshal([]byte(message), &genericChatMessage); unmarshalErr != nil {
				return messageList, unmarshalErr
			}
			messageList = append(messageList, genericChatMessage)
		}
	}
	return messageList, nil
}

func (r RedisChatMessageHistory) SetMessages(_ context.Context, messages []schema.ChatMessage) error {
	if len(messages) == 0 {
		return nil
	}
	for _, message := range messages {
		addMessageErr := r.AddMessage(context.Background(), message)
		if addMessageErr != nil {
			return addMessageErr
		}
	}
	return nil
}

func (r RedisChatMessageHistory) GetRedisKsy() string {
	return fmt.Sprintf("%s:%s", r.redisConfOptions.KeyPrefix, r.redisConfOptions.SessionId)
}
