package googleai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vendasta/langchaingo/llms"
)

func TestNew(t *testing.T) {
	t.Parallel()

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
				WithAPIKey("test-api-key"),
				WithCloudProject("test-project"),
				WithCloudLocation("us-central1"),
			},
			wantErr: false,
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
			client, err := New(context.Background(), tt.opts...)
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
	assert.Equal(t, "embedding-001", opts.DefaultEmbeddingModel)
	assert.Equal(t, 1, opts.DefaultCandidateCount)
	assert.Equal(t, 2048, opts.DefaultMaxTokens)
	assert.Equal(t, 0.5, opts.DefaultTemperature)
	assert.Equal(t, 3, opts.DefaultTopK)
	assert.Equal(t, 0.95, opts.DefaultTopP)
	assert.Equal(t, HarmBlockOnlyHigh, opts.HarmThreshold)
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

	assert.Equal(t, HarmBlockThreshold(0), HarmBlockUnspecified)
	assert.Equal(t, HarmBlockThreshold(1), HarmBlockLowAndAbove)
	assert.Equal(t, HarmBlockThreshold(2), HarmBlockMediumAndAbove)
	assert.Equal(t, HarmBlockThreshold(3), HarmBlockOnlyHigh)
	assert.Equal(t, HarmBlockThreshold(4), HarmBlockNone)
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
		{"object", "TypeObject"},
		{"string", "TypeString"},
		{"number", "TypeNumber"},
		{"integer", "TypeInteger"},
		{"boolean", "TypeBoolean"},
		{"array", "TypeArray"},
		{"unknown", "TypeUnspecified"},
		{"", "TypeUnspecified"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToolSchemaType(tt.input)
			assert.Equal(t, tt.expected, result.String())
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
