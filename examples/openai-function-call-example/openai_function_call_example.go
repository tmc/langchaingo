package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := openai.NewChat(openai.WithModel("gpt-3.5-turbo-1106"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Content: "What is the weather like in Boston?"},
	}, llms.WithFunctions(functions))
	if err != nil {
		log.Fatal(err)
	}

	if completion.FunctionCall != nil {
		fmt.Printf("Function call: %v\n", completion.FunctionCall)
	}
}

// Get the current weather in the specified location.
func getCurrentWeather(req struct {
	// The city and state, e.g. San Francisco, CA.
	location string
	// The temperature unit, either "celcius" or "fahrenheit".
	unit string
}) (string, error) {
	weatherInfo := map[string]interface{}{
		"location":    req.location,
		"temperature": "72",
		"unit":        req.unit,
		"forecast":    []string{"sunny", "windy"},
	}
	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var functions = []llms.FunctionDefinition{
	{
		Name:        "getCurrentWeather",
		Description: "Get the current weather in a given location",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"req":{"type":"object","properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA."},"unit":{"type":"string","description":"The temperature unit, either \"celcius\" or \"fahrenheit\"."}}}},"required":["req","unit"]}`),
	},
}
