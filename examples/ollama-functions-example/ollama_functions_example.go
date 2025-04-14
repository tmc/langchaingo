package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"slices"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

func main() {
	flag.Parse()
	// allow specifying your own model via OLLAMA_TEST_MODEL
	// (same as the Ollama unit tests).
	model := "llama3.2"
	if v := os.Getenv("OLLAMA_TEST_MODEL"); v != "" {
		model = v
	}

	llm, err := ollama.New(
		ollama.WithModel(model),
	)
	if err != nil {
		log.Fatal(err)
	}

	var msgs []llms.MessageContent

	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in Beijing?"))

	// options defines the available tools.
	toolOpt := llms.WithTools(tools)

	ctx := context.Background()

	for retries := 3; retries > 0; retries = retries - 1 {
		resp, err := llm.GenerateContent(ctx, msgs, toolOpt)
		if err != nil {
			log.Fatal(err)
		}

		choice1 := resp.Choices[0]

		for _, tc := range choice1.ToolCalls {
			log.Printf("Call: %v", tc.FunctionCall)

			msgs = append(msgs, llms.MessageContent{
				Role:  llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{tc},
			})

			msg, cont := dispatchCall(tc)
			if !cont {
				break
			}
			msgs = append(msgs, msg)
		}

		if len(choice1.ToolCalls) == 0 {
			log.Printf(choice1.Content)
			break
		}
	}
}

func dispatchCall(c llms.ToolCall) (llms.MessageContent, bool) {
	// ollama doesn't always respond with a *valid* function call. As we're using prompt
	// engineering to inject the tools, it may hallucinate.
	if !validTool(c.FunctionCall.Name) {
		log.Printf("invalid function call: %#v, prompting model to try again", c)
		return llms.TextParts(llms.ChatMessageTypeHuman,
			"Tool does not exist, please try again."), true
	}

	var parsedArguments map[string]any
	err := json.Unmarshal([]byte(c.FunctionCall.Arguments), &parsedArguments)
	if err != nil {
		log.Fatal(err)
	}

	// we could make this more dynamic, by parsing the function schema.
	switch c.FunctionCall.Name {
	case "getCurrentWeather":
		loc, ok := parsedArguments["location"].(string)
		if !ok {
			log.Fatal("invalid input")
		}
		unit, ok := parsedArguments["unit"].(string)
		if !ok {
			log.Fatal("invalid input")
		}

		weather, err := getCurrentWeather(loc, unit)
		if err != nil {
			log.Fatal(err)
		}
		return llms.TextParts(llms.ChatMessageTypeTool, weather), true
	default:
		// we already checked above if we had a valid tool.
		panic("unreachable")
	}
}

func validTool(name string) bool {
	var valid []string
	for _, v := range tools {
		valid = append(valid, v.Function.Name)
	}
	return slices.Contains(valid, name)
}

func getCurrentWeather(location string, unit string) (string, error) {
	weatherInfo := map[string]any{
		"location":    location,
		"temperature": "6",
		"unit":        unit,
		"forecast":    []string{"sunny", "windy"},
	}
	if unit == "fahrenheit" {
		weatherInfo["temperature"] = 43
	}

	b, err := json.Marshal(weatherInfo)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var tools = []llms.Tool{
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: json.RawMessage(`{
				"type": "object", 
				"properties": {
					"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, 
					"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
				}, 
				"required": ["location", "unit"]
			}`),
		},
	},
}
