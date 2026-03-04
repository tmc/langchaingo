package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	// Set up httprr for recording/replaying HTTP interactions
	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Skip if no recording exists and we're not recording
	if !rr.Recording() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	// Scrub dynamic headers from requests to ensure consistent replay
	rr.ScrubReq(func(req *http.Request) error {
		// Remove Authorization header if present (from OLLAMA_API_KEY env)
		req.Header.Del("Authorization")
		return nil
	})

	// Default model for testing
	ollamaModel := "gemma3:1b"
	if envModel := os.Getenv("OLLAMA_TEST_MODEL"); envModel != "" {
		ollamaModel = envModel
	}

	// Default to localhost
	serverURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		serverURL = envURL
	}

	// Always add server URL and HTTP client
	opts = append([]Option{
		WithServerURL(serverURL),
		WithHTTPClient(rr.Client()),
		WithModel(ollamaModel),
	}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

// newEmbeddingTestClient creates a test client configured for embedding operations
func newEmbeddingTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	// Default embedding model
	embeddingModel := "nomic-embed-text"
	if envModel := os.Getenv("OLLAMA_EMBEDDING_MODEL"); envModel != "" {
		embeddingModel = envModel
	}

	// Use the embedding model by default
	opts = append([]Option{WithModel(embeddingModel)}, opts...)

	return newTestClient(t, opts...)
}

// newStreamingTestClient creates a test client optimized for streaming operations
// It bypasses httprr during replay to avoid chunked encoding issues
func newStreamingTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	// Set up httprr for recording/replaying HTTP interactions
	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Skip if no recording exists and we're not recording
	if !rr.Recording() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
	}

	// Scrub dynamic headers from requests
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Authorization")
		return nil
	})

	// Force using gemma3:1b for better performance
	ollamaModel := "gemma3:1b"

	// Default to localhost
	serverURL := "http://localhost:11434"
	if envURL := os.Getenv("OLLAMA_HOST"); envURL != "" && rr.Recording() {
		serverURL = envURL
	}

	// When recording, use direct HTTP client to avoid httprr interference with streaming
	opts = append([]Option{
		WithServerURL(serverURL),
		WithHTTPClient(rr.Client()),
		WithModel(ollamaModel),
	}, opts...)

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestGenerateContent(t *testing.T) {
	ctx := t.Context()

	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestToolCall(t *testing.T) {
	llm := newTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Which date do we have today?"},
	}
	content := []llms.MessageContent{ //nolint:prealloc
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}
	toolOption := llms.WithTools([]llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getTime",
			Description: "Get the current time.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			},
			Strict: true,
		},
	}})

	resp, err := llm.GenerateContent(t.Context(), content, toolOption)
	require.NoError(t, err)

	require.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	require.NotEmpty(t, c1.ToolCalls)
	t1 := c1.ToolCalls[0]
	require.Equal(t, "getTime", t1.FunctionCall.Name)

	content = append(content, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{t1},
	}, llms.MessageContent{
		Role: llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{
			llms.ToolCallResponse{
				ToolCallID: t1.ID,
				Name:       t1.FunctionCall.Name,
				Content:    "2010-08-13 20:15:00.033067589 +0100 CET m=+32.849928139",
			},
		},
	})

	resp, err = llm.GenerateContent(t.Context(), content, toolOption)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)
	c1 = resp.Choices[0]
	assert.Regexp(t, "2010", c1.Content)
	assert.Regexp(t, "13", c1.Content)
}

func TestWithFormat(t *testing.T) {
	ctx := t.Context()

	llm := newTestClient(t, WithFormat("json"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile? Respond with JSON containing the answer."},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))

	// check whether we got *any* kind of JSON object.
	var result map[string]any
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)
	// The JSON should contain some information about feet or the answer
	assert.NotEmpty(t, result)
}

func TestWithStreaming(t *testing.T) {
	ctx := t.Context()

	// Use streaming-optimized client that avoids httprr interference
	llm := newStreamingTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var (
		sb         strings.Builder
		streamDone bool
	)
	resp, err := llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			switch chunk.Type { //nolint:exhaustive
			case streaming.ChunkTypeText:
				sb.WriteString(chunk.Content)
			case streaming.ChunkTypeDone:
				streamDone = true
			default:
				// skip other chunks
			}
			return nil
		}))

	require.NoError(t, err)

	assert.True(t, streamDone)
	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
}

func TestWithKeepAlive(t *testing.T) {
	ctx := t.Context()

	llm := newTestClient(t, WithKeepAlive("1m"))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "How many feet are in a nautical mile?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithPullModel(t *testing.T) {
	ctx := t.Context()

	// This test verifies the WithPullModel option works correctly.
	// It uses a model that's likely already available locally (gemma3:1b)
	// to avoid expensive downloads during regular test runs.

	// Use newTestClient to get httprr support
	llm := newTestClient(t, WithPullModel())

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Say hello"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// The model should be pulled automatically before generating content
	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.NotEmpty(t, c1.Content)
}

func TestCreateEmbedding(t *testing.T) {
	ctx := t.Context()

	// Use the embedding-specific test client
	llm := newEmbeddingTestClient(t)

	// Test single embedding
	embeddings, err := llm.CreateEmbedding(ctx, []string{"Hello, world!"})

	// Skip if the model is not found
	if err != nil && strings.Contains(err.Error(), "model") && strings.Contains(err.Error(), "not found") {
		t.Skipf("Embedding model not found: %v. Try running 'ollama pull nomic-embed-text' first", err)
	}

	require.NoError(t, err)
	assert.Len(t, embeddings, 1)
	assert.NotEmpty(t, embeddings[0])

	// Verify embedding has correct dimension (should be > 0)
	assert.Greater(t, len(embeddings[0]), 0, "Embedding should have non-zero dimensions")

	// Test multiple embeddings
	texts := []string{
		"The quick brown fox jumps over the lazy dog",
		"Machine learning is a subset of artificial intelligence",
		"Ollama makes it easy to run large language models locally",
	}
	embeddings, err = llm.CreateEmbedding(ctx, texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, len(texts))
	for i, emb := range embeddings {
		assert.NotEmpty(t, emb, "Embedding %d should not be empty", i)
		assert.Greater(t, len(emb), 0, "Embedding %d should have non-zero dimensions", i)
	}
}

func TestWithPullTimeout(t *testing.T) {
	ctx := t.Context()

	if testing.Short() {
		t.Skip("Skipping pull timeout test in short mode")
	}

	// This test only works in recording mode (timeout behavior cannot be replayed)
	// Skip if httprr file doesn't exist
	httprr.SkipIfNoCredentialsAndRecordingMissing(t)

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	if !rr.Recording() {
		t.Skip("Skipping pull timeout test when not recording (timeout behavior cannot be replayed)")
	}

	// Scrub dynamic headers
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Authorization")
		return nil
	})

	// Use a very short timeout that should fail for any real model pull
	llm := newTestClient(t,
		WithModel("llama2:70b"), // Large model that would take time to download
		WithPullModel(),
		WithPullTimeout(50*time.Millisecond), // Extremely short timeout
	)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Say hello"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// This should fail with a timeout error
	_, err := llm.GenerateContent(ctx, content)

	if err == nil {
		t.Fatal("Expected error due to pull timeout, but got none")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") {
		t.Fatalf("Expected timeout error, got: %v", err)
	}
}
