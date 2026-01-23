// Package googleai provides caching support for Google AI models.
package googleai

import (
	"context"
	"time"

	"github.com/vxcontrol/langchaingo/llms"

	"google.golang.org/genai"
)

// CachingHelper provides utilities for working with Google AI's cached content feature.
// Unlike Anthropic which supports inline cache control, Google AI requires
// pre-creating cached content through the API.
//
// Google AI caching is particularly useful for:
// - Large system prompts that are reused across multiple requests
// - Extensive context documents (e.g., knowledge bases, documentation)
// - Long conversation histories
//
// The minimum cacheable content size is 32,768 tokens (~24,000 words).
// Cached content has a TTL (time-to-live) and will be automatically deleted after expiration.
type CachingHelper struct {
	client *genai.Client
}

// NewCachingHelper creates a helper for managing cached content.
func NewCachingHelper(ctx context.Context, opts ...Option) (*CachingHelper, error) {
	// Create a GoogleAI client to get access to the underlying genai client
	gai, err := New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &CachingHelper{
		client: gai.client,
	}, nil
}

// CreateCachedContent creates cached content that can be reused across multiple requests.
// This is useful for caching large system prompts, context documents, or frequently used instructions.
//
// Parameters:
//   - modelName: The model to use (e.g., "gemini-2.0-flash-exp")
//   - messages: The content to cache (must be at least 32,768 tokens)
//   - ttl: Time-to-live for the cached content (e.g., 1*time.Hour)
//   - displayName: Optional human-readable name for the cache
//
// Example usage:
//
//	helper, _ := NewCachingHelper(ctx, WithAPIKey(apiKey))
//	cached, _ := helper.CreateCachedContent(ctx, "gemini-2.0-flash-exp",
//	    []llms.MessageContent{
//	        {
//	            Role: llms.ChatMessageTypeSystem,
//	            Parts: []llms.ContentPart{
//	                llms.TextPart("You are an expert assistant..."),
//	            },
//	        },
//	    },
//	    1*time.Hour,
//	    "my-expert-system-prompt",
//	)
//
//	// Use the cached content in requests
//	model, _ := New(ctx, WithAPIKey(apiKey))
//	resp, _ := model.GenerateContent(ctx, messages, WithCachedContent(cached.Name))
func (ch *CachingHelper) CreateCachedContent(
	ctx context.Context,
	modelName string,
	messages []llms.MessageContent,
	ttl time.Duration,
	displayName string,
) (*genai.CachedContent, error) {
	// Convert langchain messages to genai content
	contents := make([]*genai.Content, 0, len(messages))
	var systemInstruction *genai.Content

	for _, msg := range messages {
		parts, err := convertParts(msg.Parts)
		if err != nil {
			return nil, err
		}

		content := &genai.Content{
			Parts: parts,
		}

		// Set role
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			content.Role = RoleSystem
			systemInstruction = content
		case llms.ChatMessageTypeHuman:
			content.Role = RoleUser
			contents = append(contents, content)
		case llms.ChatMessageTypeAI:
			content.Role = RoleModel
			contents = append(contents, content)
		default:
			content.Role = RoleUser
			contents = append(contents, content)
		}
	}

	// Create the cached content using the new API
	config := &genai.CreateCachedContentConfig{
		TTL:               ttl,
		DisplayName:       displayName,
		Contents:          contents,
		SystemInstruction: systemInstruction,
	}

	return ch.client.Caches.Create(ctx, modelName, config)
}

// GetCachedContent retrieves existing cached content by name.
func (ch *CachingHelper) GetCachedContent(ctx context.Context, name string) (*genai.CachedContent, error) {
	return ch.client.Caches.Get(ctx, name, &genai.GetCachedContentConfig{})
}

// DeleteCachedContent removes cached content.
func (ch *CachingHelper) DeleteCachedContent(ctx context.Context, name string) error {
	_, err := ch.client.Caches.Delete(ctx, name, &genai.DeleteCachedContentConfig{})
	return err
}

// UpdateCachedContent updates the TTL or expiration time of cached content.
func (ch *CachingHelper) UpdateCachedContent(ctx context.Context, name string, ttl time.Duration) (*genai.CachedContent, error) {
	return ch.client.Caches.Update(ctx, name, &genai.UpdateCachedContentConfig{
		TTL: ttl,
	})
}

// ListCachedContents lists all cached content with pagination support.
// Returns a Page that can be iterated to fetch all cached contents.
func (ch *CachingHelper) ListCachedContents(ctx context.Context, pageSize int32) (genai.Page[genai.CachedContent], error) {
	config := &genai.ListCachedContentsConfig{
		PageSize: pageSize,
	}
	return ch.client.Caches.List(ctx, config)
}

// AllCachedContents returns an iterator that yields all cached contents.
// This handles pagination automatically.
func (ch *CachingHelper) AllCachedContents(ctx context.Context) func(func(*genai.CachedContent, error) bool) {
	return ch.client.Caches.All(ctx)
}
