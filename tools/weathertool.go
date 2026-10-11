package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
)

// WeatherTool is a tool that reports the weather for a city. It demonstrates a
// tool declaring its own two-parameter JSON Schema via ToolWithParameters
// instead of relying on the default "__arg1" single-string template.
type WeatherTool struct {
	CallbacksHandler callbacks.Handler
}

var _ Tool = WeatherTool{}

var _ ToolWithParameters = WeatherTool{}
var weatherDB = map[string]float64{
	"Seattle":  20,
	"Beijing":  28,
	"London":   15,
	"Tokyo":    30,
	"New York": 25,
}

// Description returns a string describing the weather tool.
func (w WeatherTool) Description() string {
	return `get the current weather for a city.`
}

// Name returns the name of the tool.
func (w WeatherTool) Name() string {
	return "get_weather"
}

// Parameters returns the two-parameter JSON Schema declared by the tool. The
// agent passes this schema through verbatim to the LLM provider for function
// calling.
func (w WeatherTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"city": map[string]any{
				"type":        "string",
				"description": "The name of the city to get weather for",
			},
			"unit": map[string]any{
				"type":        "string",
				"description": "The temperature unit, either celsius or fahrenheit",
			},
		},
		"required": []string{"city", "unit"},
	}
}

// weatherArgs is the shape of the arguments JSON the LLM returns for the tool.
type weatherArgs struct {
	City string `json:"city"`
	Unit string `json:"unit"`
}

// Call returns the given city and unit when the input is a valid two-parameter
// JSON object. It returns an error otherwise.
func (w WeatherTool) Call(ctx context.Context, input string) (string, error) {
	if w.CallbacksHandler != nil {
		w.CallbacksHandler.HandleToolStart(ctx, input)
	}

	var args weatherArgs
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("weather tool: expected JSON object with city and unit: %w", err)
	}
	if args.City == "" || args.Unit == "" {
		return "", fmt.Errorf("weather tool: city and unit are both required, got %q", input)
	}

	tempC, ok := weatherDB[args.City]
	if !ok {
		return "", fmt.Errorf("weather tool: no weather data for %q", args.City)
	}

	return fmt.Sprintf("The weather in %s is %.1f degrees %s.", args.City, tempC, args.Unit), nil
}
