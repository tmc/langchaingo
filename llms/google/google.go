// Package google provides a unified interface for Google AI LLMs that automatically
// selects the appropriate underlying provider based on the model being used.
//
// For gemini-3+ models, this package uses googleaiv2 (google.golang.org/genai SDK).
// For older models (gemini-2.x, gemini-1.x, etc.), it uses googleai (github.com/google/generative-ai-go SDK).
//
// This allows seamless migration to newer models without changing client code.
package google

import (
	"context"
	"strings"

	"github.com/vendasta/langchaingo/callbacks"
	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/googleai"
	"github.com/vendasta/langchaingo/llms/googleaiv2"
)

// GoogleAI is a unified client that automatically routes to the appropriate
// underlying provider based on the model being used.
type GoogleAI struct {
	// The underlying v1 client (for gemini-2.x and older)
	v1Client *googleai.GoogleAI
	// The underlying v2 client (for gemini-3+)
	v2Client *googleaiv2.GoogleAI
	// Which client is active
	useV2 bool
	// Store options for potential re-creation
	opts Options
}

var (
	_ llms.Model = &GoogleAI{}
)

// IsGemini3OrNewer returns true if the model name indicates gemini-3 or newer.
func IsGemini3OrNewer(model string) bool {
	model = strings.ToLower(model)

	// Check for gemini-3, gemini-4, etc.
	if strings.Contains(model, "gemini-3") ||
		strings.Contains(model, "gemini-4") ||
		strings.Contains(model, "gemini-5") {
		return true
	}

	// Also check for numeric versions like gemini-3.0, gemini-3.5, etc.
	// Pattern: gemini-X where X >= 3
	if strings.HasPrefix(model, "gemini-") {
		// Extract version part
		rest := strings.TrimPrefix(model, "gemini-")
		// Get first character/digit
		if len(rest) > 0 {
			firstChar := rest[0]
			if firstChar >= '3' && firstChar <= '9' {
				return true
			}
		}
	}

	return false
}

// New creates a new GoogleAI client that automatically selects the appropriate
// underlying provider based on the model.
func New(ctx context.Context, opts ...Option) (*GoogleAI, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	g := &GoogleAI{
		opts:  options,
		useV2: IsGemini3OrNewer(options.DefaultModel),
	}

	var err error
	if g.useV2 {
		// Use the new SDK for gemini-3+
		v2Opts := convertToV2Options(options)
		g.v2Client, err = googleaiv2.New(ctx, v2Opts...)
	} else {
		// Use the original SDK for older models
		v1Opts := convertToV1Options(options)
		g.v1Client, err = googleai.New(ctx, v1Opts...)
	}

	if err != nil {
		return nil, err
	}

	return g, nil
}

// Close closes the underlying client.
func (g *GoogleAI) Close() error {
	if g.useV2 && g.v2Client != nil {
		return g.v2Client.Close()
	}
	if !g.useV2 && g.v1Client != nil {
		return g.v1Client.Close()
	}
	return nil
}

// Call implements the llms.Model interface.
func (g *GoogleAI) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, g, prompt, options...)
}

// GenerateContent implements the llms.Model interface.
func (g *GoogleAI) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	// Check if a different model is specified in options that requires switching providers
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Determine which client to use based on the effective model
	effectiveModel := opts.Model
	if effectiveModel == "" {
		effectiveModel = g.opts.DefaultModel
	}

	useV2 := IsGemini3OrNewer(effectiveModel)

	// If the model requires a different provider than what we have initialized,
	// we need to use the appropriate client
	if useV2 {
		if g.v2Client == nil {
			// Lazy initialization of v2 client
			var err error
			v2Opts := convertToV2Options(g.opts)
			g.v2Client, err = googleaiv2.New(context.Background(), v2Opts...)
			if err != nil {
				return nil, err
			}
		}
		return g.v2Client.GenerateContent(ctx, messages, options...)
	}

	if g.v1Client == nil {
		// Lazy initialization of v1 client
		var err error
		v1Opts := convertToV1Options(g.opts)
		g.v1Client, err = googleai.New(context.Background(), v1Opts...)
		if err != nil {
			return nil, err
		}
	}
	return g.v1Client.GenerateContent(ctx, messages, options...)
}

// GetCallbackHandler returns the callbacks handler for the active client.
func (g *GoogleAI) GetCallbackHandler() callbacks.Handler {
	if g.useV2 && g.v2Client != nil {
		return g.v2Client.CallbacksHandler
	}
	if g.v1Client != nil {
		return g.v1Client.CallbacksHandler
	}
	return nil
}

// SetCallbackHandler sets the callbacks handler on the active client.
func (g *GoogleAI) SetCallbackHandler(handler callbacks.Handler) {
	if g.v2Client != nil {
		g.v2Client.CallbacksHandler = handler
	}
	if g.v1Client != nil {
		g.v1Client.CallbacksHandler = handler
	}
}

// SupportsReasoning returns true if the current model supports reasoning/thinking tokens.
func (g *GoogleAI) SupportsReasoning() bool {
	if g.useV2 && g.v2Client != nil {
		return g.v2Client.SupportsReasoning()
	}
	if g.v1Client != nil {
		return g.v1Client.SupportsReasoning()
	}
	return false
}

// convertToV1Options converts unified Options to googleai options.
func convertToV1Options(opts Options) []googleai.Option {
	var v1Opts []googleai.Option

	if opts.APIKey != "" {
		v1Opts = append(v1Opts, googleai.WithAPIKey(opts.APIKey))
	}
	if opts.CloudProject != "" {
		v1Opts = append(v1Opts, googleai.WithCloudProject(opts.CloudProject))
	}
	if opts.CloudLocation != "" {
		v1Opts = append(v1Opts, googleai.WithCloudLocation(opts.CloudLocation))
	}
	if opts.DefaultModel != "" {
		v1Opts = append(v1Opts, googleai.WithDefaultModel(opts.DefaultModel))
	}
	if opts.DefaultEmbeddingModel != "" {
		v1Opts = append(v1Opts, googleai.WithDefaultEmbeddingModel(opts.DefaultEmbeddingModel))
	}
	if opts.DefaultCandidateCount > 0 {
		v1Opts = append(v1Opts, googleai.WithDefaultCandidateCount(opts.DefaultCandidateCount))
	}
	if opts.DefaultMaxTokens > 0 {
		v1Opts = append(v1Opts, googleai.WithDefaultMaxTokens(opts.DefaultMaxTokens))
	}
	if opts.DefaultTemperature > 0 {
		v1Opts = append(v1Opts, googleai.WithDefaultTemperature(opts.DefaultTemperature))
	}
	if opts.DefaultTopK > 0 {
		v1Opts = append(v1Opts, googleai.WithDefaultTopK(opts.DefaultTopK))
	}
	if opts.DefaultTopP > 0 {
		v1Opts = append(v1Opts, googleai.WithDefaultTopP(opts.DefaultTopP))
	}
	if opts.HarmThreshold != 0 {
		v1Opts = append(v1Opts, googleai.WithHarmThreshold(googleai.HarmBlockThreshold(opts.HarmThreshold)))
	}

	return v1Opts
}

// convertToV2Options converts unified Options to googleaiv2 options.
func convertToV2Options(opts Options) []googleaiv2.Option {
	var v2Opts []googleaiv2.Option

	if opts.APIKey != "" {
		v2Opts = append(v2Opts, googleaiv2.WithAPIKey(opts.APIKey))
	}
	if opts.CloudProject != "" {
		v2Opts = append(v2Opts, googleaiv2.WithCloudProject(opts.CloudProject))
	}
	if opts.CloudLocation != "" {
		v2Opts = append(v2Opts, googleaiv2.WithCloudLocation(opts.CloudLocation))
	}
	if opts.DefaultModel != "" {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultModel(opts.DefaultModel))
	}
	if opts.DefaultEmbeddingModel != "" {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultEmbeddingModel(opts.DefaultEmbeddingModel))
	}
	if opts.DefaultCandidateCount > 0 {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultCandidateCount(opts.DefaultCandidateCount))
	}
	if opts.DefaultMaxTokens > 0 {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultMaxTokens(opts.DefaultMaxTokens))
	}
	if opts.DefaultTemperature > 0 {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultTemperature(opts.DefaultTemperature))
	}
	if opts.DefaultTopK > 0 {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultTopK(opts.DefaultTopK))
	}
	if opts.DefaultTopP > 0 {
		v2Opts = append(v2Opts, googleaiv2.WithDefaultTopP(opts.DefaultTopP))
	}
	if opts.HarmThreshold != 0 {
		v2Opts = append(v2Opts, googleaiv2.WithHarmThreshold(googleaiv2.HarmBlockThreshold(opts.HarmThreshold)))
	}

	return v2Opts
}
