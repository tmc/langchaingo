package memory

type ChatMessage interface {
	getType() string
	getText() string
}

type AiChatMessage struct {
	Text string
}

func (m AiChatMessage) getType() string { return "ai" }
func (m AiChatMessage) getText() string { return m.Text }

type HumanChatMessage struct {
	Text string
}

func (m HumanChatMessage) getType() string { return "ai" }
func (m HumanChatMessage) getText() string { return m.Text }

type SystemChatMessage struct {
	Text string
}

func (m SystemChatMessage) getType() string { return "system" }
func (m SystemChatMessage) getText() string { return m.Text }

type ChatMessageHistory struct {
	messages []ChatMessage
}

func (h *ChatMessageHistory) GetMessages() []ChatMessage { return h.messages }
func (h *ChatMessageHistory) AddAiMessage(text string) {
	h.messages = append(h.messages, AiChatMessage{Text: text})
}
func (h *ChatMessageHistory) AddUserMessage(text string) {
	h.messages = append(h.messages, HumanChatMessage{Text: text})
}

func NewChatMessageHistory(options ...NewChatMessageOption) *ChatMessageHistory {
	h := &ChatMessageHistory{
		messages: make([]ChatMessage, 0),
	}

	for _, option := range options {
		option(h)
	}

	return h
}

type NewChatMessageOption func(m *ChatMessageHistory)

// Option for NewChatMessageHistory adding previous messages to the history
func WithPreviousMessages(previousMessages []ChatMessage) NewChatMessageOption {
	return func(m *ChatMessageHistory) {
		m.messages = append(m.messages, previousMessages...)
	}
}
