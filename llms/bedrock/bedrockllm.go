package bedrock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock/internal/bedrockclient"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const defaultModel = ModelAnthropicClaudeHaiku45

// LLM is a Bedrock LLM implementation.
type LLM struct {
	modelID           string
	client            *bedrockclient.Client
	converseClient    *bedrockclient.ConverseClient
	useConverseAPI    bool
	enableAutoCaching bool
	CallbacksHandler  callbacks.Handler
}

// New creates a new Bedrock LLM implementation.
func New(opts ...Option) (*LLM, error) {
	return NewWithContext(context.Background(), opts...)
}

// NewWithContext creates a new Bedrock LLM implementation with context.
func NewWithContext(ctx context.Context, opts ...Option) (*LLM, error) {
	o, c, converseC, err := newClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &LLM{
		client:            c,
		converseClient:    converseC,
		useConverseAPI:    o.useConverseAPI,
		enableAutoCaching: o.enableAutoCaching,
		modelID:           o.modelID,
		CallbacksHandler:  o.callbackHandler,
	}, nil
}

func newClient(ctx context.Context, opts ...Option) (*options, *bedrockclient.Client, *bedrockclient.ConverseClient, error) {
	options := &options{
		modelID: defaultModel,
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, nil, nil, err
		}
		options.client = bedrockruntime.NewFromConfig(cfg)
	}

	// Create both clients - legacy and Converse API
	legacyClient := bedrockclient.NewClient(options.client)
	converseClient := bedrockclient.NewConverseClient(options.client)

	return options, legacyClient, converseClient, nil
}

// Call implements llms.Model.
func (l *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

// GenerateContent implements llms.Model.
func (l *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{
		Model: &l.modelID,
	}
	for _, opt := range options {
		opt(&opts)
	}

	// Use Converse API if enabled
	if l.useConverseAPI {
		return l.generateContentWithConverseAPI(ctx, messages, opts)
	}

	// Use legacy implementation
	return l.generateContentWithLegacyAPI(ctx, messages, opts)
}

// generateContentWithConverseAPI uses the unified Converse API
func (l *LLM) generateContentWithConverseAPI(ctx context.Context, messages []llms.MessageContent, opts llms.CallOptions) (*llms.ContentResponse, error) {
	// Apply automatic caching to bedrock messages if enabled
	shouldAutoCache := l.enableAutoCaching && l.supportsCaching(opts.GetModel())
	m, err := processMessagesWithCaching(messages, shouldAutoCache)
	if err != nil {
		return nil, err
	}

	// Check if caching should be enabled
	// Caching is enabled if either:
	// 1. Automatic caching is enabled and model supports it
	// 2. Manual cache control is present in messages
	enableCaching := shouldAutoCache || checkIfCachingRequested(messages)

	// Build Converse input
	input := &bedrockclient.ConverseInput{
		ModelID:         opts.GetModel(),
		Messages:        m,
		Tools:           opts.Tools,
		StreamingFunc:   opts.StreamingFunc,
		ReasoningConfig: opts.Reasoning,
		EnableCaching:   enableCaching,
	}

	// Set inference parameters
	if opts.MaxTokens != nil {
		input.MaxTokens = opts.MaxTokens
	}
	if opts.Temperature != nil {
		input.Temperature = opts.Temperature
	}
	if opts.TopP != nil {
		input.TopP = opts.TopP
	}
	if len(opts.StopWords) > 0 {
		input.StopSequences = opts.StopWords
	}

	res, err := l.converseClient.CreateCompletionConverse(ctx, input)
	if err != nil {
		if l.CallbacksHandler != nil {
			l.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, res)
	}

	return res, nil
}

// generateContentWithLegacyAPI uses the original model-specific implementations
func (l *LLM) generateContentWithLegacyAPI(ctx context.Context, messages []llms.MessageContent, opts llms.CallOptions) (*llms.ContentResponse, error) {
	// Apply automatic caching to bedrock messages if enabled
	shouldAutoCache := l.enableAutoCaching && l.supportsCaching(opts.GetModel())
	m, err := processMessagesWithCaching(messages, shouldAutoCache)
	if err != nil {
		return nil, err
	}

	res, err := l.client.CreateCompletion(ctx, opts.GetModel(), m, opts)
	if err != nil {
		if l.CallbacksHandler != nil {
			l.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	if l.CallbacksHandler != nil {
		l.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, res)
	}

	return res, nil
}

func processMessages(messages []llms.MessageContent) ([]bedrockclient.Message, error) {
	return processMessagesWithCaching(messages, false)
}

func processMessagesWithCaching(messages []llms.MessageContent, autoCaching bool) ([]bedrockclient.Message, error) {
	bedrockMsgs := make([]bedrockclient.Message, 0, len(messages))

	for _, m := range messages {
		for _, part := range m.Parts {
			switch part := part.(type) {
			case CachedContent:
				// Handle cached content with cache control
				cacheControl := convertCacheControl(part.CacheControl)
				// Process the wrapped content
				switch wrapped := part.ContentPart.(type) {
				case llms.TextContent:
					bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
						Role:         m.Role,
						Content:      wrapped.Text,
						Type:         "text",
						Reasoning:    wrapped.Reasoning,
						CacheControl: cacheControl,
					})
				case llms.BinaryContent:
					bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
						Role:         m.Role,
						Content:      string(wrapped.Data),
						MimeType:     wrapped.MIMEType,
						Type:         "image",
						CacheControl: cacheControl,
					})
				default:
					return nil, errors.New("unsupported cached content type")
				}
			case llms.TextContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:      m.Role,
					Content:   part.Text,
					Type:      "text",
					Reasoning: part.Reasoning,
				})
			case llms.BinaryContent:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role:     m.Role,
					Content:  string(part.Data),
					MimeType: part.MIMEType,
					Type:     "image",
				})
			case llms.ToolCall:
				if part.FunctionCall == nil {
					return nil, errors.New("tool call missing function call data")
				}
				var arguments map[string]any
				if part.FunctionCall.Arguments != "" {
					if err := json.Unmarshal([]byte(part.FunctionCall.Arguments), &arguments); err != nil {
						return nil, fmt.Errorf("failed to unmarshal tool call arguments: %w", err)
					}
				}
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role: m.Role,
					Type: "tool_use",
					ToolCall: &bedrockclient.ToolCall{
						ID:        part.ID,
						Name:      part.FunctionCall.Name,
						Arguments: arguments,
					},
				})
			case llms.ToolCallResponse:
				bedrockMsgs = append(bedrockMsgs, bedrockclient.Message{
					Role: m.Role,
					Type: "tool_result",
					ToolResult: &bedrockclient.ToolResult{
						ToolCallID: part.ToolCallID,
						ToolName:   part.Name,
						Content:    part.Content,
					},
				})
			default:
				// Check if it's unknown type - might be a specific provider type
				return nil, errors.New("unsupported message type")
			}
		}
	}

	// Apply automatic caching if requested
	if autoCaching {
		applyAutomaticCaching(bedrockMsgs)
	}

	return bedrockMsgs, nil
}

// applyAutomaticCaching adds cache control to the last assistant or tool message
// to enable conversation history caching without manual intervention.
func applyAutomaticCaching(messages []bedrockclient.Message) {
	if len(messages) == 0 {
		return
	}

	// Find the last assistant or tool message before the current user turn
	lastCacheableIdx := -1
	for i := len(messages) - 1; i >= 0; i-- {
		// Skip the last message if it's from human (current turn)
		if i == len(messages)-1 && messages[i].Role == llms.ChatMessageTypeHuman {
			continue
		}
		// Find last assistant or tool message
		if messages[i].Role == llms.ChatMessageTypeAI || messages[i].Role == llms.ChatMessageTypeTool {
			lastCacheableIdx = i
			break
		}
	}

	if lastCacheableIdx == -1 {
		return
	}

	// Add cache control to the selected message
	if messages[lastCacheableIdx].CacheControl == nil {
		messages[lastCacheableIdx].CacheControl = &bedrockclient.CacheControl{
			Type: "ephemeral",
			TTL:  "5m",
		}
	}
}

// convertCacheControl converts shared llms.CacheControl to Bedrock-specific format
func convertCacheControl(llmCache *llms.CacheControl) *bedrockclient.CacheControl {
	if llmCache == nil {
		return nil
	}

	bedrockCache := &bedrockclient.CacheControl{
		Type: llmCache.Type,
	}

	// Convert duration to TTL string
	if llmCache.Duration > 0 {
		if llmCache.Duration >= time.Hour {
			bedrockCache.TTL = "1h"
		} else {
			bedrockCache.TTL = "5m"
		}
	}

	return bedrockCache
}

// checkIfCachingRequested checks if any messages contain CachedContent
func checkIfCachingRequested(messages []llms.MessageContent) bool {
	for _, msg := range messages {
		for _, part := range msg.Parts {
			if _, ok := part.(CachedContent); ok {
				return true
			}
		}
	}
	return false
}

// supportsCaching checks if the model supports prompt caching
func (l *LLM) supportsCaching(modelID string) bool {
	// All Claude 4.x models support prompt caching (Opus, Sonnet, Haiku)
	cachingPatterns := []string{
		"claude-opus-4",
		"claude-sonnet-4",
		"claude-haiku-4",
	}

	for _, pattern := range cachingPatterns {
		if strings.Contains(modelID, pattern) {
			return true
		}
	}
	return false
}

var _ llms.Model = (*LLM)(nil)
