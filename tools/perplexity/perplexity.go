package perplexity

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/tools"
)

// Model represents a Perplexity AI model type.
type Model string

// Model pricing overview: https://docs.perplexity.ai/guides/pricing
const (
	// ModelSonarDeepResearch is the deep research model with 128k context window for chat completion.
	ModelSonarDeepResearch Model = "sonar-deep-research"
	// ModelSonarReasoningPro is the reasoning pro model with 128k context window for chat completion.
	ModelSonarReasoningPro Model = "sonar-reasoning-pro"
	// ModelSonarReasoning is the reasoning model with 128k context window for chat completion.
	ModelSonarReasoning Model = "sonar-reasoning"
	// ModelSonarPro is the pro model with 200k context window for chat completion.
	ModelSonarPro Model = "sonar-pro"
	// ModelSonar is the standard model with 128k context window for chat completion.
	ModelSonar Model = "sonar"
	// ModelR11776 is the r1-1776 model.
	ModelR11776 Model = "r1-1776"
)

// Option is a function that modifies the options for the Perplexity AI tool.
type Option func(*options)

type options struct {
	apiKey string
	model  Model
}

// WithAPIKey sets the API key for Perplexity AI.
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.apiKey = apiKey
	}
}

// WithModel sets the model to be used by Perplexity AI.
func WithModel(model Model) Option {
	return func(o *options) {
		o.model = model
	}
}

// Tool implements the Perplexity AI integration.
type Tool struct {
	llm              *openai.LLM
	CallbacksHandler callbacks.Handler
}

var _ tools.Tool = (*Tool)(nil)

// New creates a new instance of the Perplexity AI tool with the given options.
func New(opts ...Option) (*Tool, error) {
	options := &options{
		apiKey: os.Getenv("PERPLEXITY_API_KEY"),
		model:  ModelSonar, // Default model
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, fmt.Errorf("PERPLEXITY_API_KEY key not set")
	}

	llm, err := openai.New(
		openai.WithModel(string(options.model)),
		openai.WithBaseURL("https://api.perplexity.ai"),
		openai.WithToken(options.apiKey),
	)
	if err != nil {
		return nil, err
	}

	return &Tool{
		llm: llm,
	}, nil
}

// Name returns the name of the tool.
func (t *Tool) Name() string {
	return "PerplexityAI"
}

// Description returns a description of the Perplexity AI tool's capabilities.
func (t *Tool) Description() string {
	return "Perplexity AI has access to a wide range of information, as it functions as an AI-powered search engine that indexes, analyzes, and summarizes content from across the internet."
}

// Call executes a query against the Perplexity AI model and returns the response.
func (t *Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, input),
	}

	var generatedText string
	_, err := t.llm.GenerateContent(ctx, content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			generatedText += string(chunk)
			return nil
		}))
	if err != nil {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, generatedText)
	}

	return generatedText, nil
}
