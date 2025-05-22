package httprr

import (
	"testing"

	"github.com/tmc/langchaingo/llms/anthropic"
	anthropicclient "github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	openaiclient "github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

// LLMTestHelper provides convenience methods for testing LLM clients with httprr.
type LLMTestHelper struct {
	*TestHelper
}

// NewLLMTestHelper creates a new LLMTestHelper with httprr recording capabilities.
func NewLLMTestHelper(t testing.TB, recordingsDir string) *LLMTestHelper {
	return &LLMTestHelper{
		TestHelper: NewAutoHelper(t, recordingsDir),
	}
}

// NewOpenAIClient creates an OpenAI client that uses the httprr recorder.
func (h *LLMTestHelper) NewOpenAIClient(apiKey, model string, opts ...openai.Option) (*openai.LLM, error) {
	// Add the httprr client as an option
	clientOpt := openai.WithHTTPClient(h.Client)
	allOpts := append([]openai.Option{clientOpt}, opts...)
	
	return openai.New(allOpts...)
}

// NewOpenAIClientWithToken creates an OpenAI client with a specific API token that uses the httprr recorder.
func (h *LLMTestHelper) NewOpenAIClientWithToken(apiKey string, opts ...openai.Option) (*openai.LLM, error) {
	// Add the httprr client as an option along with the API key
	clientOpt := openai.WithHTTPClient(h.Client)
	tokenOpt := openai.WithToken(apiKey)
	allOpts := append([]openai.Option{clientOpt, tokenOpt}, opts...)
	
	return openai.New(allOpts...)
}

// NewAnthropicClient creates an Anthropic client that uses the httprr recorder.
func (h *LLMTestHelper) NewAnthropicClient(apiKey string, opts ...anthropic.Option) (*anthropic.LLM, error) {
	// Add the httprr client as an option
	clientOpt := anthropic.WithHTTPClient(h.Client)
	allOpts := append([]anthropic.Option{clientOpt}, opts...)
	
	return anthropic.New(allOpts...)
}

// NewOllamaClient creates an Ollama client that uses the httprr recorder.
func (h *LLMTestHelper) NewOllamaClient(opts ...ollama.Option) (*ollama.LLM, error) {
	// Add the httprr client as an option
	clientOpt := ollama.WithHTTPClient(h.Client)
	allOpts := append([]ollama.Option{clientOpt}, opts...)
	
	return ollama.New(allOpts...)
}

// WithOpenAIHTTPClient creates an OpenAI client option that uses the httprr recorder.
func (h *LLMTestHelper) WithOpenAIHTTPClient() openai.Option {
	return openai.WithHTTPClient(h.Client)
}

// WithAnthropicHTTPClient creates an Anthropic client option that uses the httprr recorder.
func (h *LLMTestHelper) WithAnthropicHTTPClient() anthropic.Option {
	return anthropic.WithHTTPClient(h.Client)
}

// WithOllamaHTTPClient creates an Ollama client option that uses the httprr recorder.
func (h *LLMTestHelper) WithOllamaHTTPClient() ollama.Option {
	return ollama.WithHTTPClient(h.Client)
}

// OpenAIClientOption creates an OpenAI client option that can be used directly in tests.
// This is useful when you want to create the client yourself but need the httprr transport.
func (h *LLMTestHelper) OpenAIClientOption() openaiclient.Option {
	return openaiclient.WithHTTPClient(h.Client)
}

// AnthropicClientOption creates an Anthropic client option that can be used directly in tests.
// This is useful when you want to create the client yourself but need the httprr transport.
func (h *LLMTestHelper) AnthropicClientOption() anthropicclient.Option {
	return anthropicclient.WithHTTPClient(h.Client)
}