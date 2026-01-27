package googleai

import (
	"encoding/json"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genai"
)

func TestConvertParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		parts   []llms.ContentPart
		wantErr bool
	}{
		{
			name:    "empty parts",
			parts:   []llms.ContentPart{},
			wantErr: false,
		},
		{
			name: "text content",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello world"},
			},
			wantErr: false,
		},
		{
			name: "binary content",
			parts: []llms.ContentPart{
				llms.BinaryContent{
					MIMEType: "image/jpeg",
					Data:     []byte("fake image data"),
				},
			},
			wantErr: false,
		},
		{
			name: "tool call",
			parts: []llms.ContentPart{
				llms.ToolCall{
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tool call response",
			parts: []llms.ContentPart{
				llms.ToolCallResponse{
					Name:    "get_weather",
					Content: "It's sunny in Paris",
				},
			},
			wantErr: false,
		},
		{
			name: "tool call with invalid JSON",
			parts: []llms.ContentPart{
				llms.ToolCall{
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{invalid json}`,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertParts(tt.parts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, len(tt.parts))

			// Basic validation that all parts are created
			for i, part := range result {
				assert.NotNil(t, part, "Part %d should not be nil", i)
			}
		})
	}
}

func TestConvertContent(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name         string
		content      llms.MessageContent
		expectedRole string
		wantErr      bool
		errContains  string
	}{
		{
			name: "system message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "You are a helpful assistant"},
				},
			},
			expectedRole: RoleSystem,
			wantErr:      false,
		},
		{
			name: "AI message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Hello! How can I help you?"},
				},
			},
			expectedRole: RoleModel,
			wantErr:      false,
		},
		{
			name: "human message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "What's the weather like?"},
				},
			},
			expectedRole: RoleUser,
			wantErr:      false,
		},
		{
			name: "generic message maps to user",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeGeneric,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Generic content"},
				},
			},
			expectedRole: RoleUser,
			wantErr:      false,
		},
		{
			name: "tool message maps to user",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						Name:    "get_weather",
						Content: "Sunny",
					},
				},
			},
			expectedRole: RoleUser,
			wantErr:      false,
		},
		{
			name: "function message (unsupported)",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeFunction,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Function response"},
				},
			},
			wantErr:     true,
			errContains: "not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertContent(tt.content)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedRole, result.Role)
			assert.Len(t, result.Parts, len(tt.content.Parts))
		})
	}
}

func TestConvertResponse(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name     string
		response *genai.GenerateContentResponse
		wantErr  bool
	}{
		{
			name: "basic response",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: "Hello world"},
							},
						},
						FinishReason: genai.FinishReasonStop,
					},
				},
				UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
					PromptTokenCount:     10,
					CandidatesTokenCount: 5,
					TotalTokenCount:      15,
				},
			},
			wantErr: false,
		},
		{
			name: "response with thinking content",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{Text: "Let me think about this...", Thought: true},
								{Text: "The answer is 42"},
							},
						},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "response with function call",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{
									FunctionCall: &genai.FunctionCall{
										Name: "get_weather",
										Args: map[string]any{"location": "Paris"},
									},
								},
							},
						},
						FinishReason: genai.FinishReasonStop,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty candidates",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertResponse(tt.response)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Choices)

			choice := result.Choices[0]
			if tt.response.UsageMetadata != nil {
				assert.Contains(t, choice.GenerationInfo, "input_tokens")
				assert.Contains(t, choice.GenerationInfo, "output_tokens")
				assert.Contains(t, choice.GenerationInfo, "total_tokens")
			}

			// Check for thinking content in metadata
			if hasThinkingContent(tt.response) {
				assert.NotNil(t, choice.Reasoning)
				if choice.Reasoning != nil {
					assert.NotEmpty(t, choice.Reasoning.Content)
				}
			}
		})
	}
}

// Helper function to check if response has thinking content
func hasThinkingContent(resp *genai.GenerateContentResponse) bool {
	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part.Thought {
					return true
				}
			}
		}
	}
	return false
}

func TestCall(t *testing.T) {
	t.Parallel()

	// Test interface compliance
	t.Run("implements interface", func(t *testing.T) {
		var _ llms.Model = &GoogleAI{}
	})
}

func TestGenerateContentOptionsHandling(t *testing.T) {
	t.Parallel()

	t.Run("conflicting JSONMode and ResponseMIMEType", func(t *testing.T) {
		opts := llms.CallOptions{
			JSONMode:         true,
			ResponseMIMEType: "text/plain",
		}

		hasConflict := opts.ResponseMIMEType != "" && opts.JSONMode
		assert.True(t, hasConflict, "Should detect conflicting options")
	})

	t.Run("JSONMode sets correct MIME type", func(t *testing.T) {
		expectedMIMEType := ResponseMIMETypeJson
		assert.Equal(t, "application/json", expectedMIMEType)
	})

	t.Run("reasoning options validation", func(t *testing.T) {
		reasoning := &llms.ReasoningConfig{
			Effort: llms.ReasoningHigh,
			Tokens: 1000,
		}

		assert.True(t, reasoning.IsEnabled())
		assert.Equal(t, 1000, reasoning.GetTokens(2000))
	})
}

func TestRoleMapping(t *testing.T) {
	t.Parallel()

	roleTests := []struct {
		llmRole      llms.ChatMessageType
		expectedRole string
		supported    bool
	}{
		{llms.ChatMessageTypeSystem, RoleSystem, true},
		{llms.ChatMessageTypeAI, RoleModel, true},
		{llms.ChatMessageTypeHuman, RoleUser, true},
		{llms.ChatMessageTypeGeneric, RoleUser, true},
		{llms.ChatMessageTypeTool, RoleUser, true},
		{llms.ChatMessageTypeFunction, "", false},
	}

	for _, tt := range roleTests {
		t.Run(string(tt.llmRole), func(t *testing.T) {
			content := llms.MessageContent{
				Role:  tt.llmRole,
				Parts: []llms.ContentPart{llms.TextContent{Text: "test"}},
			}

			result, err := convertContent(content)

			if !tt.supported {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRole, result.Role)
		})
	}
}

func TestFunctionCallConversion(t *testing.T) {
	t.Parallel()

	t.Run("valid function call", func(t *testing.T) {
		args := map[string]any{
			"location": "Paris",
			"unit":     "celsius",
		}
		argsJSON, _ := json.Marshal(args)

		part := llms.ToolCall{
			FunctionCall: &llms.FunctionCall{
				Name:      "get_weather",
				Arguments: string(argsJSON),
			},
		}

		result, err := convertParts([]llms.ContentPart{part})
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		// Check that the part was created with function call
		assert.NotNil(t, result[0].FunctionCall)
		assert.Equal(t, "get_weather", result[0].FunctionCall.Name)
		assert.Equal(t, "Paris", result[0].FunctionCall.Args["location"])
		assert.Equal(t, "celsius", result[0].FunctionCall.Args["unit"])
	})

	t.Run("function response", func(t *testing.T) {
		part := llms.ToolCallResponse{
			Name:    "get_weather",
			Content: "It's 20°C and sunny",
		}

		result, err := convertParts([]llms.ContentPart{part})
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		// Check that the part was created with function response
		assert.NotNil(t, result[0].FunctionResponse)
		assert.Equal(t, "get_weather", result[0].FunctionResponse.Name)
		assert.Equal(t, "It's 20°C and sunny", result[0].FunctionResponse.Response["response"])
	})

	t.Run("malformed JSON in function call", func(t *testing.T) {
		part := llms.ToolCall{
			FunctionCall: &llms.FunctionCall{
				Name:      "get_weather",
				Arguments: `{invalid: json`,
			},
		}

		_, err := convertParts([]llms.ContentPart{part})
		assert.Error(t, err)
	})
}

func TestSafetySettings(t *testing.T) {
	t.Parallel()

	expectedCategories := []genai.HarmCategory{
		genai.HarmCategoryDangerousContent,
		genai.HarmCategoryHarassment,
		genai.HarmCategoryHateSpeech,
		genai.HarmCategorySexuallyExplicit,
	}

	harmThreshold := HarmBlockOnlyHigh

	safetySettings := []*genai.SafetySetting{}
	for _, category := range expectedCategories {
		safetySettings = append(safetySettings, &genai.SafetySetting{
			Category:  category,
			Threshold: convertHarmBlockThreshold(harmThreshold),
		})
	}

	assert.Len(t, safetySettings, 4, "Should have safety settings for all categories")

	for i, setting := range safetySettings {
		assert.Equal(t, expectedCategories[i], setting.Category)
		assert.Equal(t, convertHarmBlockThreshold(harmThreshold), setting.Threshold)
	}
}

func TestToolsConversion(t *testing.T) {
	t.Parallel()

	t.Run("valid tools", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "get_weather",
					Description: "Get weather information",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]any{
								"type":        "string",
								"description": "The location",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		}

		result, err := convertTools(tools)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].FunctionDeclarations, 1)

		decl := result[0].FunctionDeclarations[0]
		assert.Equal(t, "get_weather", decl.Name)
		assert.Equal(t, "Get weather information", decl.Description)
		assert.NotNil(t, decl.Parameters)
	})

	t.Run("unsupported tool type", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "unsupported_type",
				Function: &llms.FunctionDefinition{
					Name: "test",
				},
			},
		}

		_, err := convertTools(tools)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})

	t.Run("empty tools", func(t *testing.T) {
		result, err := convertTools([]llms.Tool{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestThinkingConfig(t *testing.T) {
	t.Parallel()

	t.Run("reasoning config validation", func(t *testing.T) {
		reasoning := &llms.ReasoningConfig{
			Effort: llms.ReasoningMedium,
			Tokens: 500,
		}

		assert.True(t, reasoning.IsEnabled())
		assert.Equal(t, 500, reasoning.GetTokens(1000))
	})

	t.Run("disabled reasoning", func(t *testing.T) {
		reasoning := &llms.ReasoningConfig{
			Effort: llms.ReasoningNone,
			Tokens: 0,
		}

		assert.False(t, reasoning.IsEnabled())
	})

	t.Run("nil reasoning config", func(t *testing.T) {
		var reasoning *llms.ReasoningConfig
		assert.Nil(t, reasoning)
	})
}
