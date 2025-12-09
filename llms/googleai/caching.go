// Package googleai provides caching support for Google AI models.
package googleai

import (
	"context"
	"time"

	"github.com/vendasta/langchaingo/llms"
	"google.golang.org/genai"
)

// CachingHelper provides utilities for working with Google AI's cached content feature.
// Unlike Anthropic which supports inline cache control, Google AI requires
// pre-creating cached content through the API.
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
// Example usage:
//
//	helper, _ := NewCachingHelper(ctx, WithAPIKey(apiKey))
//	cached, _ := helper.CreateCachedContent(ctx, "gemini-2.0-flash", []llms.MessageContent{
//	    {
//	        Role: llms.ChatMessageTypeSystem,
//	        Parts: []llms.ContentPart{
//	            llms.TextPart("You are an expert assistant with deep knowledge..."),
//	        },
//	    },
//	}, 1*time.Hour)
//
//	// Use the cached content in requests
//	model, _ := New(ctx, WithAPIKey(apiKey))
//	resp, _ := model.GenerateContent(ctx, messages, WithCachedContent(cached.Name))
func (ch *CachingHelper) CreateCachedContent(
	ctx context.Context,
	modelName string,
	messages []llms.MessageContent,
	ttl time.Duration,
) (*genai.CachedContent, error) {
	// Convert langchain messages to genai content
	contents := make([]*genai.Content, 0, len(messages))
	var systemInstruction *genai.Content

	for _, msg := range messages {
		parts := make([]*genai.Part, 0, len(msg.Parts))
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				parts = append(parts, &genai.Part{Text: p.Text})
			case llms.CachedContent:
				// Extract the underlying content if it's wrapped with cache control
				// (though Google AI doesn't use inline cache control like Anthropic)
				if textPart, ok := p.ContentPart.(llms.TextContent); ok {
					parts = append(parts, &genai.Part{Text: textPart.Text})
				}
			}
		}

		content := &genai.Content{
			Parts: parts,
		}

		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			content.Role = "system"
			systemInstruction = content
		case llms.ChatMessageTypeHuman:
			content.Role = "user"
			contents = append(contents, content)
		case llms.ChatMessageTypeAI:
			content.Role = "model"
			contents = append(contents, content)
		}
	}

	// Create the cached content using the new SDK API
	config := &genai.CreateCachedContentConfig{
		TTL:               ttl,
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
	if err != nil {
		return err
	}
	return nil
}

// ListCachedContents returns an iterator for all cached content.
// TODO: Implement if used?
func (ch *CachingHelper) ListCachedContents(ctx context.Context) interface{} {
	return nil
}
