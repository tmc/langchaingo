package schema

import (
	"errors"
	"fmt"
	"strings"
)

// ChatMessageType is the type of a chat message.
type ChatMessageType string

// ErrUnexpectedChatMessageType is returned when a chat message is of an unexpected type.
var ErrUnexpectedChatMessageType = errors.New("unexpected chat message type")

const (
	// ChatMessageTypeAI is a message sent by an AI.
	ChatMessageTypeAI ChatMessageType = "ai"
	// ChatMessageTypeHuman is a message sent by a human.
	ChatMessageTypeHuman ChatMessageType = "human"
	// ChatMessageTypeSystem is a message sent by the system.
	ChatMessageTypeSystem ChatMessageType = "system"
	// ChatMessageTypeGeneric is a message sent by a generic user.
	ChatMessageTypeGeneric ChatMessageType = "generic"
)

// ChatMessage is a message sent by a user or the system.
type ChatMessage interface {
	GetText() string
	GetType() ChatMessageType
}

// Statically assert that the types implement the interface.
var (
	_ ChatMessage = AIChatMessage{}
	_ ChatMessage = HumanChatMessage{}
	_ ChatMessage = SystemChatMessage{}
	_ ChatMessage = GenericChatMessage{}
)

// AIChatMessage is a message sent by an AI.
type AIChatMessage struct {
	Text string
}

func (m AIChatMessage) GetType() ChatMessageType { return ChatMessageTypeAI }
func (m AIChatMessage) GetText() string          { return m.Text }

// HumanChatMessage is a message sent by a human.
type HumanChatMessage struct {
	Text string
}

func (m HumanChatMessage) GetType() ChatMessageType { return ChatMessageTypeHuman }
func (m HumanChatMessage) GetText() string          { return m.Text }

// SystemChatMessage is a chat message representing information that should be instructions to the AI system.
type SystemChatMessage struct {
	Text string
}

func (m SystemChatMessage) GetType() ChatMessageType { return ChatMessageTypeSystem }
func (m SystemChatMessage) GetText() string          { return m.Text }

// GenericChatMessage is a chat message with an arbitrary speaker.
type GenericChatMessage struct {
	Text string
	Role string
}

func (m GenericChatMessage) GetType() ChatMessageType { return ChatMessageTypeGeneric }
func (m GenericChatMessage) GetText() string          { return m.Text }

// ChatGeneration is the output of a single chat generation.
type ChatGeneration struct {
	Generation
	Message ChatMessage
}

// ChatResult is the class that contains all relevant information for a Chat Result.
type ChatResult struct {
	Generations []ChatGeneration
	LLMOutput   map[string]any
}

// GetBufferString gets the buffer string of messages.
func GetBufferString(messages []ChatMessage, humanPrefix string, aiPrefix string) (string, error) {
	stringMessages := []string{}
	for _, m := range messages {
		var role string
		switch m.GetType() {
		case ChatMessageTypeHuman:
			role = humanPrefix
		case ChatMessageTypeAI:
			role = aiPrefix
		case ChatMessageTypeSystem:
			role = "System"
		case ChatMessageTypeGeneric:
			cgm, ok := m.(GenericChatMessage)
			if !ok {
				return "", fmt.Errorf("%w -%+v", ErrUnexpectedChatMessageType, m)
			}
			role = cgm.Role
		default:
			return "", ErrUnexpectedChatMessageType
		}
		stringMessages = append(stringMessages, fmt.Sprintf("%s: %s", role, m.GetText()))
	}
	return strings.Join(stringMessages, "\n"), nil
}

// ChatMessageHistory is the chat history definition.
type ChatMessageHistory interface {
	Messages() ([]ChatMessage, error)
	AddAIMessage(string) error
	AddUserMessage(string) error
	Clear() error
}
