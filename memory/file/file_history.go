// Package file adds support for
// chat message history using file storage.
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// StoredMessage is the message format stored in the file
type StoredMessage struct {
	Type    llms.ChatMessageType `json:"type"`
	Content string               `json:"content"`
	Name    string               `json:"name,omitempty"`
}

// FileChatMessageHistory implements chat history storage using a file
type FileChatMessageHistory struct {
	// FilePath is the path to the file storing chat history
	FilePath string
	// CreateDirIfNotExist if true, creates directory if it doesn't exist
	CreateDirIfNotExist bool
	// mutex protects file operations
	mutex sync.Mutex
}

// Statically assert that FileChatMessageHistory implement the chat message history interface.
var _ schema.ChatMessageHistory = &FileChatMessageHistory{}

// NewFileChatMessageHistory creates a new file-based chat message history
func NewFileChatMessageHistory(options ...FileChatMessageHistoryOption) (*FileChatMessageHistory, error) {
	h := applyChatOptions(options...)

	// Ensure directory exists if needed
	if h.CreateDirIfNotExist {
		dir := filepath.Dir(h.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create an empty JSON array file if it doesn't exist
	if _, err := os.Stat(h.FilePath); os.IsNotExist(err) {
		if err := os.WriteFile(h.FilePath, []byte("[]"), 0644); err != nil {
			return nil, fmt.Errorf("failed to create history file: %w", err)
		}
	}

	return h, nil
}

// loadMessages loads messages from the file
func (h *FileChatMessageHistory) loadMessages() ([]StoredMessage, error) {
	data, err := os.ReadFile(h.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var messages []StoredMessage
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return messages, nil
}

// saveMessages saves messages to the file
func (h *FileChatMessageHistory) saveMessages(messages []StoredMessage) error {
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize messages: %w", err)
	}

	if err := os.WriteFile(h.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// convertToStoredMessage converts ChatMessage to StoredMessage
func convertToStoredMessage(message llms.ChatMessage) StoredMessage {
	var msgType llms.ChatMessageType
	var name string

	switch message.GetType() {
	case llms.ChatMessageTypeHuman:
		msgType = llms.ChatMessageTypeHuman
	case llms.ChatMessageTypeAI:
		msgType = llms.ChatMessageTypeAI
	case llms.ChatMessageTypeSystem:
		msgType = llms.ChatMessageTypeSystem
	default:
		msgType = llms.ChatMessageTypeGeneric
		name = string(message.GetType())
	}

	return StoredMessage{
		Type:    msgType,
		Content: message.GetContent(),
		Name:    name,
	}
}

// convertToChatMessage converts StoredMessage to ChatMessage
func convertToChatMessage(message StoredMessage) llms.ChatMessage {
	switch message.Type {
	case llms.ChatMessageTypeHuman:
		return llms.HumanChatMessage{Content: message.Content}
	case llms.ChatMessageTypeAI:
		return llms.AIChatMessage{Content: message.Content}
	case llms.ChatMessageTypeSystem:
		return llms.SystemChatMessage{Content: message.Content}
	default:
		return llms.GenericChatMessage{
			Role:    message.Name,
			Content: message.Content,
		}
	}
}

// AddMessage adds a message to the history
func (h *FileChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	messages, err := h.loadMessages()
	if err != nil {
		return err
	}

	storedMessage := convertToStoredMessage(message)
	messages = append(messages, storedMessage)

	return h.saveMessages(messages)
}

// AddUserMessage adds a user message to the history
func (h *FileChatMessageHistory) AddUserMessage(ctx context.Context, message string) error {
	return h.AddMessage(ctx, llms.HumanChatMessage{Content: message})
}

// AddAIMessage adds an AI message to the history
func (h *FileChatMessageHistory) AddAIMessage(ctx context.Context, message string) error {
	return h.AddMessage(ctx, llms.AIChatMessage{Content: message})
}

// Clear clears the history
func (h *FileChatMessageHistory) Clear(ctx context.Context) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.saveMessages([]StoredMessage{})
}

// Messages gets all history messages
func (h *FileChatMessageHistory) Messages(ctx context.Context) ([]llms.ChatMessage, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storedMessages, err := h.loadMessages()
	if err != nil {
		return nil, err
	}

	messages := make([]llms.ChatMessage, len(storedMessages))
	for i, msg := range storedMessages {
		messages[i] = convertToChatMessage(msg)
	}

	return messages, nil
}

// SetMessages sets the history messages
func (h *FileChatMessageHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	storedMessages := make([]StoredMessage, len(messages))
	for i, msg := range messages {
		storedMessages[i] = convertToStoredMessage(msg)
	}

	return h.saveMessages(storedMessages)
}
