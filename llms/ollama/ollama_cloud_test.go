package ollama

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

// removeTimestampTransport removes 'ts' query parameter before passing to httprr.
// This is needed because Ollama adds dynamic timestamp for cloud auth.
type removeTimestampTransport struct {
	base http.RoundTripper
}

func (t *removeTimestampTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying original
	clonedReq := req.Clone(req.Context())

	// Remove ts query parameter
	q := clonedReq.URL.Query()
	q.Del("ts")
	clonedReq.URL.RawQuery = q.Encode()

	return t.base.RoundTrip(clonedReq)
}

// newCloudTestClient creates a test client configured for Ollama Cloud
func newCloudTestClient(t *testing.T) *LLM {
	t.Helper()

	// Check for required credentials and skip if not available
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OLLAMA_API_KEY")

	// Set up httprr for recording/replaying HTTP interactions
	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Scrub dynamic headers
	rr.ScrubReq(func(req *http.Request) error {
		// Remove Authorization header to avoid leaking API key in recordings
		req.Header.Del("Authorization")
		return nil
	})

	// Wrap httprr client transport to remove ts parameter
	// This ensures consistent httprr matching across recordings and replays
	baseClient := rr.Client()
	wrappedClient := &http.Client{
		Transport: &removeTimestampTransport{
			base: baseClient.Transport,
		},
		Timeout:       baseClient.Timeout,
		CheckRedirect: baseClient.CheckRedirect,
		Jar:           baseClient.Jar,
	}

	// Use Ollama Cloud URL
	serverURL := CloudURL
	if rr.Recording() {
		if envURL := os.Getenv("OLLAMA_CLOUD_URL"); envURL != "" {
			serverURL = envURL
		}
	}

	// Get API key from environment
	apiKey := os.Getenv("OLLAMA_API_KEY")

	// Default cloud model for testing
	cloudModel := "gpt-oss:120b"
	if envModel := os.Getenv("OLLAMA_CLOUD_MODEL"); envModel != "" {
		cloudModel = envModel
	}

	opts := []Option{
		WithServerURL(serverURL),
		WithAPIKey(apiKey),
		WithHTTPClient(wrappedClient),
		WithModel(cloudModel),
	}

	c, err := New(opts...)
	require.NoError(t, err)
	return c
}

func TestCloudGenerateContent(t *testing.T) {
	ctx := context.Background()

	llm := newCloudTestClient(t)

	testCases := []struct {
		name     string
		prompt   string
		validate func(t *testing.T, content string)
	}{
		{
			name:   "simple question",
			prompt: "What is the capital of France?",
			validate: func(t *testing.T, content string) {
				assert.Regexp(t, "(?i)paris", content)
			},
		},
		{
			name:   "math question",
			prompt: "What is 15 multiplied by 7?",
			validate: func(t *testing.T, content string) {
				assert.Regexp(t, "105", content)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parts := []llms.ContentPart{
				llms.TextContent{Text: tc.prompt},
			}
			content := []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: parts,
				},
			}

			resp, err := llm.GenerateContent(ctx, content)
			require.NoError(t, err)
			require.NotEmpty(t, resp.Choices)

			c1 := resp.Choices[0]
			tc.validate(t, c1.Content)
		})
	}
}

func TestCloudStreaming(t *testing.T) {
	ctx := context.Background()

	llm := newCloudTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "Count from 1 to 5"},
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
	assert.True(t, streamDone, "Stream should be marked as done")
	assert.NotEmpty(t, resp.Choices)

	c1 := resp.Choices[0]
	streamedContent := sb.String()

	assert.NotEmpty(t, c1.Content)
	assert.NotEmpty(t, streamedContent)
	assert.Equal(t, c1.Content, streamedContent, "Streamed content should match final response")
}

func TestCloudToolCall(t *testing.T) {
	ctx := context.Background()

	llm := newCloudTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "What's the weather like in San Francisco?"},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	toolOption := llms.WithTools([]llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "getWeather",
			Description: "Get the current weather for a location.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The city and state, e.g. San Francisco, CA",
					},
				},
				"required": []string{"location"},
			},
		},
	}})

	resp, err := llm.GenerateContent(ctx, content, toolOption)
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	c1 := resp.Choices[0]

	// The model should either call the tool or provide a text response
	if len(c1.ToolCalls) > 0 {
		t1 := c1.ToolCalls[0]
		assert.Equal(t, "getWeather", t1.FunctionCall.Name)

		// Continue conversation with tool response
		content = append(content, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{t1},
		}, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: t1.ID,
					Name:       t1.FunctionCall.Name,
					Content:    `{"temperature": 72, "condition": "sunny"}`,
				},
			},
		})

		resp, err = llm.GenerateContent(ctx, content, toolOption)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Choices)

		c2 := resp.Choices[0]
		assert.NotEmpty(t, c2.Content)
		assert.Regexp(t, "(?i)(72|sunny|weather)", c2.Content)
	} else {
		// Model chose to respond without calling tool
		assert.NotEmpty(t, c1.Content)
	}
}

func TestCloudJSONMode(t *testing.T) {
	ctx := context.Background()

	llm := newCloudTestClient(t)

	parts := []llms.ContentPart{
		llms.TextContent{Text: "List three colors in JSON format with a 'colors' array."},
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(ctx, content, llms.WithJSONMode())
	require.NoError(t, err)
	require.NotEmpty(t, resp.Choices)

	c1 := resp.Choices[0]
	assert.NotEmpty(t, c1.Content)

	// Verify the response contains JSON structure
	responseText := strings.TrimSpace(c1.Content)
	assert.Contains(t, responseText, "{")
	assert.Contains(t, responseText, "}")
	assert.Contains(t, responseText, "colors")
}

func TestCloudWithOptions(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name        string
		opts        []llms.CallOption
		prompt      string
		minTokens   int
		expectShort bool
	}{
		{
			name: "temperature variation",
			opts: []llms.CallOption{
				llms.WithTemperature(0.1), // Very low temperature for deterministic output
			},
			prompt:    "Say hello",
			minTokens: 1,
		},
		{
			name: "seed for reproducibility",
			opts: []llms.CallOption{
				llms.WithSeed(42),
				llms.WithTemperature(0.7),
			},
			prompt:    "Pick a random number between 1 and 10",
			minTokens: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm := newCloudTestClient(t)

			parts := []llms.ContentPart{
				llms.TextContent{Text: tc.prompt},
			}
			content := []llms.MessageContent{
				{
					Role:  llms.ChatMessageTypeHuman,
					Parts: parts,
				},
			}

			resp, err := llm.GenerateContent(ctx, content, tc.opts...)
			require.NoError(t, err)
			require.NotEmpty(t, resp.Choices)

			c1 := resp.Choices[0]
			assert.NotEmpty(t, c1.Content)

			if tc.minTokens > 0 {
				tokenCount := len(strings.Fields(c1.Content))
				assert.GreaterOrEqual(t, tokenCount, tc.minTokens)
			}
		})
	}
}
