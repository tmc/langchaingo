package memory

import (
	"github.com/tmc/langchaingo/schema"
)

type DBAdapter interface {
	LoadDBMemory() ([]schema.ChatMessage, error)
	SaveDBContext(msgs []schema.ChatMessage, bufferString string) error
	ClearDBContext() error
	SetSessionID(id string)
	GetSessionID() string
}

type PersistentBuffer struct {
	ChatHistory *ChatMessageHistory
	DB          DBAdapter

	ReturnMessages bool
	InputKey       string
	OutputKey      string
	HumanPrefix    string
	AIPrefix       string
	MemoryKey      string
}

var _ schema.Memory = PersistentBuffer{}

func NewPersistentBuffer(dbAdapter DBAdapter) *PersistentBuffer {
	buffer := PersistentBuffer{
		DB: dbAdapter,

		ReturnMessages: false,
		InputKey:       "",
		OutputKey:      "",
		HumanPrefix:    "Human",
		AIPrefix:       "AI",
		MemoryKey:      "history",
	}

	return &buffer
}

func (buffer PersistentBuffer) MemoryVariables() []string {
	return []string{buffer.MemoryKey}
}

// LoadMemoryVariables returns the previous chat messages stored in postgreSQL database. Previous
// chat messages are loaded from db and loaded into the ChatMessageHistory. Messages are returned in
// a map with the key specified in the MemoryKey field. This key defaults to "history".
// If ReturnMessages is set to true the output is a slice of schema.ChatMessage. Otherwise
// the output is a buffer string of the chat messages.
func (buffer PersistentBuffer) LoadMemoryVariables(map[string]any) (map[string]any, error) {
	msgs, err := buffer.DB.LoadDBMemory()
	if err != nil {
		return nil, err
	}

	buffer.ChatHistory = NewChatMessageHistory(
		WithPreviousMessages(msgs),
	)

	if buffer.ReturnMessages {
		return map[string]any{
			buffer.MemoryKey: buffer.ChatHistory.Messages(),
		}, nil
	}

	bufferString, err := schema.GetBufferString(buffer.ChatHistory.Messages(), buffer.HumanPrefix, buffer.AIPrefix)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		buffer.MemoryKey: bufferString,
	}, nil
}

// SaveContext saves the context of the PostgreBuffer.
// It takes in two maps, inputs and outputs, which contain key-value pairs of any type.
// The function retrieves the value associated with buffer. InputKey from the inputs map
// and adds it as a user message to the ChatHistory. Then, it retrieves the value
// associated with buffer.OutputKey from the outputs map and adds it as an AI message
// to the ChatHistory. The function then uses the ChatHistory, HumanPrefix, and AIPrefix
// properties of the buffer to generate a bufferString using the GetBufferString function
// from the schema package. Finally, it saves the ChatHistory messages and bufferString
// to the DB using the SaveHistory function, and returns any error encountered.
func (buffer PersistentBuffer) SaveContext(inputs map[string]any, outputs map[string]any) error {
	userInputValue, err := getInputValue(inputs, buffer.InputKey)
	if err != nil {
		return err
	}

	buffer.ChatHistory.AddUserMessage(userInputValue)

	aiOutPutValue, err := getInputValue(outputs, buffer.OutputKey)
	if err != nil {
		return err
	}

	buffer.ChatHistory.AddAIMessage(aiOutPutValue)

	bufferString, err := schema.GetBufferString(buffer.ChatHistory.Messages(), buffer.HumanPrefix, buffer.AIPrefix)
	if err != nil {
		return err
	}

	msgs := buffer.ChatHistory.Messages()

	err = buffer.DB.SaveDBContext(msgs, bufferString)
	if err != nil {
		return err
	}

	return nil
}

// Clear clears the persistent buffer.
// It clears the DB context and chat history.
// Returns an error if there was an issue.
func (buffer PersistentBuffer) Clear() error {
	err := buffer.DB.ClearDBContext()
	if err != nil {
		return err
	}

	buffer.ChatHistory.Clear()

	return nil
}
