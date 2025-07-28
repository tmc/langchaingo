package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/googleai"
)

func main() {
	genaiKey := os.Getenv("GOOGLE_API_KEY")
	if genaiKey == "" {
		log.Fatal("please set GOOGLE_API_KEY")
	}

	ctx := context.Background()

	llm, err := googleai.New(ctx, googleai.WithAPIKey(genaiKey))
	if err != nil {
		log.Fatal(err)
	}

	// Start by sending an initial question about the weather to the model, adding
	// "available tools" that include a getCurrentWeather function.
	// Thoroughout this sample, messageHistory collects the conversation history
	// with the model - this context is needed to ensure tool calling works
	// properly.
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago?"),
	}
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	// Translate the model's response into a MessageContent element that can be
	// added to messageHistory.
	respchoice := resp.Choices[0]
	assistantResponse := llms.TextParts(llms.ChatMessageTypeAI, respchoice.Content)
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	messageHistory = append(messageHistory, assistantResponse)

	// "Execute" tool calls by calling requested function
	for _, tc := range respchoice.ToolCalls {
		switch tc.FunctionCall.Name {
		case "getCurrentWeather":
			var args struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
				log.Fatal(err)
			}
			if strings.Contains(args.Location, "Chicago") {
				toolResponse := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							Name:    tc.FunctionCall.Name,
							Content: "64 and sunny",
						},
					},
				}
				messageHistory = append(messageHistory, toolResponse)
			}
		default:
			log.Fatalf("got unexpected function call: %v", tc.FunctionCall.Name)
		}
	}

	resp, err = llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response after tool call:")
	b, _ := json.MarshalIndent(resp.Choices[0], " ", "  ")
	fmt.Println(string(b))
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
				},
				"required": []string{"location"},
			},
		},
	},
}
