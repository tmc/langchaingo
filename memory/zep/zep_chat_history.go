package zep

import (
	"context"
	"fmt"
	"log"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/getzep/zep-go/v2"
	zepClient "github.com/getzep/zep-go/v2/client"
)

// ChatMessageHistory is a struct that stores chat messages.
type ChatMessageHistory struct {
	ZepClient   *zepClient.Client
	SessionID   string
	MemoryType  zep.MemoryType
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

func (h *ChatMessageHistory) messagesFromZepMessages(zepMessages []*zep.Message) []llms.ChatMessage {
	var chatMessages []llms.ChatMessage
	for _, zepMessage := range zepMessages {
		switch zepMessage.RoleType { // nolint We do not store other message types in zep memory
		case zep.RoleTypeUserRole:
			chatMessages = append(chatMessages, llms.HumanChatMessage{Content: zepMessage.Content})
		case zep.RoleTypeAssistantRole:
			chatMessages = append(chatMessages, llms.AIChatMessage{Content: zepMessage.Content})
		case zep.RoleTypeToolRole:
		case zep.RoleTypeFunctionRole:
			chatMessages = append(chatMessages, llms.ToolChatMessage{Content: zepMessage.Content})
		default:
			log.Print(fmt.Errorf("unknown role: %s", zepMessage.RoleType))
			continue
		}
	}
	return chatMessages
}

func (h *ChatMessageHistory) messagesToZepMessages(messages []llms.ChatMessage) []*zep.Message {
	var zepMessages []*zep.Message //nolint We don't know the final size of the messages as some might be skipped due to unsupported role.
	for _, m := range messages {
		zepMessage := zep.Message{
			Content: m.GetContent(),
		}
		switch m.GetType() { // nolint We only expect to bring these three types into chat history
		case llms.ChatMessageTypeHuman:
			zepMessage.RoleType = zep.RoleTypeUserRole
			if h.HumanPrefix != "" {
				zepMessage.Role = zep.String(h.HumanPrefix)
			}
		case llms.ChatMessageTypeAI:
			zepMessage.RoleType = zep.RoleTypeAssistantRole
			if h.AIPrefix != "" {
				zepMessage.Role = zep.String(h.AIPrefix)
			}
		case llms.ChatMessageTypeFunction:
			zepMessage.RoleType = zep.RoleTypeFunctionRole
		case llms.ChatMessageTypeTool:
			zepMessage.RoleType = zep.RoleTypeToolRole
		default:
			log.Print(fmt.Errorf("unknown role: %s", zepMessage.RoleType))
			continue
		}
		zepMessages = append(zepMessages, &zepMessage)
	}
	return zepMessages
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	memory, err := h.ZepClient.Memory.Get(ctx, h.SessionID, nil)
	if err != nil {
		return nil, err
	}
	messages := h.messagesFromZepMessages(memory.Messages)

	// Use the new Context field if available
	systemPromptContent := ""
	if memory.Context != nil && *memory.Context != "" {
		systemPromptContent = *memory.Context
	} else if memory.RelevantFacts != nil {
		// Use RelevantFacts if Context is not available
		for _, fact := range memory.RelevantFacts {
			systemPromptContent += fmt.Sprintf("%s\n", fact.Content)
		}
	}

	if systemPromptContent != "" {
		// Add system prompt to the beginning of the messages.
		messages = append(
			[]llms.ChatMessage{
				llms.SystemChatMessage{
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
	_, err := h.ZepClient.Memory.Add(ctx, h.SessionID, &zep.AddMemoryRequest{
		Messages: h.messagesToZepMessages(
			[]llms.ChatMessage{
				llms.AIChatMessage{Content: text},
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
	_, err := h.ZepClient.Memory.Add(ctx, h.SessionID, &zep.AddMemoryRequest{
		Messages: h.messagesToZepMessages(
			[]llms.ChatMessage{
				llms.HumanChatMessage{Content: text},
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

func (h *ChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	_, err := h.ZepClient.Memory.Add(ctx, h.SessionID, &zep.AddMemoryRequest{
		Messages: h.messagesToZepMessages([]llms.ChatMessage{message}),
	})
	if err != nil {
		return err
	}
	return nil
}

func (*ChatMessageHistory) SetMessages(_ context.Context, _ []llms.ChatMessage) error {
	return nil
}
