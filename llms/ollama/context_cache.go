package ollama

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
)

// ContextCache provides a simple in-memory cache for conversation contexts.
// This helps reduce token usage by reusing processed context across requests.
// Note: This is different from provider-native caching like Anthropic/Google AI.
type ContextCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int           // Maximum number of entries
	ttl     time.Duration // Time to live for cache entries
}

// CacheEntry represents a cached context entry.
type CacheEntry struct {
	Messages      []llms.MessageContent
	ContextTokens int
	CreatedAt     time.Time
	LastAccessed  time.Time
	AccessCount   int
}

// NewContextCache creates a new context cache with specified capacity and TTL.
func NewContextCache(maxSize int, ttl time.Duration) *ContextCache {
	return &ContextCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// generateCacheKey creates a unique key for a set of messages.
func (c *ContextCache) generateCacheKey(messages []llms.MessageContent) string {
	h := sha256.New()
	for _, msg := range messages {
		h.Write([]byte(msg.Role))
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				h.Write([]byte(p.Text))
			}
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for brevity
}

// Get retrieves a cached context if available and not expired.
func (c *ContextCache) Get(messages []llms.MessageContent) (*CacheEntry, bool) {
	key := c.generateCacheKey(messages)

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > c.ttl {
		// Entry expired, don't return it
		// Note: We don't delete here to avoid lock upgrade
		return nil, false
	}

	// Update access info
	entry.LastAccessed = time.Now()
	entry.AccessCount++

	return entry, true
}

// Put stores a context in the cache.
func (c *ContextCache) Put(messages []llms.MessageContent, contextTokens int) {
	key := c.generateCacheKey(messages)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean up expired entries if we're at capacity
	if len(c.entries) >= c.maxSize {
		c.evictExpiredOrOldest()
	}

	c.entries[key] = &CacheEntry{
		Messages:      messages,
		ContextTokens: contextTokens,
		CreatedAt:     time.Now(),
		LastAccessed:  time.Now(),
		AccessCount:   1,
	}
}

// evictExpiredOrOldest removes expired entries or the oldest entry if at capacity.
func (c *ContextCache) evictExpiredOrOldest() {
	now := time.Now()

	// First, remove expired entries
	for key, entry := range c.entries {
		if now.Sub(entry.CreatedAt) > c.ttl {
			delete(c.entries, key)
		}
	}

	// If still at capacity, remove least recently accessed
	if len(c.entries) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time

		for key, entry := range c.entries {
			if oldestKey == "" || entry.LastAccessed.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.LastAccessed
			}
		}

		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}
}

// Clear removes all entries from the cache.
func (c *ContextCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Stats returns cache statistics.
func (c *ContextCache) Stats() (entries int, totalHits int, avgTokensSaved int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries = len(c.entries)
	totalTokens := 0

	for _, entry := range c.entries {
		totalHits += entry.AccessCount - 1 // Subtract 1 for initial put
		if entry.AccessCount > 1 {
			totalTokens += entry.ContextTokens * (entry.AccessCount - 1)
		}
	}

	if totalHits > 0 {
		avgTokensSaved = totalTokens / totalHits
	}

	return
}

// WithContextCache creates a call option to use cached context.
func WithContextCache(cache *ContextCache) llms.CallOption {
	return func(opts *llms.CallOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]interface{})
		}
		opts.Metadata["context_cache"] = cache
	}
}
