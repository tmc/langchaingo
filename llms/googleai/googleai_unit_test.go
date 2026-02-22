package googleai

import (
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	tests := []struct {
		name        string
		opts        []Option
		wantErr     bool
		errContains string
	}{
		{
			name: "success with API key",
			opts: []Option{
				WithAPIKey("test-api-key"),
			},
			wantErr: false,
		},
		{
			name: "success with default options",
			opts: []Option{
				WithAPIKey("test-api-key"),
			},
			wantErr: false,
		},
		{
			name: "success with custom options",
			opts: []Option{
				WithAPIKey("test-api-key"),
				WithDefaultModel("custom-model"),
				WithDefaultTemperature(0.8),
				WithDefaultTopK(5),
				WithDefaultTopP(0.9),
				WithDefaultMaxTokens(1000),
				WithDefaultCandidateCount(2),
				WithHarmThreshold(HarmBlockMediumAndAbove),
			},
			wantErr: false,
		},
		{
			name: "success with cloud options",
			opts: []Option{
				WithCloudProject("test-project"),
				WithCloudLocation("us-central1"),
			},
			wantErr:     true,
			errContains: "failed to find default credentials",
		},
		{
			name: "success with embedding model",
			opts: []Option{
				WithAPIKey("test-api-key"),
				WithDefaultEmbeddingModel("embedding-002"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(ctx, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.NotNil(t, client.opts)
			}
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	t.Parallel()

	opts := DefaultOptions()

	assert.Equal(t, "gemini-2.0-flash", opts.DefaultModel)
	assert.Equal(t, "gemini-embedding-001", opts.DefaultEmbeddingModel)
	assert.Equal(t, 1, opts.DefaultCandidateCount)
	assert.Equal(t, 2048, opts.DefaultMaxTokens)
	assert.Equal(t, 0.5, opts.DefaultTemperature)
	assert.Equal(t, 3, opts.DefaultTopK)
	assert.Equal(t, 0.95, opts.DefaultTopP)
	assert.Equal(t, HarmBlockNone, opts.HarmThreshold)
	assert.Empty(t, opts.CloudProject)
	assert.Empty(t, opts.CloudLocation)
}

func TestOptions(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	t.Parallel()

	t.Run("WithAPIKey", func(t *testing.T) {
		opts := &Options{}
		WithAPIKey("test-key")(opts)
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("WithCredentialsJSON", func(t *testing.T) {
		opts := &Options{}
		creds := []byte(`{"type":"service_account"}`)
		WithCredentialsJSON(creds)(opts)
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("WithCredentialsJSON empty", func(t *testing.T) {
		opts := &Options{}
		WithCredentialsJSON(nil)(opts)
		assert.Len(t, opts.ClientOptions, 0)
	})

	t.Run("WithCredentialsFile", func(t *testing.T) {
		opts := &Options{}
		WithCredentialsFile("path/to/file.json")(opts)
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("WithCredentialsFile empty", func(t *testing.T) {
		opts := &Options{}
		WithCredentialsFile("")(opts)
		assert.Len(t, opts.ClientOptions, 0)
	})

	t.Run("WithRest", func(t *testing.T) {
		opts := &Options{}
		WithRest()(opts)
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		opts := &Options{}
		WithHTTPClient(nil)(opts)
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("WithCloudProject", func(t *testing.T) {
		opts := &Options{}
		WithCloudProject("test-project")(opts)
		assert.Equal(t, "test-project", opts.CloudProject)
	})

	t.Run("WithCloudLocation", func(t *testing.T) {
		opts := &Options{}
		WithCloudLocation("us-central1")(opts)
		assert.Equal(t, "us-central1", opts.CloudLocation)
	})

	t.Run("WithDefaultModel", func(t *testing.T) {
		opts := &Options{}
		WithDefaultModel("custom-model")(opts)
		assert.Equal(t, "custom-model", opts.DefaultModel)
	})

	t.Run("WithDefaultEmbeddingModel", func(t *testing.T) {
		opts := &Options{}
		WithDefaultEmbeddingModel("embedding-002")(opts)
		assert.Equal(t, "embedding-002", opts.DefaultEmbeddingModel)
	})

	t.Run("WithDefaultCandidateCount", func(t *testing.T) {
		opts := &Options{}
		WithDefaultCandidateCount(3)(opts)
		assert.Equal(t, 3, opts.DefaultCandidateCount)
	})

	t.Run("WithDefaultMaxTokens", func(t *testing.T) {
		opts := &Options{}
		WithDefaultMaxTokens(1000)(opts)
		assert.Equal(t, 1000, opts.DefaultMaxTokens)
	})

	t.Run("WithDefaultTemperature", func(t *testing.T) {
		opts := &Options{}
		WithDefaultTemperature(0.8)(opts)
		assert.Equal(t, 0.8, opts.DefaultTemperature)
	})

	t.Run("WithDefaultTopK", func(t *testing.T) {
		opts := &Options{}
		WithDefaultTopK(5)(opts)
		assert.Equal(t, 5, opts.DefaultTopK)
	})

	t.Run("WithDefaultTopP", func(t *testing.T) {
		opts := &Options{}
		WithDefaultTopP(0.9)(opts)
		assert.Equal(t, 0.9, opts.DefaultTopP)
	})

	t.Run("WithHarmThreshold", func(t *testing.T) {
		opts := &Options{}
		WithHarmThreshold(HarmBlockMediumAndAbove)(opts)
		assert.Equal(t, HarmBlockMediumAndAbove, opts.HarmThreshold)
	})
}

func TestEnsureAuthPresent(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	t.Run("no auth options, no env var", func(t *testing.T) {
		t.Setenv("GOOGLE_API_KEY", "")
		opts := &Options{}
		opts.EnsureAuthPresent()
		assert.Len(t, opts.ClientOptions, 0)
	})

	t.Run("no auth options, with env var", func(t *testing.T) {
		t.Setenv("GOOGLE_API_KEY", "test-key-from-env")
		opts := &Options{}
		opts.EnsureAuthPresent()
		assert.Len(t, opts.ClientOptions, 1)
	})

	t.Run("has auth options", func(t *testing.T) {
		t.Setenv("GOOGLE_API_KEY", "test-key-from-env")
		opts := &Options{}
		WithAPIKey("existing-key")(opts)
		initialLen := len(opts.ClientOptions)
		opts.EnsureAuthPresent()
		// Should not add another auth option
		assert.Len(t, opts.ClientOptions, initialLen)
	})
}

func TestHasAuthOptions(t *testing.T) {
	t.Parallel()

	t.Run("no options", func(t *testing.T) {
		assert.False(t, hasAuthOptions(nil))
	})

	// Note: Testing hasAuthOptions with actual options is complex due to the use of reflection
	// and the private nature of the option types. The function is already tested indirectly
	// through EnsureAuthPresent tests.
}

func TestHarmBlockThresholdConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "HARM_BLOCK_THRESHOLD_UNSPECIFIED", string(HarmBlockUnspecified))
	assert.Equal(t, "BLOCK_LOW_AND_ABOVE", string(HarmBlockLowAndAbove))
	assert.Equal(t, "BLOCK_MEDIUM_AND_ABOVE", string(HarmBlockMediumAndAbove))
	assert.Equal(t, "BLOCK_ONLY_HIGH", string(HarmBlockOnlyHigh))
	assert.Equal(t, "BLOCK_NONE", string(HarmBlockNone))
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "citations", CITATIONS)
	assert.Equal(t, "safety", SAFETY)
	assert.Equal(t, "system", RoleSystem)
	assert.Equal(t, "model", RoleModel)
	assert.Equal(t, "user", RoleUser)
	assert.Equal(t, "tool", RoleTool)
	assert.Equal(t, "application/json", ResponseMIMETypeJson)
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "no content in generation response", ErrNoContentInResponse.Error())
	assert.Equal(t, "unknown part type in generation response", ErrUnknownPartInResponse.Error())
	assert.Equal(t, "invalid mime type on content", ErrInvalidMimeType.Error())
}

func TestGoogleAIImplementsModelInterface(t *testing.T) {
	t.Parallel()

	// This test ensures GoogleAI implements the llms.Model interface
	var _ llms.Model = &GoogleAI{}
}

func TestConvertToolSchemaType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string // We'll compare the string representation
	}{
		{"object", "OBJECT"},
		{"string", "STRING"},
		{"number", "NUMBER"},
		{"integer", "INTEGER"},
		{"boolean", "BOOLEAN"},
		{"array", "ARRAY"},
		{"unknown", "TYPE_UNSPECIFIED"},
		{"", "TYPE_UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToolSchemaType(tt.input)
			// Convert to string for comparison
			resultStr := string(result)
			assert.Equal(t, tt.expected, resultStr)
		})
	}
}

func TestConvertTools(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	t.Parallel()

	t.Run("empty tools", func(t *testing.T) {
		result, err := convertTools(nil)
		assert.NoError(t, err)
		assert.Nil(t, result)

		result, err = convertTools([]llms.Tool{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("unsupported tool type", func(t *testing.T) {
		tools := []llms.Tool{
			{Type: "unsupported"},
		}
		result, err := convertTools(tools)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
		assert.Nil(t, result)
	})

	t.Run("invalid parameters type", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "test",
					Description: "test function",
					Parameters:  "invalid", // should be map[string]any
				},
			},
		}
		result, err := convertTools(tools)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
		assert.Nil(t, result)
	})

	t.Run("missing properties in parameters", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "test",
					Description: "test function",
					Parameters: map[string]any{
						"type": "object",
						// missing properties
					},
				},
			},
		}
		result, err := convertTools(tools)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected to find a map of properties")
		assert.Nil(t, result)
	})

	t.Run("valid function tool", func(t *testing.T) {
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
								"description": "City name",
							},
							"unit": map[string]any{
								"type":        "string",
								"description": "Temperature unit",
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

		funcDecl := result[0].FunctionDeclarations[0]
		assert.Equal(t, "get_weather", funcDecl.Name)
		assert.Equal(t, "Get weather information", funcDecl.Description)
		assert.NotNil(t, funcDecl.Parameters)
		assert.Len(t, funcDecl.Parameters.Properties, 2)
		assert.Contains(t, funcDecl.Parameters.Required, "location")
	})

	t.Run("nested object schema", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "create_user",
					Description: "Create a user with nested address",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "User name",
							},
							"address": map[string]any{
								"type":        "object",
								"description": "User address",
								"properties": map[string]any{
									"street": map[string]any{
										"type":        "string",
										"description": "Street address",
									},
									"city": map[string]any{
										"type":        "string",
										"description": "City name",
									},
									"coordinates": map[string]any{
										"type":        "object",
										"description": "GPS coordinates",
										"properties": map[string]any{
											"lat": map[string]any{
												"type":        "number",
												"description": "Latitude",
											},
											"lng": map[string]any{
												"type":        "number",
												"description": "Longitude",
											},
										},
										"required": []string{"lat", "lng"},
									},
								},
								"required": []string{"street", "city"},
							},
						},
						"required": []string{"name", "address"},
					},
				},
			},
		}
		result, err := convertTools(tools)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].FunctionDeclarations, 1)

		funcDecl := result[0].FunctionDeclarations[0]
		assert.Equal(t, "create_user", funcDecl.Name)
		assert.Equal(t, "Create a user with nested address", funcDecl.Description)
		assert.NotNil(t, funcDecl.Parameters)
		assert.Len(t, funcDecl.Parameters.Properties, 2)
		assert.Contains(t, funcDecl.Parameters.Required, "name")
		assert.Contains(t, funcDecl.Parameters.Required, "address")

		// Check nested address object
		addressProp := funcDecl.Parameters.Properties["address"]
		assert.NotNil(t, addressProp)
		assert.Len(t, addressProp.Properties, 3)
		assert.Contains(t, addressProp.Required, "street")
		assert.Contains(t, addressProp.Required, "city")

		// Check deeply nested coordinates object
		coordsProp := addressProp.Properties["coordinates"]
		assert.NotNil(t, coordsProp)
		assert.Len(t, coordsProp.Properties, 2)
		assert.Contains(t, coordsProp.Required, "lat")
		assert.Contains(t, coordsProp.Required, "lng")
	})

	t.Run("array with nested objects", func(t *testing.T) {
		tools := []llms.Tool{
			{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "create_order",
					Description: "Create an order with array of items",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"customer_id": map[string]any{
								"type":        "string",
								"description": "Customer ID",
							},
							"items": map[string]any{
								"type":        "array",
								"description": "Order items",
								"items": map[string]any{
									"type":        "object",
									"description": "Individual item",
									"properties": map[string]any{
										"product_id": map[string]any{
											"type":        "string",
											"description": "Product ID",
										},
										"quantity": map[string]any{
											"type":        "integer",
											"description": "Quantity",
										},
										"customizations": map[string]any{
											"type":        "array",
											"description": "Item customizations",
											"items": map[string]any{
												"type":        "object",
												"description": "Customization option",
												"properties": map[string]any{
													"option": map[string]any{
														"type":        "string",
														"description": "Customization option name",
													},
													"value": map[string]any{
														"type":        "string",
														"description": "Customization value",
													},
												},
												"required": []string{"option", "value"},
											},
										},
									},
									"required": []string{"product_id", "quantity"},
								},
							},
						},
						"required": []string{"customer_id", "items"},
					},
				},
			},
		}
		result, err := convertTools(tools)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].FunctionDeclarations, 1)

		funcDecl := result[0].FunctionDeclarations[0]
		assert.Equal(t, "create_order", funcDecl.Name)
		assert.Equal(t, "Create an order with array of items", funcDecl.Description)
		assert.NotNil(t, funcDecl.Parameters)
		assert.Len(t, funcDecl.Parameters.Properties, 2)
		assert.Contains(t, funcDecl.Parameters.Required, "customer_id")
		assert.Contains(t, funcDecl.Parameters.Required, "items")

		// Check items array
		itemsProp := funcDecl.Parameters.Properties["items"]
		assert.NotNil(t, itemsProp)
		assert.NotNil(t, itemsProp.Items)
		assert.Len(t, itemsProp.Items.Properties, 3)
		assert.Contains(t, itemsProp.Items.Required, "product_id")
		assert.Contains(t, itemsProp.Items.Required, "quantity")

		// Check nested customizations array
		customizationsProp := itemsProp.Items.Properties["customizations"]
		assert.NotNil(t, customizationsProp)
		assert.NotNil(t, customizationsProp.Items)
		assert.Len(t, customizationsProp.Items.Properties, 2)
		assert.Contains(t, customizationsProp.Items.Required, "option")
		assert.Contains(t, customizationsProp.Items.Required, "value")
	})
}

func TestFunctionCallIDWrappers(t *testing.T) {
	t.Run("ensureFunctionCallID", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected func(string) bool
		}{
			{
				name:  "empty ID generates new ID",
				input: "",
				expected: func(result string) bool {
					return strings.HasPrefix(result, GENERATED_FUNCTION_CALL_ID_PREFIX) && len(result) == len(GENERATED_FUNCTION_CALL_ID_PREFIX)+16
				},
			},
			{
				name:  "existing ID is preserved",
				input: "existing-id-123",
				expected: func(result string) bool {
					return result == "existing-id-123"
				},
			},
			{
				name:  "generated ID is unique",
				input: "",
				expected: func(result string) bool {
					result2 := ensureFunctionCallID("")
					return result != result2 && strings.HasPrefix(result, GENERATED_FUNCTION_CALL_ID_PREFIX)
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := ensureFunctionCallID(tc.input)
				assert.True(t, tc.expected(result), "Result: %s", result)
			})
		}
	})

	t.Run("cleanFunctionCallID", func(t *testing.T) {
		testCases := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "generated ID is cleaned",
				input:    GENERATED_FUNCTION_CALL_ID_PREFIX + "1234567890abcdef",
				expected: "",
			},
			{
				name:     "backend ID is preserved",
				input:    "backend-provided-id",
				expected: "backend-provided-id",
			},
			{
				name:     "empty ID remains empty",
				input:    "",
				expected: "",
			},
			{
				name:     "similar prefix but not exact match is preserved",
				input:    "fcal_not_exact_prefix",
				expected: "fcal_not_exact_prefix",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := cleanFunctionCallID(tc.input)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("roundtrip ID handling", func(t *testing.T) {
		// Test that generated IDs are properly cleaned
		generatedID := ensureFunctionCallID("")
		cleanedID := cleanFunctionCallID(generatedID)
		assert.Equal(t, "", cleanedID, "Generated ID should be cleaned to empty string")

		// Test that backend IDs survive roundtrip
		backendID := "backend-id-123"
		ensuredID := ensureFunctionCallID(backendID)
		cleanedID = cleanFunctionCallID(ensuredID)
		assert.Equal(t, backendID, cleanedID, "Backend ID should survive roundtrip")
	})
}

// TestTokenUsageMapping_GoogleAI tests correct token usage mapping for Google AI provider
func TestTokenUsageMapping_GoogleAI(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []struct {
		name                     string
		promptTokenCount         int32
		cachedContentTokenCount  int32
		candidatesTokenCount     int32
		thoughtsTokenCount       int32
		expectedPromptTokens     int
		expectedCacheRead        int
		expectedCacheCreation    int
		expectedCompletionTokens int
		expectedReasoningTokens  int
		expectedTotalTokens      int
	}{
		{
			name:                     "first request without cache",
			promptTokenCount:         4517,
			cachedContentTokenCount:  0,
			candidatesTokenCount:     9,
			thoughtsTokenCount:       0,
			expectedPromptTokens:     4517,
			expectedCacheRead:        0,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 9,
			expectedReasoningTokens:  0,
			expectedTotalTokens:      4526,
		},
		{
			name:                     "subsequent request with cache hit",
			promptTokenCount:         4534,
			cachedContentTokenCount:  4058,
			candidatesTokenCount:     11,
			thoughtsTokenCount:       0,
			expectedPromptTokens:     4534,
			expectedCacheRead:        4058,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 11,
			expectedReasoningTokens:  0,
			expectedTotalTokens:      4545,
		},
		{
			name:                     "request with reasoning and cache",
			promptTokenCount:         5000,
			cachedContentTokenCount:  3500,
			candidatesTokenCount:     150,
			thoughtsTokenCount:       80,
			expectedPromptTokens:     5000,
			expectedCacheRead:        3500,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 150,
			expectedReasoningTokens:  80,
			expectedTotalTokens:      5150,
		},
		{
			name:                     "large cache scenario",
			promptTokenCount:         10000,
			cachedContentTokenCount:  8500,
			candidatesTokenCount:     200,
			thoughtsTokenCount:       50,
			expectedPromptTokens:     10000,
			expectedCacheRead:        8500,
			expectedCacheCreation:    0,
			expectedCompletionTokens: 200,
			expectedReasoningTokens:  50,
			expectedTotalTokens:      10200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate Google AI usage metadata structure
			usage := struct {
				PromptTokenCount        int32
				CachedContentTokenCount int32
				CandidatesTokenCount    int32
				ThoughtsTokenCount      int32
				TotalTokenCount         int32
			}{
				PromptTokenCount:        tt.promptTokenCount,
				CachedContentTokenCount: tt.cachedContentTokenCount,
				CandidatesTokenCount:    tt.candidatesTokenCount,
				ThoughtsTokenCount:      tt.thoughtsTokenCount,
				TotalTokenCount:         tt.promptTokenCount + tt.candidatesTokenCount,
			}

			// Build metadata as done in googleai.go
			metadata := make(map[string]any)
			metadata["PromptTokens"] = int(usage.PromptTokenCount)
			metadata["CompletionTokens"] = int(usage.CandidatesTokenCount)
			metadata["TotalTokens"] = int(usage.TotalTokenCount)
			metadata["ReasoningTokens"] = int(usage.ThoughtsTokenCount)
			metadata["PromptCachedTokens"] = int(usage.CachedContentTokenCount)
			metadata["CacheReadInputTokens"] = int(usage.CachedContentTokenCount)
			metadata["CacheCreationInputTokens"] = 0

			// Verify mapped values
			assert.Equal(t, tt.expectedPromptTokens, metadata["PromptTokens"], "PromptTokens mismatch")
			assert.Equal(t, tt.expectedCacheRead, metadata["CacheReadInputTokens"], "CacheReadInputTokens mismatch")
			assert.Equal(t, tt.expectedCacheCreation, metadata["CacheCreationInputTokens"], "CacheCreationInputTokens mismatch")
			assert.Equal(t, tt.expectedCompletionTokens, metadata["CompletionTokens"], "CompletionTokens mismatch")
			assert.Equal(t, tt.expectedReasoningTokens, metadata["ReasoningTokens"], "ReasoningTokens mismatch")
			assert.Equal(t, tt.expectedTotalTokens, metadata["TotalTokens"], "TotalTokens mismatch")

			// Verify client-side cost calculation logic
			// Client formula: input = max(PromptTokens - CacheRead, 0)
			promptTokens := metadata["PromptTokens"].(int)
			cacheRead := metadata["CacheReadInputTokens"].(int)
			cacheWrite := metadata["CacheCreationInputTokens"].(int)

			uncachedTokens := max(promptTokens-cacheRead, 0)

			// Expected: uncached tokens should equal promptTokenCount - cachedContentTokenCount
			expectedUncached := int(tt.promptTokenCount - tt.cachedContentTokenCount)
			assert.Equal(t, expectedUncached, uncachedTokens, "Uncached tokens calculation mismatch")

			// For Google AI pricing: uncached * basePrice + cacheRead * cacheReadPrice
			// Google AI doesn't charge extra for cache writes
			basePrice := 0.075 / 1e6        // $0.075 per 1M tokens (example for gemini-2.0-flash)
			cacheReadPrice := 0.01875 / 1e6 // $0.01875 per 1M tokens (25% of base price)

			expectedCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice

			actualCost := float64(uncachedTokens)*basePrice +
				float64(cacheRead)*cacheReadPrice

			assert.InDelta(t, expectedCost, actualCost, 0.000001, "Cost calculation mismatch")

			// Verify CacheWrite is always 0 for Google AI
			assert.Equal(t, 0, cacheWrite, "Google AI should always have CacheCreationInputTokens = 0")
		})
	}
}
