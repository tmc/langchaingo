package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/ollama"
)

var flagVerbose = flag.Bool("v", false, "verbose mode")

func main() {
	flag.Parse()
	// allow specifying your own model via OLLAMA_TEST_MODEL
	// (same as the Ollama unit tests).
	model := "llama3"
	if v := os.Getenv("OLLAMA_TEST_MODEL"); v != "" {
		model = v
	}

	llm, err := ollama.New(
		ollama.WithModel(model),
		ollama.WithFormat("json"),
	)
	if err != nil {
		log.Fatal(err)
	}

	var msgs []llms.MessageContent

	// system message defines the available tools.
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, systemMessage()))
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in Beijing?"))

	ctx := context.Background()

	for retries := 3; retries > 0; retries = retries - 1 {
		resp, err := llm.GenerateContent(ctx, msgs)
		if err != nil {
			log.Fatal(err)
		}

		choice1 := resp.Choices[0]
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, choice1.Content))

		if c := unmarshalCall(choice1.Content); c != nil {
			log.Printf("Call: %v", c.Tool)
			if *flagVerbose {
				log.Printf("Call: %v (raw: %v)", c.Tool, choice1.Content)
			}
			msg, cont := dispatchCall(c)
			if !cont {
				break
			}
			msgs = append(msgs, msg)
		} else {
			// Ollama doesn't always respond with a function call, let it try again.
			log.Printf("Not a call: %v", choice1.Content)
			msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, "Sorry, I don't understand. Please try again."))
		}

		if retries == 0 {
			log.Fatal("retries exhausted")
		}
	}
}

type Call struct {
	Tool  string         `json:"tool"`
	Input map[string]any `json:"tool_input"`
}

func unmarshalCall(input string) *Call {
	var c Call
	if err := json.Unmarshal([]byte(input), &c); err == nil && c.Tool != "" {
		return &c
	}
	return nil
}

func dispatchCall(c *Call) (llms.MessageContent, bool) {
	// ollama doesn't always respond with a *valid* function call. As we're using prompt
	// engineering to inject the tools, it may hallucinate.
	if !validTool(c.Tool) {
		log.Printf("invalid function call: %#v, prompting model to try again", c)
		return llms.TextParts(llms.ChatMessageTypeHuman,
			"Tool does not exist, please try again."), true
	}

	// we could make this more dynamic, by parsing the function schema.
	switch c.Tool {
	case "getCurrentWeather":
		loc, ok := c.Input["location"].(string)
		if !ok {
			log.Fatal("invalid input")
		}
		unit, ok := c.Input["unit"].(string)
		if !ok {
			log.Fatal("invalid input")
		}

		weather, err := getCurrentWeather(loc, unit)
		if err != nil {
			log.Fatal(err)
		}
		return llms.TextParts(llms.ChatMessageTypeHuman, weather), true
	case "finalResponse":
		resp, ok := c.Input["response"].(string)
		if !ok {
			log.Fatal("invalid input")
		}

		log.Printf("Final response: %v", resp)

		return llms.MessageContent{}, false
	default:
		// we already checked above if we had a valid tool.
		panic("unreachable")
	}
}

func validTool(name string) bool {
	var valid []string
	for _, v := range functions {
		valid = append(valid, v.Name)
	}
	return slices.Contains(valid, name)
}

func systemMessage() string {
	bs, err := json.Marshal(functions)
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf(`You have access to the following tools:

%s

To use a tool, respond with a JSON object with the following structure: 
{
	"tool": <name of the called tool>,
	"tool_input": <parameters for the tool matching the above JSON schema>
}
`, string(bs))
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

var functions = []llms.FunctionDefinition{
	{
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
	{
		// I found that providing a tool for Ollama to give the final response significantly
		// increases the chances of success.
		Name:        "finalResponse",
		Description: "Provide the final response to the user query",
		Parameters: json.RawMessage(`{
			"type": "object", 
			"properties": {
				"response": {"type": "string", "description": "The final response to the user query"}
			}, 
			"required": ["response"]
		}`),
	},
}
