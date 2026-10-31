package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// represents the length of context.
const shortTermMemory = 15

var memory = Memory{} //nolint:gochecknoglobals

type Memory struct {
	messages []llms.MessageContent
}

// AddMessage adds messages to the struct.
func (m *Memory) AddMessage(role llms.ChatMessageType, content string) {
	m.messages = append(m.messages, llms.TextParts(role, content))
	if len(m.messages) > shortTermMemory {
		m.messages = m.messages[len(m.messages)-shortTermMemory:]
	}
}

// GetContext handles the type system context.
func (m *Memory) GetContext() []llms.MessageContent {
	return m.messages
}

func main() {
	llm, err := openai.New(
		openai.WithModel("gpt-4o-mini"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Sending initial message to the model.
	// Using AddMessage function to add human messages to short term memory.
	memory.AddMessage(llms.ChatMessageTypeHuman, "List 10 reasons to work with Go in AI Engineering.")

	// Using GetContext function to add system context to short term memory.
	content := append([]llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a friendly assistant working on the artificial intelligence engineering team."),
	},
		memory.GetContext()...)

	completion, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(1024),  // example of use case for chatbots tokens limits.
		llms.WithTemperature(0.7), // example of temperature to "control" LLM response hallucinations.
		// temperature doesn't actually control hallucinations, but it helps a little.
		llms.WithStreamingFunc(showResponse),
	)
	if err != nil {
		log.Fatal(err)
	}

	llmResponse := completion.Choices[0].Content

	// Using AddMessage.
	memory.AddMessage(llms.ChatMessageTypeAI, llmResponse)
}

// showResponse print chunks for streaming func.
func showResponse(_ context.Context, chunk []byte) error {
	fmt.Print(string(chunk))
	return nil
}
