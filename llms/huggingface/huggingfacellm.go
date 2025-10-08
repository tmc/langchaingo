package huggingface

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/huggingface/internal/huggingfaceclient"
)

var (
	ErrEmptyResponse            = errors.New("empty response")
	ErrMissingToken             = errors.New("missing the Hugging Face API token. Set it in the HF_TOKEN or HUGGINGFACEHUB_API_TOKEN environment variable, or save it to ~/.cache/huggingface/token") //nolint:lll
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *huggingfaceclient.Client
}

var _ llms.Model = (*LLM)(nil)

// Call implements the LLM interface.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{Model: defaultModel}
	for _, opt := range options {
		opt(opts)
	}

	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	result, err := o.client.RunInference(ctx, &huggingfaceclient.InferenceRequest{
		Model:             o.client.Model,
		Prompt:            part.(llms.TextContent).Text,
		Task:              huggingfaceclient.InferenceTaskTextGeneration,
		Temperature:       opts.Temperature,
		TopP:              opts.TopP,
		TopK:              opts.TopK,
		MinLength:         opts.MinLength,
		MaxLength:         opts.MaxLength,
		RepetitionPenalty: opts.RepetitionPenalty,
		Seed:              opts.Seed,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}
	return resp, nil
}

func New(opts ...Option) (*LLM, error) {
	options := &options{
		token: getHuggingFaceToken(),
		model: defaultModel,
		url:   defaultURL,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	// If a provider is specified, use the router URL
	if options.provider != "" {
		options.url = routerURL
	}

	var clientOpts []huggingfaceclient.Option
	if options.httpClient != nil {
		clientOpts = append(clientOpts, huggingfaceclient.WithHTTPClient(options.httpClient))
	}
	if options.provider != "" {
		clientOpts = append(clientOpts, huggingfaceclient.WithProvider(options.provider))
	}

	c, err := huggingfaceclient.New(options.token, options.model, options.url, clientOpts...)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client: c,
	}, nil
}

// getHuggingFaceToken attempts to retrieve the Hugging Face API token from various sources
// in the following order:
// 1. HF_TOKEN environment variable (current standard)
// 2. HUGGINGFACEHUB_API_TOKEN environment variable (legacy)
// 3. Token file at path specified by HF_TOKEN_PATH
// 4. Default token file at $HF_HOME/token or ~/.cache/huggingface/token
func getHuggingFaceToken() string {
	// Try HF_TOKEN first (current standard)
	if token := os.Getenv(hfTokenEnvVarName); token != "" {
		return token
	}

	// Try legacy HUGGINGFACEHUB_API_TOKEN
	if token := os.Getenv(tokenEnvVarName); token != "" {
		return token
	}

	// Try reading from token file
	tokenPath := getTokenPath()
	if tokenPath != "" {
		if data, err := os.ReadFile(tokenPath); err == nil {
			return string(bytes.TrimSpace(data))
		}
	}

	return ""
}

// getTokenPath returns the path to the token file
func getTokenPath() string {
	// Check if HF_TOKEN_PATH is set
	if path := os.Getenv(hfTokenPathEnvVarName); path != "" {
		return path
	}

	// Try HF_HOME/token
	if hfHome := os.Getenv(hfHomeEnvVarName); hfHome != "" {
		return filepath.Join(hfHome, defaultTokenPath)
	}

	// Try XDG_CACHE_HOME/huggingface/token
	if xdgCache := os.Getenv(xdgCacheHomeEnvVar); xdgCache != "" {
		return filepath.Join(xdgCache, "huggingface", defaultTokenPath)
	}

	// Default to ~/.cache/huggingface/token
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "huggingface", defaultTokenPath)
	}

	return ""
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(
	ctx context.Context,
	inputTexts []string,
	model string,
	task string,
) ([][]float32, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, model, task, &huggingfaceclient.EmbeddingRequest{
		Inputs: inputTexts,
		Options: map[string]any{
			"use_gpu":        false,
			"wait_for_model": true,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}
