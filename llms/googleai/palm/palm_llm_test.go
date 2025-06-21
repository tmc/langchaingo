package palm

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/llms"
)

func newPalmTestLLM(t *testing.T) *LLM {
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

	// Scrub auth headers
	rr.ScrubReq(func(req *http.Request) error {
		if auth := req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", "Bearer test-token")
		}
		return nil
	})

	// Set test credentials
	os.Setenv("GOOGLE_CLOUD_PROJECT", "test-project")
	os.Setenv("GOOGLE_CLOUD_LOCATION", "test-location")

	llm, err := New(WithHTTPClient(rr.Client()))
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

func TestPaLMCall(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

	output, err := llm.Call(context.Background(), "What is the capital of France?")
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Paris")
}

func TestPaLMGenerateContent(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a joke about programming"),
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
	assert.NotEmpty(t, resp.Choices[0].Content)
}

func TestPaLMCreateEmbedding(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

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

func TestPaLMWithOptions(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Count from 1 to 5"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithMaxTokens(100),
		llms.WithTemperature(0.2),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)
}

func TestPaLMErrorHandling(t *testing.T) {
	t.Parallel()

	// Test missing project ID
	_, err := New(WithLocation("us-central1"))
	assert.Error(t, err)
	assert.Equal(t, ErrMissingProjectID, err)

	// Test missing location
	_, err = New(WithProjectID("test-project"))
	assert.Error(t, err)
	assert.Equal(t, ErrMissingLocation, err)
}

func TestPaLMMultipleTexts(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

	// Test with empty input
	_, err := llm.CreateEmbedding(context.Background(), []string{})
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyResponse, err)

	// Test with multiple texts
	texts := []string{
		"The quick brown fox",
		"jumps over the lazy dog",
		"Machine learning is fascinating",
	}
	embeddings, err := llm.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)
	assert.Len(t, embeddings, 3)

	// Each embedding should be different (different texts)
	assert.NotEqual(t, embeddings[0], embeddings[1])
	assert.NotEqual(t, embeddings[1], embeddings[2])
}

func TestPaLMWithStopWords(t *testing.T) {
	t.Parallel()

	llm := newPalmTestLLM(t)

	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Count from 1 to 10"),
			},
		},
	}

	resp, err := llm.GenerateContent(
		context.Background(),
		content,
		llms.WithStopWords([]string{"5"}),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Choices)

	// Should stop at or before "5"
	output := resp.Choices[0].Content
	assert.NotContains(t, output, "6")
	assert.NotContains(t, output, "7")
}
