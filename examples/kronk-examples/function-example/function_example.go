// tool_call_example demonstrates tool calling using the kronk abstraction
// layer. It uses prompt engineering to inject tool definitions into the
// system message and parses the model's JSON response to determine which
// tool to call — the same approach as the ollama-functions-example, but
// running a local GGUF model via llama.cpp through kronk.
//
// The first time you run this program the system will download and install
// the llama.cpp libraries and the model. Subsequent runs load from disk.
//
// Run from this directory:
//
//	go run tool_call_example.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"github.com/tmc/langchaingo/examples/kronk-examples/kronk"
	"github.com/tmc/langchaingo/llms"
)

var flagVerbose = flag.Bool("v", false, "verbose mode")

func main() {
	flag.Parse()

	modelSource := "unsloth/Qwen3-0.6B-Q8_0"
	if v := os.Getenv("KRONK_TEST_MODEL"); v != "" {
		modelSource = v
	}

	fmt.Println("Initializing kronk (first run downloads libraries and model)...")
	client, err := kronk.New(
		context.Background(),
		modelSource,
		kronk.WithAutoTune(true),
	)
	if err != nil {
		log.Fatalf("create kronk client: %v", err)
	}

	defer func() {
		fmt.Println("\nUnloading Kronk")
		if err := client.Unload(context.Background()); err != nil {
			fmt.Printf("failed to unload model: %v", err)
		}
	}()

	var msgs []llms.MessageContent

	// system message defines the available tools.
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeSystem, systemMessage()))
	msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather like in Beijing?"))

	for retries := 3; retries > 0; retries-- {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)

		resp, err := client.GenerateContent(ctx, msgs)
		cancel()
		if err != nil {
			log.Fatalf("generate content: %v", err)
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
			// The model didn't respond with a function call, let it try again.
			log.Printf("Not a call: %v", choice1.Content)
			msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeHuman, "Sorry, I don't understand. Please try again."))
		}

	}
}

// Call is the JSON structure we expect the model to respond with when it
// wants to invoke a tool.
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
	if !validTool(c.Tool) {
		log.Printf("invalid function call: %#v, prompting model to try again", c)
		return llms.TextParts(llms.ChatMessageTypeHuman,
			"Tool does not exist, please try again."), true
	}

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
