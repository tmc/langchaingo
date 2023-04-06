package schema

type ChatMessage interface {
	GetType() string
	GetText() string
}

type AiChatMessage struct {
	Text string
}

func (m AiChatMessage) GetType() string { return "ai" }
func (m AiChatMessage) GetText() string { return m.Text }

type HumanChatMessage struct {
	Text string
}

func (m HumanChatMessage) GetType() string { return "human" }
func (m HumanChatMessage) GetText() string { return m.Text }

type SystemChatMessage struct {
	Text string
}

func (m SystemChatMessage) GetType() string { return "system" }
func (m SystemChatMessage) GetText() string { return m.Text }
