package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/jsonschema"
	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New(openai.WithModel("gpt-4-turbo"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Boston?"),
		},
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Printf("Received chunk: %s\n", chunk)
			return nil
		}),
		llms.WithTools(tools))
	if err != nil {
		log.Fatal(err)
	}

	choice1 := resp.Choices[0]
	if choice1.FuncCall != nil {
		fmt.Printf("Function call: %v\n", choice1.FuncCall)
	}
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherInfo := map[string]interface{}{
		"location":    location,
		"temperature": "72",
		"unit":        unit,
		"forecast":    []string{"sunny", "windy"},
	}
	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// json.RawMessage(`{"type": "object", "properties": {"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}}, "required": ["location"]}`),

var tools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"rationale": {
						Type:        jsonschema.String,
						Description: "The rationale for choosing this function call with these parameters",
					},
					"location": {
						Type:        jsonschema.String,
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: jsonschema.String,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},
				Required: []string{"rationale", "location"},
			},
		},
	},
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getTomorrowWeather",
			Description: "Get the predicted weather in a given location",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"rationale": {
						Type:        jsonschema.String,
						Description: "The rationale for choosing this function call with these parameters",
					},
					"location": {
						Type:        jsonschema.String,
						Description: "The city and state, e.g. San Francisco, CA",
					},
					"unit": {
						Type: jsonschema.String,
						Enum: []string{"celsius", "fahrenheit"},
					},
				},
				Required: []string{"rationale", "location"},
			},
		},
	},
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getSuggestedPrompts",
			Description: "Given the user's input prompt suggest some related prompts",
			Parameters: jsonschema.Definition{
				Type: jsonschema.Object,
				Properties: map[string]jsonschema.Definition{
					"rationale": {
						Type:        jsonschema.String,
						Description: "The rationale for choosing this function call with these parameters",
					},
					"suggestions": {
						Type: jsonschema.Array,
						Items: &jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "A suggested prompt",
						},
					},
				},
				Required: []string{"rationale", "suggestions"},
			},
		},
	},
}
