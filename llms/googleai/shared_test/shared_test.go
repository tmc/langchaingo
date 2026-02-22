package shared_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/embeddings"
	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
	"github.com/vxcontrol/langchaingo/llms/googleai/vertex"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newGoogleAIClient(t *testing.T, opts ...googleai.Option) *googleai.GoogleAI {
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

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Scrub API key for security in recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("key") != "" {
			q.Set("key", "test-api-key")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Configure client with httprr and test credentials
	opts = append(opts,
		googleai.WithRest(),
		googleai.WithAPIKey("test-api-key"),
		googleai.WithHTTPClient(rr.Client()),
	)

	llm, err := googleai.New(t.Context(), opts...)
	require.NoError(t, err)

	return llm
}

func newVertexClient(t *testing.T, opts ...googleai.Option) *vertex.Vertex {
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

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Configure client with httprr and test credentials
	opts = append(opts,
		googleai.WithHTTPClient(rr.Client()),
		googleai.WithCloudProject("test-project"),
		googleai.WithCloudLocation("us-central1"),
	)

	llm, err := vertex.New(t.Context(), opts...)
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

// funcName obtains the name of the given function value, without a package
// prefix.
func funcName(f any) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// testConfigs is a list of all test functions in this file to run with both
// client types, and their client configurations.
type testConfig struct {
	testFunc func(*testing.T, llms.Model)
	opts     []googleai.Option
}

func getTestConfigs() []testConfig {
	return []testConfig{
		{testMultiContentText, nil},
		{testGenerateFromSinglePrompt, nil},
		{testMultiContentTextChatSequence, nil},
		{testMultiContentWithSystemMessage, nil},
		{testMultiContentImageLink, nil},
		{testMultiContentImageBinary, nil},
		{testEmbeddings, nil},
		{testCandidateCountSetting, nil},
		{testMaxTokensSetting, nil},
		{testTools, nil},
		{testToolsWithInterfaceRequired, nil},
		{
			testMultiContentText,
			[]googleai.Option{googleai.WithHarmThreshold(googleai.HarmBlockMediumAndAbove)},
		},
		{
			testMultiContentTextUsingTextParts,
			[]googleai.Option{googleai.WithHarmThreshold(googleai.HarmBlockMediumAndAbove)},
		},
		{testWithStreaming, nil},
		{testWithHTTPClient, getHTTPTestClientOptions()},
	}
}

func TestGoogleAIShared(t *testing.T) {
	t.Parallel()

	testConfigs := getTestConfigs()
	for idx := range testConfigs {
		c := testConfigs[idx]
		t.Run(fmt.Sprintf("%s-googleai", funcName(c.testFunc)), func(t *testing.T) {
			t.Parallel()

			c.testFunc(t, newGoogleAIClient(t, c.opts...))
		})
	}
}

func TestVertexShared(t *testing.T) {
	t.Parallel()

	testConfigs := getTestConfigs()
	for idx := range testConfigs {
		c := testConfigs[idx]
		t.Run(fmt.Sprintf("%s-vertex", funcName(c.testFunc)), func(t *testing.T) {
			t.Parallel()

			c.testFunc(t, newVertexClient(t, c.opts...))
		})
	}
}

func testMultiContentText(t *testing.T, llm llms.Model) {
	t.Helper()

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

	resp, err := llm.GenerateContent(t.Context(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "(?i)dog|carnivo|canid|canine", c1.Content)
	assert.Contains(t, c1.GenerationInfo, "output_tokens")
	assert.NotZero(t, c1.GenerationInfo["output_tokens"])
}

func testMultiContentTextUsingTextParts(t *testing.T, llm llms.Model) {
	t.Helper()

	content := llms.TextParts(
		llms.ChatMessageTypeHuman,
		"I'm a pomeranian",
		"What kind of mammal am I?",
	)

	resp, err := llm.GenerateContent(t.Context(), []llms.MessageContent{content})
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "(?i)dog|canid|canine", c1.Content)
}

func testGenerateFromSinglePrompt(t *testing.T, llm llms.Model) {
	t.Helper()

	prompt := "name all the planets in the solar system"
	resp, err := llms.GenerateFromSinglePrompt(t.Context(), llm, prompt)
	require.NoError(t, err)

	assert.Regexp(t, "(?i)jupiter", resp)
}

func testMultiContentTextChatSequence(t *testing.T, llm llms.Model) {
	t.Helper()

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

	resp, err := llm.GenerateContent(t.Context(), content, llms.WithModel("gemini-1.5-flash"))
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Regexp(t, "(?i)spain.*larger", c1.Content)
}

func testMultiContentWithSystemMessage(t *testing.T, llm llms.Model) {
	t.Helper()

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart("You are a Spanish teacher; answer in Spanish")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name the 5 most common fruits")},
		},
	}

	resp, err := llm.GenerateContent(t.Context(), content, llms.WithModel("gemini-1.5-flash"))
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	checkMatch(t, c1.Content, "(manzana|naranja)")
}

func testMultiContentImageLink(t *testing.T, llm llms.Model) {
	t.Helper()

	parts := []llms.ContentPart{
		llms.ImageURLPart(
			"https://github.com/vxcontrol/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true",
		),
		llms.TextPart("describe this image in detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithModel("gemini-2.0-flash"),
	)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	checkMatch(t, c1.Content, "parrot")
}

func testMultiContentImageBinary(t *testing.T, llm llms.Model) {
	t.Helper()

	b, err := os.ReadFile(filepath.Join("testdata", "parrot-icon.png"))
	if err != nil {
		t.Fatal(err)
	}

	parts := []llms.ContentPart{
		llms.BinaryPart("image/png", b),
		llms.TextPart("what does this image show? please use detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithModel("gemini-2.0-flash"),
	)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	checkMatch(t, c1.Content, "parrot")
}

func testEmbeddings(t *testing.T, llm llms.Model) {
	t.Helper()

	texts := []string{"foo", "parrot", "foo"}
	emb := llm.(embeddings.EmbedderClient)
	res, err := emb.CreateEmbedding(t.Context(), texts)
	require.NoError(t, err)

	assert.Equal(t, len(texts), len(res))
	assert.NotEmpty(t, res[0])
	assert.NotEmpty(t, res[1])
	assert.Equal(t, res[0], res[2])
}

func testCandidateCountSetting(t *testing.T, llm llms.Model) {
	t.Helper()

	parts := []llms.ContentPart{
		llms.TextPart("Name five countries in Africa"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	{
		resp, err := llm.GenerateContent(t.Context(), content,
			llms.WithCandidateCount(1), llms.WithTemperature(1))
		require.NoError(t, err)

		assert.Len(t, resp.Choices, 1)
	}

	// TODO: test multiple candidates when the backend supports it
}

func testWithStreaming(t *testing.T, llm llms.Model) {
	t.Helper()

	content := llms.TextParts(
		llms.ChatMessageTypeHuman,
		"I'm a pomeranian",
		"Tell me more about my taxonomy",
	)

	var (
		sb         strings.Builder
		streamDone bool
	)
	resp, err := llm.GenerateContent(
		t.Context(),
		[]llms.MessageContent{content},
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
	checkMatch(t, c1.Content, "(dog|canid)")
	checkMatch(t, sb.String(), "(dog|canid)")
}

func testTools(t *testing.T, llm llms.Model) {
	t.Helper()

	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
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
		},
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago?"),
	}
	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithTools(availableTools))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)

	c1 := resp.Choices[0]

	// Update chat history with assistant's response, with its tool calls.
	assistantResp := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
	}
	for _, tc := range c1.ToolCalls {
		assistantResp.Parts = append(assistantResp.Parts, tc)
	}
	content = append(content, assistantResp)

	// "Execute" tool calls by calling requested function
	for _, tc := range c1.ToolCalls {
		switch tc.FunctionCall.Name {
		case "getCurrentWeather":
			var args struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(args.Location, "Chicago") {
				toolResponse := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							Name:    tc.FunctionCall.Name,
							Content: "64 and sunny",
						},
					},
				}
				content = append(content, toolResponse)
			}
		default:
			t.Errorf("got unexpected function call: %v", tc.FunctionCall.Name)
		}
	}

	resp, err = llm.GenerateContent(t.Context(), content, llms.WithTools(availableTools))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)

	c1 = resp.Choices[0]
	checkMatch(t, c1.Content, "(64 and sunny|64 degrees)")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "input_tokens")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "output_tokens")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "total_tokens")
	assert.NotZero(t, resp.Choices[0].GenerationInfo["total_tokens"])
}

func testToolsWithInterfaceRequired(t *testing.T, llm llms.Model) {
	t.Helper()

	availableTools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					// json.Unmarshal() may return []interface{} instead of []string
					"required": []interface{}{"location"},
				},
			},
		},
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the weather like in Chicago?"),
	}
	resp, err := llm.GenerateContent(
		t.Context(),
		content,
		llms.WithTools(availableTools))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)

	c1 := resp.Choices[0]
	assert.Contains(t, c1.GenerationInfo, "output_tokens")
	assert.NotZero(t, c1.GenerationInfo["output_tokens"])

	// Update chat history with assistant's response, with its tool calls.
	assistantResp := llms.MessageContent{
		Role: llms.ChatMessageTypeAI,
	}
	for _, tc := range c1.ToolCalls {
		assistantResp.Parts = append(assistantResp.Parts, tc)
	}
	content = append(content, assistantResp)

	// "Execute" tool calls by calling requested function
	for _, tc := range c1.ToolCalls {
		switch tc.FunctionCall.Name {
		case "getCurrentWeather":
			var args struct {
				Location string `json:"location"`
			}
			if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(args.Location, "Chicago") {
				toolResponse := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{
							Name:    tc.FunctionCall.Name,
							Content: "64 and sunny",
						},
					},
				}
				content = append(content, toolResponse)
			}
		default:
			t.Errorf("got unexpected function call: %v", tc.FunctionCall.Name)
		}
	}

	resp, err = llm.GenerateContent(t.Context(), content, llms.WithTools(availableTools))
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Choices)

	c1 = resp.Choices[0]
	checkMatch(t, c1.Content, "(64 and sunny|64 degrees)")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "input_tokens")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "output_tokens")
	assert.Contains(t, resp.Choices[0].GenerationInfo, "total_tokens")
	assert.NotZero(t, resp.Choices[0].GenerationInfo["total_tokens"])
}

func testMaxTokensSetting(t *testing.T, llm llms.Model) {
	t.Helper()

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("Describe my taxonomy, health and care"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// First, try this with a very low MaxTokens setting for such a query; expect
	// a stop reason that max of tokens was reached.
	{
		resp, err := llm.GenerateContent(t.Context(), content, llms.WithMaxTokens(24))
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Choices)
		c1 := resp.Choices[0]
		// TODO: Google genai models are returning "FinishReasonStop" instead of "MaxTokens".
		assert.Regexp(t, "(?i)(MaxTokens|FinishReasonStop)", c1.StopReason)
	}

	// Now, try it again with a much larger MaxTokens setting and expect to
	// finish successfully and generate a response.
	{
		resp, err := llm.GenerateContent(t.Context(), content, llms.WithMaxTokens(2048))
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Choices)
		c1 := resp.Choices[0]
		checkMatch(t, c1.StopReason, "stop")
		checkMatch(t, c1.Content, "(dog|breed|canid|canine)")
	}
}

func testWithHTTPClient(t *testing.T, llm llms.Model) {
	t.Helper()

	resp, err := llm.GenerateContent(
		t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "testing")},
	)
	require.NoError(t, err)
	require.EqualValues(t, "test-ok", resp.Choices[0].Content)
}

func getHTTPTestClientOptions() []googleai.Option {
	client := &http.Client{Transport: &testRequestInterceptor{}}
	return []googleai.Option{googleai.WithRest(), googleai.WithHTTPClient(client)}
}

type testRequestInterceptor struct{}

func (i *testRequestInterceptor) RoundTrip(req *http.Request) (*http.Response, error) {
	defer req.Body.Close()
	content := `{
	"candidates": [{
		"content": {
			"parts": [{"text": "test-ok"}]
		},
		"finishReason": "STOP"
	}],
	"usageMetadata": {
		"promptTokenCount": 7,
		"candidatesTokenCount": 7,
		"totalTokenCount": 14
	}
}`

	resp := &http.Response{
		StatusCode: http.StatusOK, Request: req,
		Body:   io.NopCloser(bytes.NewBufferString(content)),
		Header: http.Header{},
	}
	resp.Header.Set("Content-Type", "application/json")
	return resp, nil
}

// checkMatch is a testing helper that checks `got` for regexp matches vs.
// `wants`. Each of `wants` has to match.
func checkMatch(t *testing.T, got string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		re, err := regexp.Compile("(?i:" + want + ")")
		if err != nil {
			t.Fatal(err)
		}
		if !re.MatchString(got) {
			t.Errorf("\ngot %q\nwanted to match %q", got, want)
		}
	}
}

// mockGoogleAIServer creates a test HTTP server that mimics Google AI API responses.
// It returns a simple mock response similar to what Google AI API would return.
func mockGoogleAIServer(t *testing.T) *http.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Verify this is a POST request
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read and validate request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Basic validation - ensure it's JSON
		var reqData map[string]any
		if err := json.Unmarshal(body, &reqData); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Mock response matching Google AI API format
		response := map[string]any{
			"candidates": []map[string]any{
				{
					"content": map[string]any{
						"parts": []map[string]any{
							{
								"text": "Mock response from custom endpoint",
							},
						},
						"role": "model",
					},
					"finishReason": "STOP",
					"safetyRatings": []map[string]any{
						{
							"category":    "HARM_CATEGORY_HATE_SPEECH",
							"probability": "NEGLIGIBLE",
						},
					},
				},
			},
			"usageMetadata": map[string]any{
				"promptTokenCount":     10,
				"candidatesTokenCount": 8,
				"totalTokenCount":      18,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server
}

// TestEndpointConfiguration demonstrates the current issue with WithEndpoint option.
// This test shows that WithEndpoint does NOT work as expected - requests still go
// to the default Google API endpoint instead of the custom endpoint.
func TestEndpointConfiguration(t *testing.T) {
	t.Parallel()

	// Start mock server
	server := mockGoogleAIServer(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	serverURL := fmt.Sprintf("http://%s", listener.Addr().String())
	t.Logf("Mock server listening on: %s", serverURL)

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.Close()

	t.Run("WithEndpoint_DoesNotWork", func(t *testing.T) {
		// This test demonstrates the PROBLEM: WithEndpoint option is ignored
		// The request will try to go to the real Google API endpoint instead of our mock server

		ctx := context.Background()

		// Try to create client with WithEndpoint pointing to our mock server
		llm, err := googleai.New(ctx,
			googleai.WithRest(),
			googleai.WithAPIKey("test-api-key"),
			googleai.WithEndpoint(serverURL), // This option is IGNORED!
		)
		require.NoError(t, err)

		// Attempt to make a request
		// NOTE: This will FAIL because the request goes to the real Google API endpoint,
		// not to our mock server at serverURL
		content := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart("test prompt")},
			},
		}

		// This call will fail with a network error or authentication error
		// because it's trying to reach the real Google API
		resp, err := llm.GenerateContent(ctx, content)

		// Expected behavior: Should successfully call our mock server
		// Actual behavior: Fails because WithEndpoint is ignored
		if err != nil {
			t.Logf("EXPECTED FAILURE: WithEndpoint does not work. Error: %v", err)
			t.Logf("ROOT CAUSE: In new.go, the ClientOptions field (which contains WithEndpoint) is never used when creating genai.Client")
			return
		}

		// If we somehow got here, verify the response
		if resp != nil && len(resp.Choices) > 0 {
			t.Logf("Response content: %s", resp.Choices[0].Content)
			// This would only work if WithEndpoint was properly implemented
			assert.Contains(t, resp.Choices[0].Content, "Mock response from custom endpoint")
		}
	})

	t.Run("WithHTTPClient_AndCustomTransport_Works", func(t *testing.T) {
		// This test demonstrates the WORKAROUND: Using WithHTTPClient with custom transport
		// that rewrites URLs works correctly

		ctx := context.Background()

		// Create custom transport that rewrites URLs to our mock server
		customTransport := &httputil.ApiKeyTransport{
			Transport: http.DefaultTransport,
			APIKey:    "test-api-key",
			BaseURL:   serverURL, // This will rewrite all requests to our mock server
		}

		customClient := &http.Client{
			Transport: customTransport,
		}

		// Create client with custom HTTP client
		llm, err := googleai.New(ctx,
			googleai.WithRest(),
			googleai.WithAPIKey("test-api-key"),
			googleai.WithHTTPClient(customClient), // Use custom client with URL rewriting
		)
		require.NoError(t, err)

		// Make a request
		content := []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart("test prompt")},
			},
		}

		resp, err := llm.GenerateContent(ctx, content)

		// This SHOULD work because our custom transport rewrites the URL
		require.NoError(t, err, "WithHTTPClient with custom transport should work")
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.Choices)

		// Verify we got the mock response
		assert.Contains(t, resp.Choices[0].Content, "Mock response from custom endpoint",
			"Should receive response from mock server")

		t.Logf("SUCCESS: Custom transport with BaseURL works correctly")
		t.Logf("Response: %s", resp.Choices[0].Content)
	})
}

// TestEndpointConfigurationAnalysis provides detailed analysis of the issue.
func TestEndpointConfigurationAnalysis(t *testing.T) {
	t.Run("RootCauseAnalysis", func(t *testing.T) {
		analysis := `
ROOT CAUSE ANALYSIS: Why WithEndpoint Does Not Work
====================================================

PROBLEM:
--------
The WithEndpoint option is defined and appears to be available for use,
but it has NO EFFECT when creating a Google AI client.

LOCATION OF THE ISSUE:
---------------------
File: llms/googleai/new.go
Lines: 25-69

DETAILED EXPLANATION:
--------------------
1. In option.go (lines 108-115), WithEndpoint is defined:
   func WithEndpoint(endpoint string) Option {
       return func(opts *Options) {
           opts.ClientOptions = append(opts.ClientOptions, option.WithEndpoint(endpoint))
       }
   }

2. The WithEndpoint option correctly adds option.WithEndpoint(endpoint) to opts.ClientOptions

3. However, in new.go, the New() function creates a genai.Client like this:
   - Line 36: config := &genai.ClientConfig{}
   - Lines 38-60: Various config fields are set (HTTPClient, Backend, Project, etc.)
   - Line 62: client, err := genai.NewClient(ctx, config)

4. THE PROBLEM: The opts.ClientOptions field is NEVER USED!
   The ClientOptions array (which contains the WithEndpoint option) is stored
   in the Options struct but is never passed to genai.NewClient.

5. The new Google genai SDK (google.golang.org/genai) uses ClientConfig struct,
   not the old option.ClientOption pattern. The code keeps ClientOptions for
   backward compatibility but doesn't use them.

EVIDENCE:
---------
Search for "ClientOptions" in new.go shows:
- It's defined in Options struct (option.go:30)
- It's populated by various With* functions (option.go)
- It's NEVER READ OR USED in new.go when creating the client

WHY THE WORKAROUND WORKS:
-------------------------
Using WithHTTPClient with a custom http.RoundTripper that rewrites URLs works because:
1. The HTTPClient IS properly set in ClientConfig (new.go:39-41)
2. The genai.Client respects the custom HTTP client
3. Our custom transport intercepts requests and rewrites URLs before they're sent

SOLUTION:
---------
The ClientOptions field should be removed OR the new.go code should be updated
to properly support endpoint configuration through the genai.ClientConfig.
However, the new genai SDK may not support custom endpoints in the same way,
so the BaseURL approach in ApiKeyTransport is the recommended workaround.
`
		t.Log(analysis)
	})
}
