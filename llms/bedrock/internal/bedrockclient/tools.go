package bedrockclient

import (
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// BedrockTool represents a tool definition for Bedrock API
type BedrockTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema"`
}

// BedrockToolChoice represents tool choice for Bedrock API
type BedrockToolChoice struct {
	Type string `json:"type,omitempty"` // "auto", "any", "tool"
	Name string `json:"name,omitempty"` // specific tool name when type is "tool"
}

// BedrockToolCall represents a tool call in Bedrock response
type BedrockToolCall struct {
	Type  string      `json:"type"` // "tool_use"
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

// convertToolsToBedrockTools converts llms.Tool to BedrockTool format
func convertToolsToBedrockTools(tools []llms.Tool) ([]BedrockTool, error) {
	bedrockTools := make([]BedrockTool, len(tools))

	for i, tool := range tools {
		// Convert function definition to Bedrock format
		if tool.Type != "function" {
			return nil, fmt.Errorf("only function tools are supported, got: %s", tool.Type)
		}

		bedrockTools[i] = BedrockTool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters, // Bedrock uses input_schema instead of parameters
		}
	}

	return bedrockTools, nil
}

// convertToolChoiceToBedrockToolChoice converts llms tool choice to Bedrock format
func convertToolChoiceToBedrockToolChoice(toolChoice interface{}) (*BedrockToolChoice, error) {
	if toolChoice == nil {
		return nil, nil
	}

	switch choice := toolChoice.(type) {
	case string:
		switch choice {
		case "auto":
			return &BedrockToolChoice{Type: "auto"}, nil
		case "none":
			return nil, nil // Bedrock doesn't have explicit "none", just omit tools
		case "required":
			return &BedrockToolChoice{Type: "any"}, nil
		default:
			return nil, fmt.Errorf("unsupported tool choice string: %s", choice)
		}
	case map[string]interface{}:
		// Handle structured tool choice like {"type": "tool", "function": {"name": "get_weather"}}
		if typeVal, ok := choice["type"].(string); ok && typeVal == "function" {
			if function, ok := choice["function"].(map[string]interface{}); ok {
				if name, ok := function["name"].(string); ok {
					return &BedrockToolChoice{Type: "tool", Name: name}, nil
				}
			}
		}
		return nil, fmt.Errorf("unsupported tool choice structure")
	default:
		return nil, fmt.Errorf("unsupported tool choice type: %T", toolChoice)
	}
}

// convertBedrockToolCallToLLMToolCall converts Bedrock tool call to llms.ToolCall
func convertBedrockToolCallToLLMToolCall(bedrockCall BedrockToolCall) (llms.ToolCall, error) {
	// Convert input to JSON string for Arguments field
	inputJSON, err := json.Marshal(bedrockCall.Input)
	if err != nil {
		return llms.ToolCall{}, fmt.Errorf("failed to marshal tool input: %w", err)
	}

	return llms.ToolCall{
		ID:   bedrockCall.ID,
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      bedrockCall.Name,
			Arguments: string(inputJSON),
		},
	}, nil
}
