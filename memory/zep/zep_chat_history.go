package zep

import (
	"context"
	"fmt"
	"github.com/getzep/zep-go"
	zepClient "github.com/getzep/zep-go/client"
	"github.com/rs/zerolog/log"
	"github.com/tmc/langchaingo/schema"
)

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	ZepClient   *zepClient.Client
	SessionID   string
	MemoryType  zep.MemoryGetRequestMemoryType
	HumanPrefix string
	AIPrefix    string
}

// Statically assert that ZepChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &ChatMessageHistory{}

// NewZepChatMessageHistory creates a new ZepChatMessageHistory using chat message options.
func NewZepChatMessageHistory(zep *zepClient.Client, sessionID string, options ...ChatMessageHistoryOption) *ChatMessageHistory {
	messageHistory := applyZepChatHistoryOptions(options...)
	messageHistory.ZepClient = zep
	messageHistory.SessionID = sessionID
	return messageHistory
}

func (h *ChatMessageHistory) messagesFromZepMessages(zepMessages []*zep.Message) []schema.ChatMessage {
	var chatMessages []schema.ChatMessage
	for _, zepMessage := range zepMessages {
		switch *zepMessage.RoleType {
		case zep.ModelsRoleTypeUserRole:
			chatMessages = append(chatMessages, schema.HumanChatMessage{Content: *zepMessage.Content})
		case zep.ModelsRoleTypeAssistantRole:
			chatMessages = append(chatMessages, schema.AIChatMessage{Content: *zepMessage.Content})
		case zep.ModelsRoleTypeFunctionRole:
			chatMessages = append(chatMessages, schema.ToolChatMessage{Content: *zepMessage.Content})
		default:
			log.Err(fmt.Errorf("unknown role: %s", *zepMessage.RoleType))
			continue
		}
	}
	return chatMessages
}

func (h *ChatMessageHistory) messagesToZepMessages(messages []schema.ChatMessage) []*zep.Message {
	var zepMessages []*zep.Message
	for _, m := range messages {
		zepMessage := zep.Message{
			Content: zep.String(m.GetContent()),
		}
		switch m.GetType() {
		case schema.ChatMessageTypeHuman:
			zepMessage.RoleType = zep.ModelsRoleTypeUserRole.Ptr()
			if h.HumanPrefix != "" {
				zepMessage.Role = zep.String(h.HumanPrefix)
			}
		case schema.ChatMessageTypeAI:
			zepMessage.RoleType = zep.ModelsRoleTypeAssistantRole.Ptr()
			if h.AIPrefix != "" {
				zepMessage.Role = zep.String(h.AIPrefix)
			}
		case schema.ChatMessageTypeFunction:
			zepMessage.RoleType = zep.ModelsRoleTypeFunctionRole.Ptr()
		default:
			log.Err(fmt.Errorf("unknown role: %s", *zepMessage.RoleType))
			continue
		}
		zepMessages = append(zepMessages, &zepMessage)
	}
	return zepMessages
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(ctx context.Context) ([]schema.ChatMessage, error) {
	memory, err := h.ZepClient.Memory.Get(ctx, h.SessionID, &zep.MemoryGetRequest{
		MemoryType: h.MemoryType.Ptr(),
	})
	if err != nil {
		return nil, err
	}
	messages := h.messagesFromZepMessages(memory.Messages)
	zepFacts := memory.Facts
	systemPromptContent := ""
	for _, fact := range zepFacts {
		systemPromptContent += fmt.Sprintf("%s\n", fact)
	}
	if memory.Summary != nil && memory.Summary.Content != nil {
		systemPromptContent += fmt.Sprintf("%s\n", *memory.Summary.Content)
	}
	if systemPromptContent != "" {
		// Add system prompt to the beginning of the messages.
		messages = append(
			[]schema.ChatMessage{
				schema.SystemChatMessage{
					Content: systemPromptContent,
				},
			},
			messages...,
		)
	}
	return messages, nil
}

// AddAIMessage adds an AIMessage to the chat message history.
func (h *ChatMessageHistory) AddAIMessage(ctx context.Context, text string) error {
	err := h.ZepClient.Memory.Create(ctx, h.SessionID, &zep.Memory{
		Messages: h.messagesToZepMessages(
			[]schema.ChatMessage{
				schema.AIChatMessage{Content: text},
			},
		),
	})
	if err != nil {
		return err
	}
	return nil
}

// AddUserMessage adds a user to the chat message history.
func (h *ChatMessageHistory) AddUserMessage(ctx context.Context, text string) error {
	err := h.ZepClient.Memory.Create(ctx, h.SessionID, &zep.Memory{
		Messages: h.messagesToZepMessages(
			[]schema.ChatMessage{
				schema.HumanChatMessage{Content: text},
			},
		),
	})
	if err != nil {
		return err
	}
	return nil
}

func (h *ChatMessageHistory) Clear(ctx context.Context) error {
	_, err := h.ZepClient.Memory.Delete(ctx, h.SessionID)
	if err != nil {
		return err
	}
	return nil
}

func (h *ChatMessageHistory) AddMessage(ctx context.Context, message schema.ChatMessage) error {
	err := h.ZepClient.Memory.Create(ctx, h.SessionID, &zep.Memory{
		Messages: h.messagesToZepMessages([]schema.ChatMessage{message}),
	})
	if err != nil {
		return err
	}
	return nil
}

func (*ChatMessageHistory) SetMessages(_ context.Context, _ []schema.ChatMessage) error {
	return nil
}
