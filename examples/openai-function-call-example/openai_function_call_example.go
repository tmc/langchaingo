package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := openai.New(
		openai.WithModel("gpt-3.5-turbo-0125"),
		openai.WithHTTPClient(httputil.DebugHTTPClient),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	messageHistory := []llms.MessageContent{
		llms.TextParts(schema.ChatMessageTypeHuman, "What is the weather like in Boston and Chicago?"),
	}
	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			llms.TextParts(schema.ChatMessageTypeHuman, "What is the weather like in Boston?"),
		},
		llms.WithFunctions(functions))
	if err != nil {
		log.Fatal(err)
	}

	// add the assistant's response to the message history:
	assistantResponse := llms.MessageContent{
		Role: schema.ChatMessageTypeAI,
	}
	// if mainChoice.Content != "" {
	// 	assistantResponse.Parts = append(assistantResponse.Parts, llms.TextPart(mainChoice.Content))
	// }
	// if len(mainChoice.ToolCalls) > 0 {
	// 	for _, toolCall := range mainChoice.ToolCalls {
	// 		messageHistory = append(messageHistory, llms.ToolResponsePart(toolCall.ID, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments

	// 		// assistantResponse.Parts = append(
	// 		// 	assistantResponse.Parts,
	// 		// 	llms.ToolResponsePart(toolCall.ID, toolCall.FunctionCall.Name, toolCall.Response),
	// 		// )
	// 	}
	// }

	// messageHistory = append(messageHistory, llms.MessageContent{
	// 	Role:  schema.ChatMessageTypeAI,
	// 	Parts: resp.Choices[0].Parts,
	// })

	choice1 := resp.Choices[0]
	if len(choice1.ToolCalls) > 0 {
		fmt.Printf("Tool calls: ")
		json.NewEncoder(os.Stdout).Encode(choice1.ToolCalls)
		fmt.Println()
	}

	for _, toolCall := range choice1.ToolCalls {
		assistantResponse.Parts = append(assistantResponse.Parts, llms.ToolResponse{
			ToolCallID: toolCall.ID,
			Name:       toolCall.FunctionCall.Name,
			Arguments:  toolCall.FunctionCall.Arguments,
		})
	}

	// walk the tool calls and execute them:
	for _, toolCall := range choice1.ToolCalls {
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

			weather := llms.ToolResponsePart(toolCall.ID, toolCall.FunctionCall.Name, response)
			messageHistory = append(messageHistory, weather)
		}
	}

	messageHistory = append(messageHistory, llms.TextParts(schema.ChatMessageTypeHuman, "Can you compare the two?"))

	resp, err = llm.GenerateContent(ctx,
		messageHistory,
		llms.WithFunctions(functions))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Choices[0])
}

func getCurrentWeather(location string, unit string) (string, error) {
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

var weatherResponses = map[string]string{
	"boston":  "72 and sunny",
	"chicago": "65 and windy",
}

var functions = []llms.FunctionDefinition{
	{
		Name:        "getCurrentWeather",
		Description: "Get the current weather in a given location",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}}, "required": ["location"]}`),
	},
}
