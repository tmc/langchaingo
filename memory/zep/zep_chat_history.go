package zep

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"

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

func (h *ChatMessageHistory) messagesFromZepMessages(zepMessages []*zep.Message) []llms.ChatMessage {
	var chatMessages []llms.ChatMessage
	for _, zepMessage := range zepMessages {
		switch *zepMessage.RoleType { // nolint We do not store other message types in zep memory
		case zep.RoleTypeUserRole:
			chatMessages = append(chatMessages, llms.HumanChatMessage{Content: *zepMessage.Content})
		case zep.RoleTypeAssistantRole:
			chatMessages = append(chatMessages, llms.AIChatMessage{Content: *zepMessage.Content})
		case zep.RoleTypeToolRole:
		case zep.RoleTypeFunctionRole:
			chatMessages = append(chatMessages, llms.ToolChatMessage{Content: *zepMessage.Content})
		default:
			log.Print(fmt.Errorf("unknown role: %s", *zepMessage.RoleType))
			continue
		}
	}
	return chatMessages
}

func (h *ChatMessageHistory) messagesToZepMessages(messages []llms.ChatMessage) []*zep.Message {
	var zepMessages []*zep.Message //nolint We don't know the final size of the messages as some might be skipped due to unsupported role.
	for _, m := range messages {
		zepMessage := zep.Message{
			Content: zep.String(m.GetContent()),
		}
		switch m.GetType() { // nolint We only expect to bring these three types into chat history
		case llms.ChatMessageTypeHuman:
			zepMessage.RoleType = zep.RoleTypeUserRole.Ptr()
			if h.HumanPrefix != "" {
				zepMessage.Role = zep.String(h.HumanPrefix)
			}
		case llms.ChatMessageTypeAI:
			zepMessage.RoleType = zep.RoleTypeAssistantRole.Ptr()
			if h.AIPrefix != "" {
				zepMessage.Role = zep.String(h.AIPrefix)
			}
		case llms.ChatMessageTypeFunction:
			zepMessage.RoleType = zep.RoleTypeFunctionRole.Ptr()
		case llms.ChatMessageTypeTool:
			zepMessage.RoleType = zep.RoleTypeToolRole.Ptr()
		default:
			log.Print(fmt.Errorf("unknown role: %s", *zepMessage.RoleType))
			continue
		}
		zepMessages = append(zepMessages, &zepMessage)
	}
	return zepMessages
}

// Messages returns all messages stored.
func (h *ChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
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
