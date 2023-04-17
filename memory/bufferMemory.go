package memory

type BufferMemory struct {
	ChatHistory    *ChatMessageHistory
	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AiPrefix       string
	MemoryKey      string
}

func (m BufferMemory) SaveContext(inputValues InputValues, outputValues InputValues) error {
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

func (m BufferMemory) LoadMemoryVariables(inputValuesGiven InputValues) (InputValues, error) {
	if m.ReturnMessages {
		return InputValues{
			m.MemoryKey: m.ChatHistory.messages,
		}, nil
	}

	return InputValues{
		m.MemoryKey: getBufferString(m.ChatHistory.messages, m.HumanPrefix, m.AiPrefix),
	}, nil
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
