package openai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httprr"
	"github.com/tmc/langchaingo/llms"
)

// newTestClientWithHTTPRR creates a test client that uses httprr for HTTP recording/replay.
func newTestClientWithHTTPRR(t *testing.T, recordingsDir string, opts ...Option) llms.Model {
	t.Helper()
	
	// Create httprr helper
	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	t.Cleanup(helper.Cleanup)
	
	// Get API key (use placeholder for replay mode)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// In replay mode, we can use a placeholder key
		if os.Getenv("HTTPRR_MODE") == "replay" {
			apiKey = "test-api-key-for-replay"
		} else {
			t.Skip("OPENAI_API_KEY not set and not in replay mode")
			return nil
		}
	}
	
	// Create client with httprr transport
	llm, err := helper.NewOpenAIClientWithToken(apiKey, opts...)
	require.NoError(t, err)
	
	return llm
}

func TestMultiContentText_WithHTTPRR(t *testing.T) {
	t.Parallel()
	
	recordingsDir := filepath.Join("testdata", "openai_multicontent_text")
	llm := newTestClientWithHTTPRR(t, recordingsDir)

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
}

func TestMultiContentTextChatSequence_WithHTTPRR(t *testing.T) {
	t.Parallel()
	
	recordingsDir := filepath.Join("testdata", "openai_multicontent_chat_sequence")
	llm := newTestClientWithHTTPRR(t, recordingsDir)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name some countries")},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Spain and Lesotho")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Which if these is larger?")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "spain.*larger", strings.ToLower(c1.Content))
}

func TestMultiContentImage_WithHTTPRR(t *testing.T) {
	t.Parallel()

	recordingsDir := filepath.Join("testdata", "openai_multicontent_image")
	llm := newTestClientWithHTTPRR(t, recordingsDir, WithModel("gpt-4o"))

	parts := []llms.ContentPart{
		llms.ImageURLPart("https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"),
		llms.TextPart("describe this image in detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithMaxTokens(300))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "parrot", strings.ToLower(c1.Content))
}

// TestHTTPRRFunctionality tests that httprr is actually working as expected.
func TestHTTPRRFunctionality(t *testing.T) {
	recordingsDir := filepath.Join("testdata", "openai_httprr_functionality")
	
	// Create httprr helper directly to test its functionality
	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Get API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		if os.Getenv("HTTPRR_MODE") == "replay" {
			apiKey = "test-api-key-for-replay"
		} else {
			t.Skip("OPENAI_API_KEY not set and not in replay mode")
		}
	}

	// Create client
	llm, err := helper.NewOpenAIClientWithToken(apiKey)
	require.NoError(t, err)

	// Make a simple request
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Say hello")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)
	require.NotEmpty(t, rsp.Choices)

	// Test httprr assertions
	helper.AssertRequestCount(1)
	helper.AssertURLCalled("api.openai.com")
	
	// Get URLs to verify the right endpoint was called
	urls := helper.GetRequestURLs()
	assert.Len(t, urls, 1)
	assert.Contains(t, urls[0], "openai.com")
	
	// Test finding responses
	resp, body, err := helper.FindResponse("completions")
	if err == nil {
		assert.Equal(t, 200, resp.StatusCode)
		assert.NotEmpty(t, body)
	}
}

// BenchmarkMultiContentWithHTTPRR benchmarks the performance with httprr.
func BenchmarkMultiContentWithHTTPRR(b *testing.B) {
	// Force replay mode for benchmarks to ensure consistent results
	origMode := os.Getenv("HTTPRR_MODE")
	os.Setenv("HTTPRR_MODE", "replay")
	defer func() {
		if origMode == "" {
			os.Unsetenv("HTTPRR_MODE")
		} else {
			os.Setenv("HTTPRR_MODE", origMode)
		}
	}()

	recordingsDir := filepath.Join("testdata", "openai_benchmark")
	helper := httprr.NewLLMTestHelper(&testing.T{}, recordingsDir)
	defer helper.Cleanup()

	llm, err := helper.NewOpenAIClientWithToken("test-api-key")
	if err != nil {
		b.Fatal(err)
	}

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Hello")},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := llm.GenerateContent(context.Background(), content)
		if err != nil {
			b.Fatal(err)
		}
	}
}