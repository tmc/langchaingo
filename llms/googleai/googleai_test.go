package googleai

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

// apiKeyTransport adds the API key to requests
// This is needed because the Google API library doesn't add the API key
// when WithHTTPClient is used with WithAPIKey
type apiKeyTransport struct {
	wrapped http.RoundTripper
	apiKey  string
}

func (t *apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	newReq := req.Clone(req.Context())
	q := newReq.URL.Query()
	if q.Get("key") == "" && t.apiKey != "" {
		q.Set("key", t.apiKey)
		newReq.URL.RawQuery = q.Encode()
	}
	return t.wrapped.RoundTrip(newReq)
}

func newHTTPRRClient(t *testing.T, opts ...Option) *GoogleAI {
	t.Helper()

	// Skip if no credentials and no recording
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	// Create httprr with API key transport wrapper
	// This is necessary because the Google API library doesn't add the API key
	// when a custom HTTP client is provided via WithHTTPClient
	apiKey := os.Getenv("GOOGLE_API_KEY")
	transport := httputil.DefaultTransport
	if apiKey != "" {
		transport = &apiKeyTransport{
			wrapped: transport,
			apiKey:  apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)

	// Scrub API key for security in recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Configure client with httprr
	opts = append(opts, WithRest(), WithHTTPClient(rr.Client()))

	llm, err := New(context.Background(), opts...)
	require.NoError(t, err)
	return llm
}

func TestGoogleAIGenerateContent(t *testing.T) {

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.Contains(t, resp.Choices[0].Content, "Paris")
}

func TestGoogleAIGenerateContentWithMultipleMessages(t *testing.T) {

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("My name is Alice"),
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart("Nice to meet you, Alice!"),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's my name?"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-1.5-flash"))
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.Contains(t, resp.Choices[0].Content, "Alice")
}

func TestGoogleAIGenerateContentWithSystemMessage(t *testing.T) {

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are a helpful assistant that always responds in haiku format."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me about the ocean"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-1.5-flash"))
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestGoogleAICall(t *testing.T) {

	llm := newHTTPRRClient(t)

	output, err := llm.Call(context.Background(), "What is 2 + 2?")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "4")
}

func TestGoogleAICreateEmbedding(t *testing.T) {

	llm := newHTTPRRClient(t)

	texts := []string{"hello world", "goodbye world", "hello world"}

	embeddings, err := llm.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
	assert.NotEmpty(t, embeddings[2])
	// First and third should be identical since they're the same text
	assert.Equal(t, embeddings[0], embeddings[2])
}

func TestGoogleAIWithOptions(t *testing.T) {

	llm := newHTTPRRClient(t,
		WithDefaultModel("gemini-1.5-flash"),
		WithDefaultMaxTokens(100),
		WithDefaultTemperature(0.1),
	)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Count from 1 to 5"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestGoogleAIWithStreaming(t *testing.T) {

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a short story about a cat"),
			},
		},
	}

	var streamedContent string
	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			streamedContent += string(chunk)
			return nil
		}),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.NotEmpty(t, streamedContent)
	// Check for cat-related content (the AI might use the cat's name instead of "cat")
	catRelated := strings.Contains(strings.ToLower(streamedContent), "cat") ||
		strings.Contains(streamedContent, "Clementine") ||
		strings.Contains(streamedContent, "purr") ||
		strings.Contains(streamedContent, "meow")
	assert.True(t, catRelated, "Response should contain cat-related content")
}

func TestGoogleAIWithTools(t *testing.T) {

	llm := newHTTPRRClient(t)

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getWeather",
				Description: "Get the weather for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What's the weather in New York?"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithTools(tools),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)

	// Check if tool call was made
	if len(resp.Choices[0].ToolCalls) > 0 {
		toolCall := resp.Choices[0].ToolCalls[0]
		assert.Equal(t, "getWeather", toolCall.FunctionCall.Name)
		assert.Contains(t, toolCall.FunctionCall.Arguments, "New York")
	}
}

func TestGoogleAIWithJSONMode(t *testing.T) {

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("List three colors as a JSON array"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithJSONMode(),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	// Response should be valid JSON
	assert.Contains(t, resp.Choices[0].Content, "[")
	assert.Contains(t, resp.Choices[0].Content, "]")
}

func TestGoogleAIErrorHandling(t *testing.T) {

	// Skip test if no credentials and recording is missing
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Scrub API key for security in recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "invalid-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Create client with invalid API key
	llm, err := New(context.Background(),
		WithRest(),
		WithAPIKey("invalid-key"),
		WithHTTPClient(rr.Client()),
	)
	require.NoError(t, err)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello"),
			},
		},
	}

	_, err = llm.GenerateContent(context.Background(), content)
	assert.Error(t, err)
}

func TestGoogleAIMultiModalContent(t *testing.T) {

	llm := newHTTPRRClient(t)

	// Read the test image
	imageData, err := os.ReadFile("shared_test/testdata/parrot-icon.png")
	require.NoError(t, err)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.BinaryPart("image/png", imageData),
				llms.TextPart("What is in this image?"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithModel("gemini-1.5-flash"),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestGoogleAIBatchEmbedding(t *testing.T) {

	llm := newHTTPRRClient(t)

	// Test with more than 100 texts to trigger batching
	texts := make([]string, 105)
	for i := range texts {
		texts[i] = "test text " + string(rune('a'+i%26))
	}

	embeddings, err := llm.CreateEmbedding(context.Background(), texts)

	require.NoError(t, err)
	assert.Len(t, embeddings, 105)
	for i, emb := range embeddings {
		assert.NotEmpty(t, emb, "embedding at index %d should not be empty", i)
	}
}

func TestGoogleAIWithHarmThreshold(t *testing.T) {

	llm := newHTTPRRClient(t,
		WithHarmThreshold(HarmBlockNone),
	)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me about safety features"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestGoogleAIToolCallResponse(t *testing.T) {

	llm := newHTTPRRClient(t)

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "calculate",
				Description: "Perform a calculation",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"expression": map[string]any{
							"type":        "string",
							"description": "Mathematical expression to evaluate",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}

	// Initial request
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 15 * 7?"),
			},
		},
	}

	resp1, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithTools(tools),
	)
	require.NoError(t, err)
	require.NotNil(t, resp1)

	// If tool was called, send back response
	if len(resp1.Choices[0].ToolCalls) > 0 {
		// Add assistant's tool call to history
		content = append(content, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{resp1.Choices[0].ToolCalls[0]},
		})

		// Add tool response
		content = append(content, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					Name:    resp1.Choices[0].ToolCalls[0].FunctionCall.Name,
					Content: "105",
				},
			},
		})

		// Get final response
		resp2, err := llm.GenerateContent(
			context.Background(),
			content,
			llms.WithTools(tools),
		)
		require.NoError(t, err)
		require.NotNil(t, resp2)
		assert.Contains(t, resp2.Choices[0].Content, "105")
	}
}

func Test_convertTools(t *testing.T) {
	tests := []struct {
		name        string
		tools       []llms.Tool
		expected    []*genai.FunctionDeclaration
		expectedErr string
	}{
		{
			name:  "no tools",
			tools: nil,
		},
		{
			name: "unsupported tool type",
			tools: []llms.Tool{
				{Type: "unsupported"},
			},
			expectedErr: `unsupported type "unsupported", want 'function'`,
		},
		{
			name: "unsupported tool parameter type",
			tools: []llms.Tool{
				{Type: "function", Function: &llms.FunctionDefinition{Parameters: "unsupported"}},
			},
			expectedErr: `unsupported type string of Parameters`,
		},
		{
			name: "missing type",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{},
				}},
			},
			expectedErr: "type is missing",
		},
		{
			name: "type is not string",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type": 123,
					},
				}},
			},
			expectedErr: "type is not a string",
		},
		{
			name: "description is not string",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":        "object",
						"description": 123,
					},
				}},
			},
			expectedErr: "description is not a string",
		},
		{
			name: "enum is not a slice",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type": "object",
						"enum": 123,
					},
				}},
			},
			expectedErr: "enum is not a slice",
		},
		{
			name: "required is not a slice",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":     "object",
						"required": 123,
					},
				}},
			},
			expectedErr: "required field is not a slice",
		},
		{
			name: "required items are not strings",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":     "object",
						"required": []any{"string", 123},
					},
				}},
			},
			expectedErr: "expected string for required",
		},
		{
			name: "items is not a map",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":  "object",
						"items": 123,
					},
				}},
			},
			expectedErr: "items is not a map",
		},
		{
			name: "properties is not a map",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Parameters: map[string]any{
						"type":       "object",
						"properties": 123,
					},
				}},
			},
			expectedErr: "properties is not a map",
		},
		{
			name: "use given schema",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters:  &genai.Schema{},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters:  &genai.Schema{},
			}},
		},
		{
			name: "simple parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type":        "string",
						"description": "A simple string parameter",
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type:        genai.TypeString,
					Description: "A simple string parameter",
				},
			}},
		},
		{
			name: "object parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "string",
								"description": "A string parameter",
							},
							"number": map[string]any{
								"type":        "number",
								"description": "A number parameter",
							},
							"integer": map[string]any{
								"type":        "integer",
								"description": "An integer parameter",
							},
							"boolean": map[string]any{
								"type":        "boolean",
								"description": "A boolean parameter",
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeString,
							Description: "A string parameter",
						},
						"number": {
							Type:        genai.TypeNumber,
							Description: "A number parameter",
						},
						"integer": {
							Type:        genai.TypeInteger,
							Description: "An integer parameter",
						},
						"boolean": {
							Type:        genai.TypeBoolean,
							Description: "A boolean parameter",
						},
					},
				},
			}},
		},
		{
			name: "required parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "string",
								"description": "A string parameter",
							},
							"number": map[string]any{
								"type":        "number",
								"description": "A number parameter",
							},
						},
						"required": []string{"string"},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeString,
							Description: "A string parameter",
						},
						"number": {
							Type:        genai.TypeNumber,
							Description: "A number parameter",
						},
					},
					Required: []string{"string"},
				},
			}},
		},
		{
			name: "enum parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"enum": map[string]any{
								"type":        "string",
								"description": "A enum parameter",
								"enum":        []string{"option1", "option2", "option3"},
							},
							"anyEnum": map[string]any{
								"type":        "string",
								"description": "A any enum parameter",
								"enum":        []any{1, 1.2, "option3"},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"enum": {
							Type:        genai.TypeString,
							Description: "A enum parameter",
							Enum:        []string{"option1", "option2", "option3"},
						},
						"anyEnum": {
							Type:        genai.TypeString,
							Description: "A any enum parameter",
							Enum:        []string{"1", "1.2", "option3"},
						},
					},
				},
			}},
		},
		{
			name: "simple array parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"string": map[string]any{
								"type":        "array",
								"description": "A string array parameter",
								"items": map[string]any{
									"type": "string",
								},
							},
							"number": map[string]any{
								"type":        "array",
								"description": "A number array parameter",
								"items": map[string]any{
									"type": "number",
								},
							},
							"integer": map[string]any{
								"type":        "array",
								"description": "An integer array parameter",
								"items": map[string]any{
									"type": "integer",
								},
							},
							"boolean": map[string]any{
								"type":        "array",
								"description": "A boolean array parameter",
								"items": map[string]any{
									"type": "boolean",
								},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"string": {
							Type:        genai.TypeArray,
							Description: "A string array parameter",
							Items: &genai.Schema{
								Type: genai.TypeString,
							},
						},
						"number": {
							Type:        genai.TypeArray,
							Description: "A number array parameter",
							Items: &genai.Schema{
								Type: genai.TypeNumber,
							},
						},
						"integer": {
							Type:        genai.TypeArray,
							Description: "An integer array parameter",
							Items: &genai.Schema{
								Type: genai.TypeInteger,
							},
						},
						"boolean": {
							Type:        genai.TypeArray,
							Description: "A boolean array parameter",
							Items: &genai.Schema{
								Type: genai.TypeBoolean,
							},
						},
					},
				},
			}},
		},
		{
			name: "complex array parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"array": map[string]any{
								"type":        "array",
								"description": "A array parameter",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"string": map[string]any{
											"type":        "string",
											"description": "A string parameter",
										},
										"number": map[string]any{
											"type":        "number",
											"description": "A number parameter",
										},
									},
									"required": []string{"string"},
								},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"array": {
							Type:        genai.TypeArray,
							Description: "A array parameter",
							Items: &genai.Schema{
								Type: genai.TypeObject,
								Properties: map[string]*genai.Schema{
									"string": {
										Type:        genai.TypeString,
										Description: "A string parameter",
									},
									"number": {
										Type:        genai.TypeNumber,
										Description: "A number parameter",
									},
								},
								Required: []string{"string"},
							},
						},
					},
				},
			}},
		},
		{
			name: "complex object parameter",
			tools: []llms.Tool{{
				Type: "function",
				Function: &llms.FunctionDefinition{
					Name:        "TestFunction",
					Description: "A test function",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"param1": map[string]any{
								"type":        "object",
								"description": "A object parameter",
								"properties": map[string]any{
									"array": map[string]any{
										"type":        "array",
										"description": "A array parameter",
										"items": map[string]any{
											"type": "object",
											"properties": map[string]any{
												"string": map[string]any{
													"type":        "string",
													"description": "A string parameter",
												},
												"number": map[string]any{
													"type":        "number",
													"description": "A number parameter",
												},
											},
											"required": []string{"string"},
										},
									},
								},
							},
							"param2": map[string]any{
								"type":        "string",
								"description": "A string parameter",
								"enum":        []string{"option1", "option2", "option3"},
							},
						},
					},
				}},
			},
			expected: []*genai.FunctionDeclaration{{
				Name:        "TestFunction",
				Description: "A test function",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {
							Type:        genai.TypeObject,
							Description: "A object parameter",
							Properties: map[string]*genai.Schema{
								"array": {
									Type:        genai.TypeArray,
									Description: "A array parameter",
									Items: &genai.Schema{
										Type: genai.TypeObject,
										Properties: map[string]*genai.Schema{
											"string": {
												Type:        genai.TypeString,
												Description: "A string parameter",
											},
											"number": {
												Type:        genai.TypeNumber,
												Description: "A number parameter",
											},
										},
										Required: []string{"string"},
									},
								},
							},
						},
						"param2": {
							Type:        genai.TypeString,
							Description: "A string parameter",
							Enum:        []string{"option1", "option2", "option3"},
						},
					},
				},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := convertTools(tt.tools)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			}

			if tt.expected == nil {
				assert.Empty(t, result)
			} else {
				assert.Len(t, result, 1)
				assert.Equal(t, tt.expected, result[0].FunctionDeclarations)
			}
		})
	}
}
