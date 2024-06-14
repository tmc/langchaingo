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
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Boston and Chicago?"),
	}

	fmt.Println("Querying for weather in Boston and Chicago..")
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}
	messageHistory = updateMessageHistory(messageHistory, resp)

	// Execute tool calls requested by the model
	messageHistory = executeToolCalls(ctx, llm, messageHistory, resp)
	messageHistory = append(messageHistory, llms.TextParts(llms.ChatMessageTypeHuman, "Can you compare the two?"))

	// Send query to the model again, this time with a history containing its
	// request to invoke a tool and our response to the tool call.
	fmt.Println("Querying with tool response...")
	resp, err = llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Choices[0].Content)
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
func executeToolCalls(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	for _, choice := range resp.Choices {
		for _, toolCall := range choice.ToolCalls {
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
					fmt.Printf("error: %v\n", err)
					log.Fatal(err)
				}

				fmt.Printf("Weather of %s is %s\n", args.Location, response)

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

func showResponse(resp *llms.ContentResponse) string {
	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}
