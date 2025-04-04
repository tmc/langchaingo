package inmemory

import (
	"context"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/averikitsch/langchaingo/llms"
)

// InMemory is an in-memory `cache.Backend`.
type InMemory struct {
	Options Options
	cache   *cache.Cache[string, *llms.ContentResponse]
}

// New creates a new in-memory `cache.Backend` implementation with the supplied
// options. Note that this starts a go-routine to evict expired items from the
// cache. This go-routine is terminated when the context is cancelled.
func New(ctx context.Context, opts ...Option) (*InMemory, error) {
	options, err := applyOptions(opts...)
	if err != nil {
		return nil, err
	}

	return &InMemory{
		Options: *options,
		cache:   cache.NewContext(ctx, options.CacheOptions...),
	}, nil
}

// Get a value from the cache. If the key is not found, return `nil`.
func (im *InMemory) Get(_ context.Context, key string) *llms.ContentResponse {
	// errors are ignored, instead we return `nil` and pretend the key
	// wasn't found.
	v, _ := im.cache.Get(key)

	return v
}

// Put a value into the cache.
func (im *InMemory) Put(_ context.Context, key string, value *llms.ContentResponse) {
	im.cache.Set(key, value, im.Options.ItemOptions...)
}
