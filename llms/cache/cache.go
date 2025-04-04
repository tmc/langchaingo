package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/averikitsch/langchaingo/llms"
)

// Backend is the interface that needs to be implemented by cache backends.
type Backend interface {
	// Get a value from the cache. If the key is not found, return `nil`.
	Get(ctx context.Context, key string) *llms.ContentResponse
	// Put a value into the cache.
	Put(ctx context.Context, key string, response *llms.ContentResponse)
}

// Cacher is an LLM wrapper that caches the responses from the LLM.
type Cacher struct {
	llm   llms.Model
	cache Backend
}

// assert that `Cacher` implements the `llms.Model` interface.
var _ llms.Model = (*Cacher)(nil)

// New wraps a Model and adds caching capabilities using the provided
// cache backend.
func New(llm llms.Model, backend Backend) *Cacher {
	return &Cacher{
		llm:   llm,
		cache: backend,
	}
}

// Call is a simplified interface for a text-only Model, generating a single
// string response from a single string prompt.
//
// Deprecated: this method is retained for backwards compatibility. Use the
// more general [GenerateContent] instead. You can also use
// the [GenerateFromSinglePrompt] function which provides a similar capability
// to Call and is built on top of the new interface.
func (c *Cacher) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, c, prompt, options...)
}

// GenerateContent asks the model to generate content from a sequence of
// messages. It's the most general interface for multi-modal LLMs that support
// chat-like interactions.
func (c *Cacher) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	var opts llms.CallOptions
	for _, opt := range options {
		opt(&opts)
	}

	key, err := hashKeyForCache(messages, opts)
	if err != nil {
		return nil, err
	}

	if response := c.cache.Get(ctx, key); response != nil {
		if opts.StreamingFunc != nil && len(response.Choices) > 0 {
			// only stream the first choice.
			if err := opts.StreamingFunc(ctx, []byte(response.Choices[0].Content)); err != nil {
				return nil, err
			}
		}

		return response, nil
	}

	response, err := c.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return nil, err
	}

	c.cache.Put(ctx, key, response)

	return response, nil
}

// hashKeyForCache is a helper function that generates a unique key for a given
// set of messages and call options.
func hashKeyForCache(messages []llms.MessageContent, opts llms.CallOptions) (string, error) {
	hash := sha256.New()
	enc := json.NewEncoder(hash)
	if err := enc.Encode(messages); err != nil {
		return "", err
	}
	if err := enc.Encode(opts); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
