// Package deepseek provides a convenient interface for DeepSeek's language models.
// DeepSeek's API is OpenAI-compatible, so this package wraps the OpenAI client
// with DeepSeek-specific configurations and models.
package deepseek

import (
	"context"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// DeepSeek wraps the OpenAI client configured for DeepSeek's API.
type DeepSeek struct {
	client llms.Model
}

// Models supported by DeepSeek
const (
	ModelDeepSeekChat      = "deepseek-chat"
	ModelDeepSeekCoder     = "deepseek-coder"
	ModelDeepSeekReasoner  = "deepseek-reasoner"
	ModelDeepSeekV3        = "deepseek-v3"
)

// DefaultBaseURL is the default API base URL for DeepSeek
const DefaultBaseURL = "https://api.deepseek.com/v1"

var _ llms.Model = (*DeepSeek)(nil)

// New creates a new DeepSeek client with the given options.
func New(opts ...Option) (*DeepSeek, error) {
	config := defaultConfig()
	
	for _, opt := range opts {
		opt(&config)
	}

	// Configure OpenAI client for DeepSeek
	openaiOpts := []openai.Option{
		openai.WithBaseURL(config.BaseURL),
		openai.WithModel(config.Model),
	}
	
	if config.APIKey != "" {
		openaiOpts = append(openaiOpts, openai.WithToken(config.APIKey))
	}
	
	if config.CallbacksHandler != nil {
		openaiOpts = append(openaiOpts, openai.WithCallback(config.CallbacksHandler))
	}

	client, err := openai.New(openaiOpts...)
	if err != nil {
		return nil, err
	}

	return &DeepSeek{
		client: client,
	}, nil
}

// Call implements the llms.Model interface.
func (d *DeepSeek) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return d.client.Call(ctx, prompt, options...)
}

// GenerateContent implements the llms.Model interface.
func (d *DeepSeek) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return d.client.GenerateContent(ctx, messages, options...)
}


// Configuration for DeepSeek client
type config struct {
	APIKey           string
	Model            string
	BaseURL          string
	CallbacksHandler callbacks.Handler
}

func defaultConfig() config {
	return config{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   ModelDeepSeekChat,
		BaseURL: DefaultBaseURL,
	}
}

// Option is a function type for configuring the DeepSeek client.
type Option func(*config)

// WithAPIKey sets the API key for DeepSeek.
func WithAPIKey(apiKey string) Option {
	return func(c *config) {
		c.APIKey = apiKey
	}
}

// WithModel sets the model name for DeepSeek.
func WithModel(model string) Option {
	return func(c *config) {
		c.Model = model
	}
}

// WithBaseURL sets the base URL for DeepSeek API.
func WithBaseURL(baseURL string) Option {
	return func(c *config) {
		c.BaseURL = baseURL
	}
}

// WithCallback sets the callback handler for DeepSeek.
func WithCallback(callbacksHandler callbacks.Handler) Option {
	return func(c *config) {
		c.CallbacksHandler = callbacksHandler
	}
}