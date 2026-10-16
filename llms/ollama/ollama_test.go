package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

func newTestClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	// Set up httprr for recording/replaying HTTP interactions
	rr := httprr.OpenForTest(t, http.DefaultTransport)

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

	// Skip if no recording exists and we're not recording
	if !rr.Recording() {
		httprr.SkipIfNoCredentialsAndRecordingMissing(t)
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

func TestGenerateContent(t *testing.T) {
	ctx := context.Background()

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

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
}

func TestWithFormat(t *testing.T) {
	ctx := context.Background()

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

	rsp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]

	// check whether we got *any* kind of JSON object.
	var result map[string]any
	err = json.Unmarshal([]byte(c1.Content), &result)
	require.NoError(t, err)
	// The JSON should contain some information about feet or the answer
	assert.NotEmpty(t, result)
}

func TestWithStreaming(t *testing.T) {
	ctx := context.Background()

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

	var sb strings.Builder
	rsp, err := llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "feet", strings.ToLower(c1.Content))
	assert.Regexp(t, "feet", strings.ToLower(sb.String()))
}

func TestWithKeepAlive(t *testing.T) {
	ctx := context.Background()

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

	// Note: gemma3:1b doesn't support embeddings
	// Use TestCreateEmbedding for embedding tests
}

func TestWithThink(t *testing.T) {
	ctx := context.Background()

	// Test that WithThink option correctly sets the think parameter
	llm := newTestClient(t, WithThink(true))

	parts := []llms.ContentPart{
		llms.TextContent{Text: "What is 2+2? Explain your reasoning step by step."},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// The request should include think:true in options
	resp, err := llm.GenerateContent(ctx, content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	// The response should contain the answer
	assert.Contains(t, strings.ToLower(c1.Content), "4")
}

func TestWithPullModel(t *testing.T) {
	ctx := context.Background()

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
	ctx := context.Background()

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
	}
}

func TestWithTruncate(t *testing.T) {
	ctx := context.Background()

	// Generate a long prompt that exceeds a small context window
	longPrompt := strings.Repeat("This is a test sentence to fill up the context window. ", 100)
	longPrompt += "What is 2+2?"

	// Use a very small context window with truncate enabled (default behavior)
	// The prompt should be silently truncated and still produce a response
	llm := newTestClient(t, WithTruncate(true), WithRunnerNumCtx(128))

	parts := []llms.ContentPart{
		llms.TextContent{Text: longPrompt},
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
}

func TestWithTruncateFalse(t *testing.T) {
	ctx := context.Background()

	// Generate a long prompt that exceeds a small context window
	longPrompt := strings.Repeat("This is a test sentence to fill up the context window. ", 100)
	longPrompt += "What is 2+2?"

	// Use a very small context window with truncate disabled
	// This should return an error when the context is exceeded
	llm := newTestClient(t, WithTruncate(false), WithRunnerNumCtx(128))

	parts := []llms.ContentPart{
		llms.TextContent{Text: longPrompt},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	_, err := llm.GenerateContent(ctx, content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input length exceeds the context length")
}

func TestWithPullTimeout(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("Skipping pull timeout test in short mode")
	}

	// Check if we're recording - timeout tests don't work with replay
	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	if rr.Replaying() {
		t.Skip("Skipping pull timeout test when not recording (timeout behavior cannot be replayed)")
	}

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
