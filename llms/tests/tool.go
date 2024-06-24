package tests

import (
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"strings"
)

func getCurrentWeather(location string, unit string) (string, error) {
	weatherResponses := map[string]string{
		"boston":   "29℃ and sunny",
		"chicago":  "15℃ and windy",
		"new york": "34℃ and sunny",
	}
	for k, v := range weatherResponses {
		if strings.Contains(strings.ToLower(location), k) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	return "", fmt.Errorf("no weather info for %q", location)
}

func getIpLocation(ip string) (string, error) {
	locationResponses := map[string]string{
		"8.8.8.0":   "United States, California, Mountain View",
		"127.0.0.1": "local host",
	}
	location, ok := locationResponses[ip]
	if !ok {
		return "", fmt.Errorf("no location info for %q", ip)
	}
	resp := map[string]string{
		"location": location,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// availableTools simulates the tools/functions we're making available for the model.
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
					"unit": map[string]any{
						"type": "string",
						"enum": []string{"fahrenheit", "celsius"},
					},
				},
				"required": []string{"location"},
			},
		},
	},
	{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getIpLocation",
			Description: "Get the location of an IP address",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ip": map[string]any{
						"type":        "string",
						"description": "The IP address to look up",
					},
				},
				"required": []string{"ip"},
			},
		},
	},
}

func callFunction(toolCall llms.ToolCall) (*llms.ToolCallResponse, error) {
	resp := &llms.ToolCallResponse{
		ToolCallID: toolCall.ID,
		Name:       toolCall.FunctionCall.Name,
	}
	args := make(map[string]any)
	if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
		resp.Content = "invalid arguments"
		return resp, err
	}
	var result string
	var err error
	switch toolCall.FunctionCall.Name {
	case "getCurrentWeather":
		location, ok := args["location"].(string)
		if !ok {
			resp.Content = "invalid location"
			return resp, fmt.Errorf("invalid location")
		}
		unit, ok := args["unit"].(string)
		if !ok {
			resp.Content = "invalid unit"
			return resp, fmt.Errorf("invalid unit")
		}
		result, err = getCurrentWeather(location, unit)
	case "getIpLocation":
		ip, ok := args["ip"].(string)
		if !ok {
			resp.Content = "invalid ip"
			return resp, fmt.Errorf("invalid ip")
		}
		result, err = getIpLocation(ip)
	default:
		return nil, fmt.Errorf("unsupported function: %s", toolCall.FunctionCall.Name)
	}
	resp.Content = result
	return resp, err
}
