// Set the VERTEX_PROJECT env var to your GCP project with Vertex AI APIs
// enabled. Set VERTEX_LOCATION to a GCP location (region); if you're not sure
// about the location, set us-central1
// Set the VERTEX_CREDENTIALS env var to the path of your GCP service account
// credentials JSON file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
)

func main() {
	ctx := context.Background()
	project := os.Getenv("VERTEX_PROJECT")
	location := os.Getenv("VERTEX_LOCATION")
	credentialsJSONFile := os.Getenv("VERTEX_CREDENTIALS")
	llm, err := vertex.New(
		ctx,
		googleai.WithCloudProject(project),
		googleai.WithCloudLocation(location),
		googleai.WithCredentialsFile(credentialsJSONFile),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start by sending an initial question about the weather to the model, adding
	// "available tools" that include a getCurrentWeather function.
	// Thoroughout this sample, messageHistory collects the conversation history
	// with the model - this context is needed to ensure tool calling works
	// properly.
	messageHistory := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago? And what's the elevation in Chicago?"),
	}
	resp, err := llm.GenerateContent(ctx, messageHistory, llms.WithTools(availableTools))
	if err != nil {
		log.Fatal(err)
	}

	// Translate the model's response into a MessageContent element that can be
	// added to messageHistory.
	respchoice := resp.Choices[0]
	assistantResponse := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: make([]llms.ContentPart, 0),
	}
	// If you set a text message part with empty content vertex will throw an error
	if len(respchoice.Content) > 0 {
		assistantResponse.Parts = append(assistantResponse.Parts, llms.TextPart(respchoice.Content))
	}
	for _, tc := range respchoice.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, tc)
	}
	messageHistory = append(messageHistory, assistantResponse)

	// Create the tool response here because the number of parts in the toolResponse must match len(respchoice.ToolCalls)
	toolResponse := llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: make([]llms.ContentPart, 0),
	}

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
			log.Printf("getting current weather for %s, location: %s", tc.FunctionCall.Name, args.Location)
			if strings.Contains(args.Location, "Chicago") {
				toolResponse.Parts = append(toolResponse.Parts, llms.ToolCallResponse{
					ToolCallID: tc.ID,
					Name:       tc.FunctionCall.Name,
					Content:    "64 and sunny",
				})
			}
		case "getElevation":
			var args struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
				log.Fatal(err)
			}
			log.Printf("getting elevation for %s, location: %s", tc.FunctionCall.Name, args.Location)
			if strings.Contains(args.Location, "Chicago") {
				toolResponse.Parts = append(toolResponse.Parts, llms.ToolCallResponse{
					ToolCallID: tc.ID,
					Name:       tc.FunctionCall.Name,
					Content:    "597.18 ft",
				})
			}
		default:
			log.Fatalf("got unexpected function call: %v", tc.FunctionCall.Name)
		}
	}
	if len(toolResponse.Parts) > 0 {
		messageHistory = append(messageHistory, toolResponse)
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
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getElevation",
			Description: "Get the elevation in a given location",
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
