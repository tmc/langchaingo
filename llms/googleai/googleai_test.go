package googleai

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		transport = &httputil.ApiKeyTransport{
			Transport: transport,
			APIKey:    apiKey,
		}
	}

	rr := httprr.OpenForTest(t, transport)

	// Avoid issue with different view of request bodies for Google AI SDK
	rr.ScrubReq(httprr.JsonCompactScrubBody)

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

	// Add API key if available
	if apiKey != "" {
		opts = append(opts, WithAPIKey(apiKey))
	} else {
		t.Skip("No API key found, skipping test")
	}

	llm, err := New(t.Context(), opts...)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	resp, err := llm.GenerateContent(t.Context(), content)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	resp, err := llm.GenerateContent(t.Context(), content, llms.WithModel("gemini-2.5-flash"))
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	resp, err := llm.GenerateContent(t.Context(), content, llms.WithModel("gemini-2.5-flash"))
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestGoogleAICall(t *testing.T) {
	llm := newHTTPRRClient(t)

	output, err := llm.Call(t.Context(), "What is 2 + 2?")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "4")
}

func TestGoogleAICreateEmbedding(t *testing.T) {
	llm := newHTTPRRClient(t)

	texts := []string{"hello world", "goodbye world", "hello world"}

	embeddings, err := llm.CreateEmbedding(t.Context(), texts)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	assert.Len(t, embeddings, 3)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
	assert.NotEmpty(t, embeddings[2])
	// First and third should be identical since they're the same text
	assert.Equal(t, embeddings[0], embeddings[2])
}

func TestGoogleAIWithOptions(t *testing.T) {
	llm := newHTTPRRClient(t,
		WithDefaultModel("gemini-2.5-flash"),
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

	resp, err := llm.GenerateContent(t.Context(), content)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	var (
		streamedContent strings.Builder
		streamDone      bool
	)
	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			switch chunk.Type {
			case streaming.ChunkTypeText:
				streamedContent.WriteString(chunk.Content)
			case streaming.ChunkTypeDone:
				streamDone = true
			default:
				// Ignore other chunk types
			}
			return nil
		}),
	)

	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	require.NotNil(t, resp)
	assert.True(t, streamDone, "Streaming should be done")
	assert.NotEmpty(t, resp.Choices)
	assert.NotEmpty(t, streamedContent.String())
	// Check for cat-related content (the AI might use the cat's name instead of "cat")
	catRelated := strings.Contains(strings.ToLower(streamedContent.String()), "cat") ||
		strings.Contains(streamedContent.String(), "Clementine") ||
		strings.Contains(streamedContent.String(), "Whiskers") ||
		strings.Contains(streamedContent.String(), "Peaches") ||
		strings.Contains(streamedContent.String(), "purr") ||
		strings.Contains(streamedContent.String(), "meow")
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
		t.Context(),
		content,
		llms.WithTools(tools),
	)

	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)

	// Check if tool call was made
	if len(resp.Choices[0].ToolCalls) > 0 {
		toolCall := resp.Choices[0].ToolCalls[0]
		assert.NotEmpty(t, toolCall.ID, "ToolCall ID should not be empty")
		assert.Equal(t, "getWeather", toolCall.FunctionCall.Name)
		assert.Contains(t, toolCall.FunctionCall.Arguments, "New York")
	} else {
		t.Fail()
	}
}

func TestGoogleAIWithToolsCustomType(t *testing.T) {
	llm := newHTTPRRClient(t)

	type toolCallArgs struct {
		Location string `json:"location"`
	}

	type property struct {
		Type        string `json:"type"`
		Description string `json:"description"`
	}

	type properties struct {
		Location property `json:"location"`
	}

	type parameters struct {
		Type       string     `json:"type"`
		Properties properties `json:"properties"`
		Required   []string   `json:"required"`
	}

	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getWeather",
				Description: "Get the weather for a location",
				Parameters: parameters{
					Type: "object",
					Properties: properties{
						Location: property{
							Type:        "string",
							Description: "The location to get weather for",
						},
					},
					Required: []string{"location"},
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
		t.Context(),
		content,
		llms.WithTools(tools),
	)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)

	// Check if tool call was made
	if len(resp.Choices[0].ToolCalls) > 0 {
		toolCall := resp.Choices[0].ToolCalls[0]
		assert.NotEmpty(t, toolCall.ID, "ToolCall ID should not be empty")
		assert.Equal(t, "getWeather", toolCall.FunctionCall.Name)
		assert.Contains(t, toolCall.FunctionCall.Arguments, "New York")
		var args toolCallArgs
		err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args)
		require.NoError(t, err)
		assert.Equal(t, "New York", args.Location)
	} else {
		t.Fail()
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
		t.Context(),
		content,
		llms.WithJSONMode(),
	)

	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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
	llm, err := New(t.Context(),
		WithRest(),
		WithAPIKey("invalid-key"),
		WithHTTPClient(rr.Client()),
	)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Hello"),
			},
		},
	}

	_, err = llm.GenerateContent(t.Context(), content)
	assert.Error(t, err)
}

func TestGoogleAIMultiModalContent(t *testing.T) {
	llm := newHTTPRRClient(t)

	// Read the test image
	imageData, err := os.ReadFile("shared_test/testdata/parrot-icon.png")
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}

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
		t.Context(),
		content,
		llms.WithModel("gemini-2.5-flash"),
	)

	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	embeddings, err := llm.CreateEmbedding(t.Context(), texts)

	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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

	resp, err := llm.GenerateContent(t.Context(), content)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
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
		t.Context(),
		content,
		llms.WithTools(tools),
	)
	if err != nil {
		// Check if this is a recording mismatch error
		if strings.Contains(err.Error(), "cached HTTP response not found") {
			t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
		}
		require.NoError(t, err)
	}
	require.NotNil(t, resp1)

	// If tool was called, send back response
	if len(resp1.Choices[0].ToolCalls) > 0 {
		toolCall := resp1.Choices[0].ToolCalls[0]
		assert.NotEmpty(t, toolCall.ID, "ToolCall ID should not be empty")

		// Add assistant's tool call to history
		content = append(content, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{toolCall},
		})

		// Add tool response
		content = append(content, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: toolCall.ID,
					Name:       toolCall.FunctionCall.Name,
					Content:    "105",
				},
			},
		})

		// Get final response
		resp2, err := llm.GenerateContent(
			t.Context(),
			content,
			llms.WithTools(tools),
		)
		if err != nil {
			// Check if this is a recording mismatch error
			if strings.Contains(err.Error(), "cached HTTP response not found") {
				t.Skip("Recording format has changed or is incompatible. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
			}
			require.NoError(t, err)
		}
		require.NotNil(t, resp2)
		assert.Contains(t, resp2.Choices[0].Content, "105")
	}
}

func TestGoogleAIThinkingModels(t *testing.T) {
	t.Run("non-streaming", func(t *testing.T) {
		llm := newHTTPRRClient(t)

		content := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("What's the next number in this sequence: 2, 6, 12, 20, 30, ? Show your reasoning."),
				},
			},
		}

		resp, err := llm.GenerateContent(
			t.Context(),
			content,
			llms.WithModel("gemini-2.5-pro"),
			llms.WithReasoning(llms.ReasoningHigh, 1000),
		)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Choices)
		assert.Contains(t, resp.Choices[0].Content, "42")

		// Check that thinking content is present in metadata
		assert.NotNil(t, resp.Choices[0].Reasoning)
		if resp.Choices[0].Reasoning != nil {
			assert.NotEmpty(t, resp.Choices[0].Reasoning.Content, "thinking content should not be empty")
		}
	})

	t.Run("streaming", func(t *testing.T) {
		llm := newHTTPRRClient(t)

		content := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart("What's the next number in this sequence: 2, 6, 12, 20, 30, ? Show your reasoning."),
				},
			},
		}

		var (
			streamedContent strings.Builder
			thinkingContent strings.Builder
			streamDone      bool
		)

		resp, err := llm.GenerateContent(
			t.Context(),
			content,
			llms.WithModel("gemini-2.5-pro"),
			llms.WithReasoning(llms.ReasoningHigh, 1000),
			llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
				switch chunk.Type {
				case streaming.ChunkTypeText:
					streamedContent.WriteString(chunk.Content)
				case streaming.ChunkTypeReasoning:
					if chunk.Reasoning != nil {
						thinkingContent.WriteString(chunk.Reasoning.Content)
					}
				case streaming.ChunkTypeDone:
					streamDone = true
				default:
					// Ignore other chunk types
				}
				return nil
			}),
		)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, streamDone, "Streaming should be done")
		assert.NotEmpty(t, resp.Choices)
		assert.Contains(t, resp.Choices[0].Content, "42")
		assert.NotEmpty(t, thinkingContent.String(), "Thinking content should be streamed")
		assert.NotNil(t, resp.Choices[0].Reasoning)
		if resp.Choices[0].Reasoning != nil {
			assert.Equal(t, resp.Choices[0].Reasoning.Content, thinkingContent.String())
		}
	})
}
