package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms/ernie"

	"github.com/averikitsch/langchaingo/llms"
)

func main() {
	llm, err := ernie.New(
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
	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Boston?"),
		},
		llms.WithFunctions(functions))
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

var functions = []llms.FunctionDefinition{
	{
		Name:        "getCurrentWeather",
		Description: "Get the current weather in a given location",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}}, "required": ["location"]}`),
	},
}
