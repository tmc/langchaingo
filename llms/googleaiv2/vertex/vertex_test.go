package vertex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/httputil"
	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/googleaiv2"
)

func newHTTPRRClient(t *testing.T, opts ...googleaiv2.Option) *Vertex {
	t.Helper()

	// Always check for recordings first - prefer recordings over environment variables
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	// Temporarily unset Google API key environment variable to prevent bypass
	oldKey := os.Getenv("GOOGLE_API_KEY")
	os.Unsetenv("GOOGLE_API_KEY")
	t.Cleanup(func() {
		if oldKey != "" {
			os.Setenv("GOOGLE_API_KEY", oldKey)
		}
	})

	// Use httputil.DefaultTransport - httprr handles wrapping
	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Configure client with httprr and test values
	opts = append(opts,
		googleaiv2.WithHTTPClient(rr.Client()),
		googleaiv2.WithCloudProject("test-project"),
		googleaiv2.WithCloudLocation("us-central1"),
	)

	llm, err := New(context.Background(), opts...)
	require.NoError(t, err)
	return llm
}

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestVertexGenerateContent(t *testing.T) {
	t.Parallel()

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

func TestVertexGenerateContentWithMultipleMessages(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("My name is Bob"),
			},
		},
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart("Nice to meet you, Bob!"),
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
	assert.Contains(t, resp.Choices[0].Content, "Bob")
}

func TestVertexGenerateContentWithSystemMessage(t *testing.T) {
	t.Parallel()

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
				llms.TextPart("Tell me about the moon"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-1.5-flash"))
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestVertexCall(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	output, err := llm.Call(context.Background(), "What is 3 + 3?")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "6")
}

func TestVertexCreateEmbedding(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	texts := []string{"hello vertex", "goodbye vertex", "hello vertex"}

	embeddings, err := llm.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
	assert.NotEmpty(t, embeddings[2])
	// First and third should be identical since they're the same text
	assert.Equal(t, embeddings[0], embeddings[2])
}

func TestVertexWithOptions(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t,
		googleaiv2.WithDefaultModel("gemini-1.5-flash"),
		googleaiv2.WithDefaultMaxTokens(150),
		googleaiv2.WithDefaultTemperature(0.2),
	)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("List the primary colors"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestVertexWithStreaming(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a short story about a robot"),
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
	assert.Contains(t, streamedContent, "robot")
}

func TestVertexWithTools(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getTemperature",
				Description: "Get the temperature for a location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get temperature for",
						},
						"unit": map[string]any{
							"type":        "string",
							"description": "Temperature unit (celsius or fahrenheit)",
							"enum":        []string{"celsius", "fahrenheit"},
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
				llms.TextPart("What's the temperature in Tokyo?"),
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
		assert.Equal(t, "getTemperature", toolCall.FunctionCall.Name)
		assert.Contains(t, toolCall.FunctionCall.Arguments, "Tokyo")
	}
}

func TestVertexWithJSONMode(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("List three animals as a JSON array"),
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

func TestVertexMultiModalContent(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	// Create a small test image data
	imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.BinaryPart("image/png", imageData),
				llms.TextPart("Describe this image"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithModel("gemini-pro-vision"),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestVertexBatchEmbedding(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	// Test with more than 100 texts to trigger batching
	texts := make([]string, 105)
	for i := range texts {
		texts[i] = "vertex text " + string(rune('a'+i%26))
	}

	embeddings, err := llm.CreateEmbedding(context.Background(), texts)

	require.NoError(t, err)
	assert.Len(t, embeddings, 105)
	for i, emb := range embeddings {
		assert.NotEmpty(t, emb, "embedding at index %d should not be empty", i)
	}
}

func TestVertexWithHarmThreshold(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t,
		googleaiv2.WithHarmThreshold(googleaiv2.HarmBlockLowAndAbove),
	)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me about content moderation"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestVertexToolCallResponse(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "multiply",
				Description: "Multiply two numbers",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"a": map[string]any{
							"type":        "number",
							"description": "First number",
						},
						"b": map[string]any{
							"type":        "number",
							"description": "Second number",
						},
					},
					"required": []string{"a", "b"},
				},
			},
		},
	}

	// Initial request
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 12 times 8?"),
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
					Content: "96",
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
		assert.Contains(t, resp2.Choices[0].Content, "96")
	}
}

func TestVertexWithResponseMIMEType(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Return a JSON object with name and age fields"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithResponseMIMEType("application/json"),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	// Response should be valid JSON
	assert.Contains(t, resp.Choices[0].Content, "{")
	assert.Contains(t, resp.Choices[0].Content, "}")
}

func TestVertexErrorOnConflictingOptions(t *testing.T) {
	t.Parallel()

	llm := newHTTPRRClient(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello"),
			},
		},
	}

	// Should error when both JSONMode and ResponseMIMEType are set
	_, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithJSONMode(),
		llms.WithResponseMIMEType("application/json"),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conflicting options")
}
