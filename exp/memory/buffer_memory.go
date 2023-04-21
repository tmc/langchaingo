package memory

// BufferMemory is a simple form of memory that remembers previous conversational back and forths directly.
type BufferMemory struct {
	ChatHistory    *ChatMessageHistory
	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AiPrefix       string
	MemoryKey      string
}

func (m BufferMemory) SaveContext(inputValues map[string]any, outputValues map[string]any) error {
	userInputValue, err := getInputValue(inputValues, m.InputKey)
	if err != nil {
		return err
	}

	m.ChatHistory.AddUserMessage(userInputValue)

	aiOutputValue, err := getInputValue(outputValues, m.OutputKey)
	if err != nil {
		return nil
	}

	m.ChatHistory.AddAiMessage(aiOutputValue)

	return nil
}

func (m BufferMemory) LoadMemoryVariables(inputValuesGiven map[string]any) map[string]any {
	if m.ReturnMessages {
		return map[string]any{
			m.MemoryKey: m.ChatHistory.messages,
		}
	}

	return map[string]any{
		m.MemoryKey: getBufferString(m.ChatHistory.messages, m.HumanPrefix, m.AiPrefix),
	}
}

func NewBufferMemory() BufferMemory {
	m := BufferMemory{
		ChatHistory:    NewChatMessageHistory(),
		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AiPrefix:       "AI",
		MemoryKey:      "history",
	}

	return m
}
