// Package deepseek provides convenience functions for using DeepSeek models.
// This package wraps the OpenAI client since DeepSeek provides an OpenAI-compatible API.
package deepseek

import (
	"context"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

// Model represents available DeepSeek models.
type Model string

const (
	// ModelReasoner is the advanced reasoning model with step-by-step thinking.
	ModelReasoner Model = "deepseek-reasoner"
	// ModelChat is the general chat model.
	ModelChat Model = "deepseek-chat"
	// ModelCoder is the code-specialized model.
	ModelCoder Model = "deepseek-coder"
)

// DefaultBaseURL is the default DeepSeek API base URL.
const DefaultBaseURL = "https://api.deepseek.com/v1"

// LLM wraps the OpenAI client for DeepSeek models.
type LLM struct {
	*openai.LLM
}

// Option configures the DeepSeek client.
type Option func(*config)

type config struct {
	token    string
	model    Model
	baseURL  string
	options  []openai.Option
}

// WithToken sets the API token.
func WithToken(token string) Option {
	return func(c *config) {
		c.token = token
	}
}

// WithModel sets the model to use.
func WithModel(model Model) Option {
	return func(c *config) {
		c.model = model
	}
}

// WithBaseURL sets a custom base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *config) {
		c.baseURL = baseURL
	}
}

// WithOpenAIOption allows passing any OpenAI client option.
func WithOpenAIOption(opt openai.Option) Option {
	return func(c *config) {
		c.options = append(c.options, opt)
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client openaiclient.Doer) Option {
	return WithOpenAIOption(openai.WithHTTPClient(client))
}

// New creates a new DeepSeek LLM client.
func New(opts ...Option) (*LLM, error) {
	cfg := &config{
		model:   ModelReasoner, // Default to reasoning model
		baseURL: DefaultBaseURL,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Auto-detect API token from environment if not explicitly set
	if cfg.token == "" {
		// Try DEEPSEEK_API_KEY first, then fall back to OPENAI_API_KEY for compatibility
		if token := os.Getenv("DEEPSEEK_API_KEY"); token != "" {
			cfg.token = token
		} else if token := os.Getenv("OPENAI_API_KEY"); token != "" {
			cfg.token = token
		}
	}

	// Build OpenAI options
	openaiOpts := []openai.Option{
		openai.WithModel(string(cfg.model)),
		openai.WithBaseURL(cfg.baseURL),
	}

	if cfg.token != "" {
		openaiOpts = append(openaiOpts, openai.WithToken(cfg.token))
	}

	// Add any additional OpenAI options
	openaiOpts = append(openaiOpts, cfg.options...)

	client, err := openai.New(openaiOpts...)
	if err != nil {
		return nil, err
	}

	return &LLM{LLM: client}, nil
}

// GenerateWithReasoning is a convenience method for reasoning models that
// returns both reasoning content and final answer separately.
func (d *LLM) GenerateWithReasoning(
	ctx context.Context,
	messages []llms.MessageContent,
	options ...llms.CallOption,
) (reasoning, content string, err error) {
	response, err := d.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", "", err
	}

	if len(response.Choices) == 0 {
		return "", "", nil
	}

	choice := response.Choices[0]
	return choice.ReasoningContent, choice.Content, nil
}

// Chat is a convenience method for simple chat interactions.
func (d *LLM) Chat(ctx context.Context, message string, options ...llms.CallOption) (string, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, message),
	}

	response, err := d.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", nil
	}

	return response.Choices[0].Content, nil
}

// ChatWithReasoning is a convenience method that returns both reasoning and answer for simple chat.
func (d *LLM) ChatWithReasoning(ctx context.Context, message string, options ...llms.CallOption) (reasoning, answer string, err error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, message),
	}

	return d.GenerateWithReasoning(ctx, messages, options...)
}