package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func main() {
	llm, err := anthropic.New(anthropic.WithModel("claude-3-haiku-20240307"))
	if err != nil {
		log.Fatal(err)
	}

	// Sending initial message to the model, with a list of available tools.
	ctx := context.Background()
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Boston?"),
	}

	fmt.Println("Querying for weather in Boston..")
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	// Execute tool calls requested by the model
	messageHistory = executeToolCalls(ctx, llm, messageHistory, resp)
	// messageHistory = append(messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, "Can you compare the two?"))

	// Send query to the model again, this time with a history containing its
	// request to invoke a tool and our response to the tool call.
	fmt.Println("Querying with tool response...")
	resp, err = llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Choices[0].Content)
}

// executeToolCalls executes the tool calls in the response and returns the
// updated message history.
func executeToolCalls(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	for _, choice := range resp.Choices {
		for _, toolCall := range choice.ToolCalls {

			// Append tool_use to messageHistory
			assistantResponse := llms.TextParts(llms.ChatMessageTypeAI)
			for _, tc := range choice.ToolCalls {
				assistantResponse.Parts = append(assistantResponse.Parts, tc)
			}
			messageHistory = append(messageHistory, assistantResponse)

			switch toolCall.FunctionCall.Name {
			case "getCurrentWeather":
				var args struct {
					Location string `json:"location"`
					Unit     string `json:"unit"`
				}
				if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
					log.Fatal(err)
				}

				// Perform Function Calling
				response, err := getCurrentWeather(args.Location, args.Unit)
				if err != nil {
					log.Fatal(err)
				}

				// Append tool_result to messageHistory
				weatherCallResponse := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							ToolCallID: toolCall.ID,
							Name:       toolCall.FunctionCall.Name,
							Content:    response,
						},
					},
				}
				messageHistory = append(messageHistory, weatherCallResponse)
			default:
				log.Fatalf("Unsupported tool: %s", toolCall.FunctionCall.Name)
			}
		}
	}

	return messageHistory
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherResponses := map[string]string{
		"boston":  "72 and sunny",
		"chicago": "65 and windy",
	}

	loweredLocation := strings.ToLower(location)

	var weatherInfo string
	found := false
	for key, value := range weatherResponses {
		if strings.Contains(loweredLocation, key) {
			weatherInfo = value
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("no weather info for %q", location)
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

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
						"description": "The city and state, e.g. San Francisco, CA",
					},
					"unit": map[string]any{
						"type": "string",
						"enum": []string{"fahrenheit", "celsius"},
					},
				},
				"required": []string{"location"},
			},
		},
	},
}
