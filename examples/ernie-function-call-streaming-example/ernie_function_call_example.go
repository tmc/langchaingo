package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms/ernie"
	"log"

	"github.com/tmc/langchaingo/jsonschema"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := ernie.NewChat(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		// Fill in your AK and SK here.
		ernie.WithAKSK("ak", "sk"),
		// Use an external cache for the access token.
		ernie.WithAccessToken("accesstoken"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Content: "What is the weather going to be like in Boston?"},
	}, llms.WithFunctions(functions), llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Printf("Received chunk: %s\n", chunk)
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}

	if completion.FunctionCall != nil {
		fmt.Printf("Function call: %+v\n", completion.FunctionCall)
	}
	fmt.Println(completion.Content)
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

var functions = []llms.FunctionDefinition{
	{
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
	{
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
	{
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
}
