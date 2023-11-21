package schema

import (
	"encoding/json"
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
	// ChatMessageTypeFunction is a message sent by a function.
	ChatMessageTypeFunction ChatMessageType = "function"
)

// ChatMessage represents a message in a chat.
type ChatMessage interface {
	// GetType gets the type of the message.
	GetType() ChatMessageType
	// GetContent gets the content of the message.
	GetContent() string
}

// Named is an interface for objects that have a name.
type Named interface {
	GetName() string
}

// ContentList is an interface for objects that have a list of content.
type ContentList interface {
	GetContentList() []ChatMessageContentPart
}

// Statically assert that the types implement the interface.
var (
	_ ChatMessage = AIChatMessage{}
	_ ChatMessage = HumanChatMessage{}
	_ ChatMessage = SystemChatMessage{}
	_ ChatMessage = GenericChatMessage{}
	_ ChatMessage = FunctionChatMessage{}
	_ ChatMessage = CompoundChatMessage{}
)

// ContentType is the type of content in a message.
type ContentType string

const (
	// ContentTypeText is text.
	ContentTypeText ContentType = "text"
	// ContentTypeImage is an image.
	ContentTypeImage ContentType = "image_url"
)

// AIChatMessage is a message sent by an AI.
type AIChatMessage struct {
	// Content is the content of the message.
	Content string

	// FunctionCall represents the model choosing to call a function.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
}

func (m AIChatMessage) GetType() ChatMessageType { return ChatMessageTypeAI }
func (m AIChatMessage) GetContent() string       { return m.Content }

// HumanChatMessage is a message sent by a human.
type HumanChatMessage struct {
	Content string
}

func (m HumanChatMessage) GetType() ChatMessageType { return ChatMessageTypeHuman }
func (m HumanChatMessage) GetContent() string       { return m.Content }

// SystemChatMessage is a chat message representing information that should be instructions to the AI system.
type SystemChatMessage struct {
	Content string
}

func (m SystemChatMessage) GetType() ChatMessageType { return ChatMessageTypeSystem }
func (m SystemChatMessage) GetContent() string       { return m.Content }

// GenericChatMessage is a chat message with an arbitrary speaker.
type GenericChatMessage struct {
	Content string
	Role    string
	Name    string
}

func (m GenericChatMessage) GetType() ChatMessageType { return ChatMessageTypeGeneric }
func (m GenericChatMessage) GetContent() string       { return m.Content }
func (m GenericChatMessage) GetName() string          { return m.Name }

// FunctionChatMessage is a chat message representing the result of a function call.
type FunctionChatMessage struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// FunctionCall is the name and arguments of a function call.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (m FunctionChatMessage) GetType() ChatMessageType { return ChatMessageTypeFunction }
func (m FunctionChatMessage) GetContent() string       { return m.Content }
func (m FunctionChatMessage) GetName() string          { return m.Name }

// CompoundChatMessage is a chat message with multiple parts.
type CompoundChatMessage struct {
	Type  ChatMessageType
	Parts []ChatMessageContentPart
}

func (m CompoundChatMessage) GetType() ChatMessageType { return m.Type }
func (m CompoundChatMessage) GetContent() string {
	b, _ := json.Marshal(m.Parts)
	return string(b)
}

func (m CompoundChatMessage) GetContentList() []ChatMessageContentPart {
	return m.Parts
}

// ChatMessageContentPart is a part of a chat message.
type ChatMessageContentPart interface {
	isChatMessageContentPart()
}

// ChatMessageContentPartText is a text part of a chat message.
type ChatMessageContentPartText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ChatMessageContentPartImageURL is an image part of a chat message.
type ChatMessageContentPartImageURL struct {
	URL     string      `json:"url"`
	Details ImageDetail `json:"detail,omitempty"`
}

// ImageDetail is the detail of an image.
type ImageDetail string

const (
	// ImageDetailAuto is the default image detail.
	ImageDetailAuto ImageDetail = "auto"
	// ImageDetailLow is the low image detail.
	ImageDetailLow ImageDetail = "low"
	// ImageDetailHigh is the high image detail.
	ImageDetailHigh ImageDetail = "high"
)

func (ChatMessageContentPartText) isChatMessageContentPart() {}

// ChatMessageContentPartImage is an image part of a chat message.
type ChatMessageContentPartImage struct {
	Type     string                         `json:"type"`
	ImageURL ChatMessageContentPartImageURL `json:"image_url,omitempty"`
}

func (ChatMessageContentPartImage) isChatMessageContentPart() {}

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
	result := []string{}
	for _, m := range messages {
		role, err := getMessageRole(m, humanPrefix, aiPrefix)
		if err != nil {
			return "", err
		}
		msg := fmt.Sprintf("%s: %s", role, m.GetContent())
		if m, ok := m.(AIChatMessage); ok && m.FunctionCall != nil {
			j, err := json.Marshal(m.FunctionCall)
			if err != nil {
				return "", err
			}
			msg = fmt.Sprintf("%s %s", msg, string(j))
		}
		result = append(result, msg)
	}
	return strings.Join(result, "\n"), nil
}

func getMessageRole(m ChatMessage, humanPrefix, aiPrefix string) (string, error) {
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
	case ChatMessageTypeFunction:
		role = "Function"
	default:
		return "", ErrUnexpectedChatMessageType
	}
	return role, nil
}
