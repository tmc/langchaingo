package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := openai.New(openai.WithModel("gpt-3.5-turbo-0125"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
<<<<<<< HEAD
	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			llms.TextParts(schema.ChatMessageTypeHuman, "What is the weather like in Boston?"),
		},
		llms.WithFunctions(functions))
=======
	messageHistory := []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeHuman, "What is the weather like in Boston and Chicago?"),
	}

	fmt.Println("Querying for weather in Boston and Chicago..")
	resp := queryLLM(ctx, llm, messageHistory, availableTools)
	fmt.Println("Initial response:", showResponse(resp))
	messageHistory = updateMessageHistory(messageHistory, resp)

	messageHistory = executeToolCalls(ctx, llm, messageHistory, resp)

	messageHistory = append(messageHistory, llms.TextParts(schema.ChatMessageTypeHuman, "Can you compare the two?"))
	fmt.Println("Querying with tool response...")
	resp = queryLLM(ctx, llm, messageHistory, availableTools)
	fmt.Println(resp.Choices[0].Content)
}

// queryLLM queries the LLM with the given message history and list of available
// tools, and returns the response.
func queryLLM(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, tools []llms.Tool) *llms.ContentResponse {
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(tools))
>>>>>>> upstream/main
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

// updateMessageHistory updates the message history with the assistant's
// response, and translates tool calls.
func updateMessageHistory(messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
	assistantResponse := llms.MessageContent{
		Role: schema.ChatMessageTypeAI,
	}
	for _, tc := range resp.Choices[0].ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, llms.ToolCall{
			ID:   tc.ID,
			Type: tc.Type,
			FunctionCall: &schema.FunctionCall{
				Name:      tc.FunctionCall.Name,
				Arguments: tc.FunctionCall.Arguments,
			},
		})
	}
	return append(messageHistory, assistantResponse)
}

// executeToolCalls executes the tool calls in the response and returns the
// updated message history.
func executeToolCalls(ctx context.Context, llm llms.Model, messageHistory []llms.MessageContent, resp *llms.ContentResponse) []llms.MessageContent {
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
				Role: schema.ChatMessageTypeTool,
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

	return messageHistory
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherResponses := map[string]string{
		"boston":  "72 and sunny",
		"chicago": "65 and windy",
	}

	weatherInfo, ok := weatherResponses[strings.ToLower(location)]
	if !ok {
		return "", fmt.Errorf("no weather info for %q", location)
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
