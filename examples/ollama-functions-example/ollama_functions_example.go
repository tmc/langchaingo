package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	// allow specifying your own model via OLLAMA_TEST_MODEL
	// (same as the Ollama unit tests).
	model := "llama3.1"
	if v := os.Getenv("OLLAMA_TEST_MODEL"); v != "" {
		model = v
	}

	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in Beijing and Shenzhen?"),
	}

	fmt.Println("Querying for weather in Beijing and Shenzhen.")
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}
	messageHistory = updateMessageHistory(messageHistory, resp)

	// Execute tool calls requested by the model
	messageHistory = executeToolCalls(messageHistory, resp)
	messageHistory = append(messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, "Can you compare the two?"))

	// Send query to the model again, this time with a history containing its
	// request to invoke a tool and our response to the tool call.
	fmt.Println("Querying with tool response...")
	resp, err = llm.GenerateContent(ctx, messageHistory)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(showResponse(resp))
}

// updateMessageHistory updates the message history with the assistant's
// response and requested tool calls.
func updateMessageHistory(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	respchoice := resp.Choices[0]

	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	return append(messageHistory, assistantResponse)
}

// executeToolCalls executes the tool calls in the response and returns the
// updated message history.
func executeToolCalls(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	fmt.Println("Executing", len(resp.Choices[0].ToolCalls), "tool calls")
	for _, toolCall := range resp.Choices[0].ToolCalls {
		switch toolCall.FunctionCall.Name {
		case "getCurrentWeather":
			var args struct {
				Location string `json:"location"`
				Unit     string `json:"unit"`
			}
			if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
				log.Fatal(err)
			}

			response, err := getCurrentWeather(args.Location, args.Unit)
			if err != nil {
				log.Fatal(err)
			}

			weatherCallResponse := llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Name:    toolCall.FunctionCall.Name,
						Content: response,
					},
				},
			}
			messageHistory = append(messageHistory, weatherCallResponse)
		default:
			log.Fatalf("Unsupported tool: %s", toolCall.FunctionCall.Name)
		}
	}

	return messageHistory
}

type Weather struct {
	Location string `json:"location"`
	Forecast string `json:"forecast"`
}

func getCurrentWeather(location string, unit string) (string, error) {

	var weatherInfo Weather
	switch location {
	case "Shenzhen":
		weatherInfo = Weather{Location: location, Forecast: "74 and cloudy"}
	case "Beijing":
		weatherInfo = Weather{Location: location, Forecast: "80 and rainy"}
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// availableTools simulates the tools/functions we're making available for
// the model.
var availableTools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city, e.g. San Francisco",
					},
					"unit": map[string]any{
						"type": "string",
						"enum": []string{"fahrenheit", "celsius"},
					},
				},
				"required": []string{"location", "unit"},
			},
		},
	},
}

func showResponse(resp *llms.ContentResponse) string {
	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}
