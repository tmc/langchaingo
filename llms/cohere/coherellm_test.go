package cohere

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/vendasta/langchaingo/internal/httprr"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/cohere/internal/cohereclient"
	"github.com/vendasta/langchaingo/schema"
)

// newClientWithHTTPClient creates a cohere client with a custom HTTP client for testing
func newClientWithHTTPClient(httpClient *http.Client, opts ...Option) (*cohereclient.Client, error) {
	options := &options{
		token:   os.Getenv(tokenEnvVarName),
		baseURL: os.Getenv(baseURLEnvVarName),
		model:   os.Getenv(modelEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return cohereclient.New(options.token, options.baseURL, options.model, cohereclient.WithHTTPClient(httpClient))
}

func TestNew(t *testing.T) {
	// Save original env vars
	origToken := os.Getenv("COHERE_API_KEY")
	origModel := os.Getenv("COHERE_MODEL")
	origBaseURL := os.Getenv("COHERE_BASE_URL")
	defer func() {
		os.Setenv("COHERE_API_KEY", origToken)
		os.Setenv("COHERE_MODEL", origModel)
		os.Setenv("COHERE_BASE_URL", origBaseURL)
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "with token option",
			opts:    []Option{WithToken("test-token")},
			wantErr: false,
		},
		{
			name: "with all options",
			opts: []Option{
				WithToken("test-token"),
				WithModel("command"),
				WithBaseURL("https://api.example.com"),
			},
			wantErr: false,
		},
		{
			name: "from env vars",
			envVars: map[string]string{
				"COHERE_API_KEY":  "test-token",
				"COHERE_MODEL":    "command",
				"COHERE_BASE_URL": "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "missing token",
			envVars: map[string]string{
				"COHERE_API_KEY": "",
			},
			opts:    []Option{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Setenv("COHERE_API_KEY", "")
			os.Setenv("COHERE_MODEL", "")
			os.Setenv("COHERE_BASE_URL", "")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestCall(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &mockHTTPClient{
		response: &http.Response{
			StatusCode: 200,
			Body: newMockBody(`{
				"generations": [{
					"text": "Hello from Cohere!"
				}]
			}`),
		},
	}

	client, err := cohereclient.New("test-token", "", "", cohereclient.WithHTTPClient(mockClient))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	llm := &LLM{client: client}

	response, err := llm.Call(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("Call() error: %v", err)
	}

	if response != "Hello from Cohere!" {
		t.Errorf("expected 'Hello from Cohere!', got %q", response)
	}
}

func TestGenerateContent(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "COHERE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	var opts []Option

	// Use test token when replaying, real token when recording
	if rr.Replaying() {
		opts = append(opts, WithToken("test-api-key"))
	}

	// Create LLM with httprr client - need to pass through to internal client
	client, err := newClientWithHTTPClient(rr.Client(), opts...)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	llm := &LLM{client: client}

	messages := []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello, Cohere!"},
			},
		},
	}

	resp, err := llm.GenerateContent(context.Background(), messages)
	if err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Content == "" {
		t.Error("expected non-empty content")
	}
}

type testCallbackHandler struct {
	generateStartCalled bool
	generateEndCalled   bool
	errorCalled         bool
}

func (h *testCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, messages []llms.MessageContent) {
	h.generateStartCalled = true
}

func (h *testCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, resp *llms.ContentResponse) {
	h.generateEndCalled = true
}

func (h *testCallbackHandler) HandleLLMError(ctx context.Context, err error) {
	h.errorCalled = true
}

func (h *testCallbackHandler) HandleText(ctx context.Context, text string)                      {}
func (h *testCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string)             {}
func (h *testCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any)      {}
func (h *testCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any)       {}
func (h *testCallbackHandler) HandleChainError(ctx context.Context, err error)                  {}
func (h *testCallbackHandler) HandleToolStart(ctx context.Context, input string)                {}
func (h *testCallbackHandler) HandleToolEnd(ctx context.Context, output string)                 {}
func (h *testCallbackHandler) HandleToolError(ctx context.Context, err error)                   {}
func (h *testCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}
func (h *testCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}
func (h *testCallbackHandler) HandleRetrieverStart(ctx context.Context, query string)           {}
func (h *testCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (h *testCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

func TestCallbacksHandler(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "COHERE_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	var opts []Option

	// Use test token when replaying, real token when recording
	if rr.Replaying() {
		opts = append(opts, WithToken("test-api-key"))
	}

	client, err := newClientWithHTTPClient(rr.Client(), opts...)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	handler := &testCallbackHandler{}
	llm := &LLM{
		client:           client,
		CallbacksHandler: handler,
	}

	messages := []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "test"},
			},
		},
	}

	_, err = llm.GenerateContent(context.Background(), messages)
	// The callback handler should always be called for start
	if !handler.generateStartCalled {
		t.Error("Expected HandleLLMGenerateContentStart to be called")
	}

	// For successful requests, end should be called; for errors, error should be called
	if err != nil {
		if !handler.errorCalled {
			t.Error("Expected HandleLLMError to be called on error")
		}
	} else {
		if !handler.generateEndCalled {
			t.Error("Expected HandleLLMGenerateContentEnd to be called on success")
		}
	}
}

// Mock HTTP client for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.response, nil
}

// Helper to create a mock response body
type mockBody struct {
	content []byte
	pos     int
}

func newMockBody(content string) *mockBody {
	return &mockBody{content: []byte(content)}
}

func (m *mockBody) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, nil
	}
	n = copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockBody) Close() error {
	return nil
}
