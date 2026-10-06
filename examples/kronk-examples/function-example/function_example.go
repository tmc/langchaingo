// function_example demonstrates tool calling using the kronk abstraction
// layer. It passes native LangChainGo tool definitions to a local GGUF model,
// executes the requested tool, and returns the result for a final answer.
//
// The first time you run this program the system will download and install
// the llama.cpp libraries and the model. Subsequent runs load from disk.
//
// Run from this directory:
//
//	go run function_example.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

var flagVerbose = flag.Bool("v", false, "verbose mode")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	modelSource := "unsloth/Qwen3-0.6B-Q8_0"
	if v := os.Getenv("KRONK_TEST_MODEL"); v != "" {
		modelSource = v
	}

	fmt.Println("Initializing kronk (first run downloads libraries and model)...")
	client, err := kronk.New(
		context.Background(),
		modelSource,
		kronk.WithAutoTune(true),
	)
	if err != nil {
		return fmt.Errorf("create kronk client: %w", err)
	}

	defer func() {
		fmt.Println("\nUnloading Kronk")
		if err := client.Unload(context.Background()); err != nil {
			fmt.Printf("failed to unload model: %v", err)
		}
	}()

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "Use the available tool to answer weather questions."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in Beijing?"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	resp, err := client.GenerateContent(ctx, messages,
		llms.WithTools(tools),
		llms.WithToolChoice("required"),
		llms.WithStreamingFunc(func(context.Context, []byte) error { return nil }),
	)
	cancel()
	if err != nil {
		return fmt.Errorf("generate tool call: %w", err)
	}
	if len(resp.Choices) == 0 || len(resp.Choices[0].ToolCalls) == 0 {
		return errors.New("model returned no tool calls")
	}

	choice := resp.Choices[0]
	assistantParts := make([]llms.ContentPart, 0, len(choice.ToolCalls)+1)
	if choice.Content != "" {
		assistantParts = append(assistantParts, llms.TextPart(choice.Content))
	}
	for _, toolCall := range choice.ToolCalls {
		assistantParts = append(assistantParts, toolCall)
	}
	messages = append(messages, llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: assistantParts})

	for _, toolCall := range choice.ToolCalls {
		log.Printf("Call: %s(%s)", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
		result, err := dispatchCall(toolCall)
		if err != nil {
			return err
		}
		messages = append(messages, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Name:       toolCall.FunctionCall.Name,
				Content:    result,
			}},
		})
	}

	ctx, cancel = context.WithTimeout(context.Background(), 120*time.Second)
	resp, err = client.GenerateContent(ctx, messages,
		llms.WithTools(tools),
		llms.WithToolChoice("none"),
	)
	cancel()
	if err != nil {
		return fmt.Errorf("generate final answer: %w", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		return errors.New("model returned an empty final answer")
	}

	log.Printf("Final response: %s", resp.Choices[0].Content)
	if *flagVerbose {
		log.Printf("Generation info: %#v", resp.Choices[0].GenerationInfo)
	}
	return nil
}

func dispatchCall(call llms.ToolCall) (string, error) {
	if call.FunctionCall == nil || call.FunctionCall.Name != "getCurrentWeather" {
		return "", fmt.Errorf("unsupported tool call %#v", call.FunctionCall)
	}

	var input struct {
		Location string `json:"location"`
		Unit     string `json:"unit"`
	}
	if err := json.Unmarshal([]byte(call.FunctionCall.Arguments), &input); err != nil {
		return "", fmt.Errorf("decode getCurrentWeather arguments: %w", err)
	}
	if input.Location == "" || input.Unit == "" {
		return "", errors.New("getCurrentWeather requires location and unit")
	}

	return getCurrentWeather(input.Location, input.Unit)
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherInfo := map[string]any{
		"location":    location,
		"temperature": "6",
		"unit":        unit,
		"forecast":    []string{"sunny", "windy"},
	}
	if unit == "fahrenheit" {
		weatherInfo["temperature"] = 43
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var tools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: json.RawMessage(`{
			"type": "object", 
			"properties": {
				"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, 
				"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
			}, 
			"required": ["location", "unit"]
		}`),
		},
	},
}
