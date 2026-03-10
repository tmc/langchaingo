package googleai

import (
	"encoding/json"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

// TestSignatureDeduplication tests that empty TextContent with signature
// is skipped when ToolCall already has signature, preventing duplicate signatures.
//
// Context: When client code uses universal pattern for Anthropic compatibility:
//
//	aiParts := []llms.ContentPart{
//	    llms.TextPartWithReasoning(choice.Content, choice.Reasoning),
//	    toolCall1, // toolCall1 already has Reasoning with signature
//	}
//
// For Gemini:
// - choice.Reasoning is nil when tool calls present
// - choice.ToolCalls[0].Reasoning contains the signature
//
// Result: TextPartWithReasoning("", nil) + ToolCall(signature) works correctly
// But if client mistakenly adds TextPartWithReasoning("", signature), we should handle it.
func TestSignatureDeduplication(t *testing.T) { //nolint:funlen
	t.Parallel()

	testSig := []byte("gemini-thought-signature-xyz123")
	testReasoning := &reasoning.ContentReasoning{
		Content:   "Step 1: analyze the problem...",
		Signature: testSig,
	}

	t.Run("scenario_1_universal_pattern_correct", func(t *testing.T) {
		// Correct universal pattern: empty text (Gemini has no choice.Reasoning), tool call has signature
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "",  // Empty because Gemini puts reasoning in tool call
				Reasoning: nil, // nil for Gemini when tool calls present
			},
			llms.ToolCall{
				ID: "call_123",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "weather"}`,
				},
				Reasoning: testReasoning, // Gemini puts signature here
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// Empty text without reasoning is SKIPPED (prevents Gemini API errors)
		// Only tool call should remain
		assert.Len(t, result, 1, "Empty text without reasoning should be skipped")

		// Only tool call should have signature
		assert.Equal(t, testSig, result[0].ThoughtSignature, "Tool call should have signature")
		assert.NotNil(t, result[0].FunctionCall, "Remaining part should be function call")
	})

	t.Run("scenario_2_mistaken_duplicate_signature", func(t *testing.T) {
		// Mistaken pattern: client adds reasoning to BOTH text and tool call
		// This could happen if client code doesn't check provider-specific reasoning location
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "",            // Empty content
				Reasoning: testReasoning, // Mistakenly added (should be nil for Gemini with tools)
			},
			llms.ToolCall{
				ID: "call_123",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "weather"}`,
				},
				Reasoning: testReasoning, // Correct - Gemini puts signature here
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// OPTIMIZATION: Empty text with signature is SKIPPED when tool call has signature
		assert.Len(t, result, 1, "Empty text part should be skipped to prevent duplicate signature")

		// Only tool call should remain with signature
		assert.Equal(t, testSig, result[0].ThoughtSignature, "Only tool call signature should remain")
		assert.NotNil(t, result[0].FunctionCall, "Remaining part should be function call")
	})

	t.Run("scenario_3_non_empty_text_with_signature", func(t *testing.T) {
		// Edge case: non-empty text with signature + tool call with signature
		// This could happen with Anthropic pattern applied to Gemini
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "I will search for the weather", // Non-empty
				Reasoning: testReasoning,
			},
			llms.ToolCall{
				ID: "call_123",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "weather"}`,
				},
				Reasoning: testReasoning,
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// Non-empty text is KEPT even with tool call signature
		// (text content may be important)
		assert.Len(t, result, 2, "Non-empty text part should be kept")

		// Both parts should have signature
		assert.Equal(t, testSig, result[0].ThoughtSignature, "Text part should have signature")
		assert.Equal(t, testSig, result[1].ThoughtSignature, "Tool call should have signature")

		t.Log("WARNING: This creates duplicate signatures - may not be optimal for Gemini")
	})

	t.Run("scenario_4_empty_text_no_tool_calls", func(t *testing.T) {
		// Edge case: empty text with signature, no tool calls
		// Signature must be preserved (no tool call to carry it)
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "", // Empty
				Reasoning: testReasoning,
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// Empty text is KEPT when there are no tool calls (signature would be lost otherwise)
		assert.Len(t, result, 1, "Empty text part must be kept when no tool calls")
		assert.Equal(t, testSig, result[0].ThoughtSignature, "Signature must be preserved")
	})

	t.Run("scenario_5_empty_text_with_signature_tool_without", func(t *testing.T) {
		// Edge case: empty text with signature + tool call WITHOUT signature
		// Both must be kept (tool call can't carry the signature)
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "", // Empty
				Reasoning: testReasoning,
			},
			llms.ToolCall{
				ID: "call_123",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"query": "weather"}`,
				},
				Reasoning: nil, // No signature in tool call
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// Empty text is KEPT because tool call doesn't have signature
		assert.Len(t, result, 2, "Empty text part must be kept when tool call lacks signature")
		assert.Equal(t, testSig, result[0].ThoughtSignature, "Text part should have signature")
		assert.Nil(t, result[1].ThoughtSignature, "Tool call should NOT have signature")
	})
}

// TestSignaturePlacementInRequest verifies the actual genai.Part structure
// that would be sent to Gemini API in various scenarios.
func TestSignaturePlacementInRequest(t *testing.T) {
	t.Parallel()

	signature := []byte("test-sig")
	reasoning := &reasoning.ContentReasoning{
		Content:   "thinking...",
		Signature: signature,
	}

	t.Run("gemini_tool_call_pattern", func(t *testing.T) {
		// Typical Gemini response pattern:
		// - choice.Reasoning is nil
		// - choice.ToolCalls[0].Reasoning has signature
		//
		// Client should add: ToolCall part (signature embedded)
		parts := []llms.ContentPart{
			llms.ToolCall{
				ID: "call_1",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"q": "test"}`,
				},
				Reasoning: reasoning,
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)
		require.Len(t, result, 1)

		// Verify structure
		assert.NotNil(t, result[0].FunctionCall, "Should be function call")
		assert.Equal(t, signature, result[0].ThoughtSignature, "Should have signature")
		assert.Equal(t, "search", result[0].FunctionCall.Name)
	})

	t.Run("anthropic_pattern_on_gemini", func(t *testing.T) {
		// Anthropic pattern applied to Gemini (suboptimal but may happen):
		// - Client adds: TextPartWithReasoning(choice.Content, choice.Reasoning) + ToolCall
		// - For Anthropic: choice.Reasoning has signature
		// - For Gemini: choice.Reasoning is nil, ToolCall.Reasoning has signature
		//
		// After extraction from Gemini response, client builds:
		// - TextPartWithReasoning("", toolCall.Reasoning) <- empty text with signature
		// - ToolCall <- already has signature
		parts := []llms.ContentPart{
			llms.TextContent{
				Text:      "", // Empty (Gemini doesn't put content in choice when tool calls)
				Reasoning: reasoning,
			},
			llms.ToolCall{
				ID: "call_1",
				FunctionCall: &llms.FunctionCall{
					Name:      "search",
					Arguments: `{"q": "test"}`,
				},
				Reasoning: reasoning, // Same signature
			},
		}

		result, err := convertParts(parts)
		require.NoError(t, err)

		// OPTIMIZATION: Empty text should be skipped
		require.Len(t, result, 1, "Empty text part should be skipped")
		assert.NotNil(t, result[0].FunctionCall, "Remaining part should be function call")
		assert.Equal(t, signature, result[0].ThoughtSignature, "Tool call should have signature")
	})
}

func TestConvertParts_DuplicateSignatures(t *testing.T) { //nolint:funlen
	t.Parallel()

	testSignature := []byte("test-signature-bytes")
	testReasoning := &reasoning.ContentReasoning{
		Content:   "Test reasoning content",
		Signature: testSignature,
	}

	tests := []struct {
		name              string
		parts             []llms.ContentPart
		expectedPartCount int
		expectedSigCount  int // Number of parts with non-nil ThoughtSignature
		description       string
	}{
		{
			name: "non-empty text with signature + tool call with signature",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "I will call the tool", // Non-empty text
					Reasoning: testReasoning,
				},
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: testReasoning,
				},
			},
			expectedPartCount: 2,
			expectedSigCount:  2, // Both parts have signature - acceptable for non-empty text
			description:       "Non-empty text and tool call both have signatures - kept because text has content",
		},
		{
			name: "empty text with signature + tool call with signature",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "", // Empty text but has reasoning
					Reasoning: testReasoning,
				},
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: testReasoning,
				},
			},
			expectedPartCount: 1, // Empty text part is SKIPPED (optimization)
			expectedSigCount:  1, // Only tool call gets signature
			description:       "Empty text part with signature is skipped when tool call has signature - prevents duplicate",
		},
		{
			name: "text with signature + tool call without signature",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "I will call the tool",
					Reasoning: testReasoning,
				},
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: nil, // No reasoning in tool call
				},
			},
			expectedPartCount: 2,
			expectedSigCount:  1, // Only text part has signature
			description:       "Text has signature, tool call doesn't",
		},
		{
			name: "text without signature + tool call with signature",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "I will call the tool",
					Reasoning: nil,
				},
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: testReasoning,
				},
			},
			expectedPartCount: 2,
			expectedSigCount:  1, // Only tool call has signature
			description:       "Tool call has signature, text doesn't - this is typical for Gemini",
		},
		{
			name: "empty text with signature + NO tool calls",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "", // Empty text but has reasoning
					Reasoning: testReasoning,
				},
			},
			expectedPartCount: 1, // Part is KEPT (no tool calls to carry signature)
			expectedSigCount:  1, // Text part keeps signature
			description:       "Empty text with signature is preserved when no tool calls - prevents losing signature",
		},
		{
			name: "multiple tool calls - only first has signature",
			parts: []llms.ContentPart{
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: testReasoning, // First has signature
				},
				llms.ToolCall{
					ID: "call_2",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "London"}`,
					},
					Reasoning: nil, // Second doesn't have signature
				},
			},
			expectedPartCount: 2,
			expectedSigCount:  1, // Only first tool call has signature
			description:       "Parallel tool calls - signature only in first (correct pattern)",
		},
		{
			name: "empty text with signature + tool call WITHOUT signature",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text:      "", // Empty text but has reasoning
					Reasoning: testReasoning,
				},
				llms.ToolCall{
					ID: "call_1",
					FunctionCall: &llms.FunctionCall{
						Name:      "get_weather",
						Arguments: `{"location": "Paris"}`,
					},
					Reasoning: nil, // Tool call has NO signature
				},
			},
			expectedPartCount: 2, // Both parts kept (tool call has no signature to carry)
			expectedSigCount:  1, // Only text part has signature
			description:       "Empty text with signature preserved when tool call lacks signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := convertParts(tt.parts)
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedPartCount, tt.description)

			// Count parts with ThoughtSignature
			sigCount := 0
			for i, part := range result {
				if len(part.ThoughtSignature) > 0 {
					sigCount++
					t.Logf("Part %d has signature: %d bytes", i, len(part.ThoughtSignature))

					// Verify signature matches expected
					assert.Equal(t, testSignature, part.ThoughtSignature,
						"Part %d signature should match test signature", i)
				}
			}

			assert.Equal(t, tt.expectedSigCount, sigCount,
				"Expected %d parts with signature, got %d. %s",
				tt.expectedSigCount, sigCount, tt.description)

			// Log what would be sent to API
			t.Logf("Would send to Gemini API: %d parts, %d with signatures", len(result), sigCount)
		})
	}
}

func TestConvertParts(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name           string
		parts          []llms.ContentPart
		wantErr        bool
		expectedLength int // -1 means same as input
	}{
		{
			name:           "empty parts",
			parts:          []llms.ContentPart{},
			wantErr:        false,
			expectedLength: 0,
		},
		{
			name: "text content",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello world"},
			},
			wantErr:        false,
			expectedLength: 1,
		},
		{
			name: "empty text without reasoning - should be skipped",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "", Reasoning: nil},
			},
			wantErr:        false,
			expectedLength: 0, // Empty text without reasoning is skipped
		},
		{
			name: "empty text with reasoning - should be kept",
			parts: []llms.ContentPart{
				llms.TextContent{
					Text: "",
					Reasoning: &reasoning.ContentReasoning{
						Content:   "thinking",
						Signature: []byte("sig"),
					},
				},
			},
			wantErr:        false,
			expectedLength: 1, // Empty text with reasoning is kept
		},
		{
			name: "mixed: empty text + tool call",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "", Reasoning: nil}, // Will be skipped
				llms.ToolCall{
					FunctionCall: &llms.FunctionCall{
						Name:      "search",
						Arguments: `{"q":"test"}`,
					},
				},
			},
			wantErr:        false,
			expectedLength: 1, // Only tool call remains
		},
		{
			name: "binary content",
			parts: []llms.ContentPart{
				llms.BinaryContent{
					MIMEType: "image/jpeg",
					Data:     []byte("fake image data"),
				},
			},
			wantErr:        false,
			expectedLength: 1,
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
			wantErr:        false,
			expectedLength: 1,
		},
		{
			name: "tool call response",
			parts: []llms.ContentPart{
				llms.ToolCallResponse{
					Name:    "get_weather",
					Content: "It's sunny in Paris",
				},
			},
			wantErr:        false,
			expectedLength: 1,
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

			expectedLen := tt.expectedLength
			if expectedLen == -1 {
				expectedLen = len(tt.parts)
			}
			assert.Len(t, result, expectedLen, "Expected %d parts, got %d", expectedLen, len(result))

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
			ResponseMIMEType: getStringPointer("text/plain"),
		}

		hasConflict := opts.ResponseMIMEType != nil && opts.JSONMode
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

	safetySettings := make([]*genai.SafetySetting, 0, len(expectedCategories))
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
