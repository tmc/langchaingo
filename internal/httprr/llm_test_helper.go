package httprr

import (
	"context"
	"net/http"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// TestWithRecording is a helper function that sets up HTTP recording for LLM tests.
// It creates a temporary test environment with HTTP request/response recording.
func TestWithRecording(t *testing.T, testName string, testFunc func(*testing.T, *http.Client)) {
	client := NewTestClient(testName)
	testFunc(t, client)
}

// NewOpenAIClientWithRecording creates an OpenAI client with httprr recording.
func NewOpenAIClientWithRecording(testName string, opts ...openai.Option) (*openai.LLM, error) {
	client := NewTestClient(testName)
	
	// Add the custom HTTP client to the OpenAI options
	allOpts := append(opts, openai.WithHTTPClient(client))
	
	return openai.New(allOpts...)
}

// LLMTestHelper provides utilities for testing LLM implementations with httprr.
type LLMTestHelper struct {
	recorder  *Recorder
	testName  string
	httpClient *http.Client
}

// NewLLMTestHelper creates a new test helper for LLM testing.
func NewLLMTestHelper(testName string, opts ...TestClientOption) *LLMTestHelper {
	config := &TestClientConfig{
		TestName:      testName,
		RecordingsDir: getRecordingsDir(testName),
		Mode:          getRecordingMode(),
	}
	
	for _, opt := range opts {
		opt(config)
	}
	
	recorder := New(config.RecordingsDir, config.Mode)
	
	return &LLMTestHelper{
		recorder:   recorder,
		testName:   testName,
		httpClient: recorder.Client(),
	}
}

// HTTPClient returns the HTTP client with recording capabilities.
func (h *LLMTestHelper) HTTPClient() *http.Client {
	return h.httpClient
}

// TestOpenAI creates an OpenAI client for testing.
func (h *LLMTestHelper) TestOpenAI(opts ...openai.Option) (*openai.LLM, error) {
	allOpts := append(opts, openai.WithHTTPClient(h.httpClient))
	return openai.New(allOpts...)
}

// Reset clears all recordings (useful for test cleanup).
func (h *LLMTestHelper) Reset(removeFiles bool) error {
	return h.recorder.Reset(removeFiles)
}

// GenerateContentWithRecording is a helper that wraps LLM.GenerateContent with recording.
func (h *LLMTestHelper) GenerateContentWithRecording(
	ctx context.Context,
	model llms.Model,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (*llms.ContentResponse, error) {
	// The recording is handled automatically by the HTTP transport
	return model.GenerateContent(ctx, messages, options...)
}

// CallWithRecording is a helper that wraps LLM.Call with recording.
func (h *LLMTestHelper) CallWithRecording(
	ctx context.Context,
	model llms.Model,
	prompt string,
	options ...llms.CallOption,
) (string, error) {
	// The recording is handled automatically by the HTTP transport
	return model.Call(ctx, prompt, options...)
}