package anthropic

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

// newTestClientWithHTTPRR creates a test Anthropic client that uses httprr for HTTP recording/replay.
func newTestClientWithHTTPRR(t *testing.T, recordingsDir string, opts ...Option) llms.Model {
	t.Helper()
	
	// Create httprr helper
	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	t.Cleanup(helper.Cleanup)
	
	// Get API key (use placeholder for replay mode)
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// In replay mode, we can use a placeholder key
		if os.Getenv("HTTPRR_MODE") == "replay" {
			apiKey = "test-api-key-for-replay"
		} else {
			t.Skip("ANTHROPIC_API_KEY not set and not in replay mode")
			return nil
		}
	}
	
	// Create client with httprr transport
	llm, err := helper.NewAnthropicClient(apiKey, opts...)
	require.NoError(t, err)
	
	return llm
}

func TestAnthropicBasicCompletion_WithHTTPRR(t *testing.T) {
	t.Parallel()
	
	recordingsDir := filepath.Join("testdata", "anthropic_basic_completion")
	llm := newTestClientWithHTTPRR(t, recordingsDir)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("What is the capital of France?")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Contains(t, strings.ToLower(c1.Content), "paris")
}

func TestAnthropicMultiTurn_WithHTTPRR(t *testing.T) {
	t.Parallel()
	
	recordingsDir := filepath.Join("testdata", "anthropic_multi_turn")
	llm := newTestClientWithHTTPRR(t, recordingsDir)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("I'm thinking of a number between 1 and 10")},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Great! I'd love to try guessing it. Is it 7?")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("No, it's lower than that")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	// Should contain a lower number or acknowledgment
	assert.NotEmpty(t, c1.Content)
}

func TestAnthropicWithMaxTokens_WithHTTPRR(t *testing.T) {
	t.Parallel()
	
	recordingsDir := filepath.Join("testdata", "anthropic_max_tokens")
	llm := newTestClientWithHTTPRR(t, recordingsDir)

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Write a very long story about a robot")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithMaxTokens(50))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Contains(t, strings.ToLower(c1.Content), "robot")
	// Should be truncated due to max tokens
	assert.LessOrEqual(t, len(strings.Split(c1.Content, " ")), 100) // rough token count
}

// TestAnthropicHTTPRRFunctionality tests that httprr is working correctly with Anthropic.
func TestAnthropicHTTPRRFunctionality(t *testing.T) {
	recordingsDir := filepath.Join("testdata", "anthropic_httprr_functionality")
	
	// Create httprr helper directly to test its functionality
	helper := httprr.NewLLMTestHelper(t, recordingsDir)
	defer helper.Cleanup()

	// Get API key
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		if os.Getenv("HTTPRR_MODE") == "replay" {
			apiKey = "test-api-key-for-replay"
		} else {
			t.Skip("ANTHROPIC_API_KEY not set and not in replay mode")
		}
	}

	// Create client
	llm, err := helper.NewAnthropicClient(apiKey)
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
	helper.AssertURLCalled("api.anthropic.com")
	
	// Get URLs to verify the right endpoint was called
	urls := helper.GetRequestURLs()
	assert.Len(t, urls, 1)
	assert.Contains(t, urls[0], "anthropic.com")
	
	// Test finding responses
	resp, body, err := helper.FindResponse("messages")
	if err == nil {
		assert.Equal(t, 200, resp.StatusCode)
		assert.NotEmpty(t, body)
	}
}

// ExampleAnthropicWithHTTPRR demonstrates using httprr with Anthropic clients.
func ExampleAnthropicWithHTTPRR() {
	// This example would typically be in a test function
	recordingsDir := "testdata/anthropic_example"
	
	helper := httprr.NewLLMTestHelper(&testing.T{}, recordingsDir)
	defer helper.Cleanup()

	// Create Anthropic client with httprr
	client, err := helper.NewAnthropicClient("your-api-key")
	if err != nil {
		panic(err)
	}

	// Make API call - will be recorded/replayed
	response, err := client.GenerateContent(context.Background(), []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Hello Claude!")},
		},
	})

	if err != nil {
		panic(err)
	}

	// Verify response
	if len(response.Choices) > 0 {
		println("Response:", response.Choices[0].Content)
	}

	// Verify HTTP interactions
	helper.AssertRequestCount(1)
	helper.AssertURLCalled("api.anthropic.com")
}