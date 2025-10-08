package inmemory

import (
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/tmc/langchaingo/llms"
)

// Option is a functional argument that configures the Options.
type Option func(*Options) error

// Options is a set of options for the in-memory cache.
type Options struct {
	CacheOptions []cache.Option[string, *llms.ContentResponse]
	ItemOptions  []cache.ItemOption
}

// WithCacheOptions specifies the options for the underlying cache. Note: multiple
// instances are appended, so `New(ctx, WithCacheOptions(opt1, opt2))` is the same
// as `New(ctx, WithCacheOptions(opt1), WithCacheOptions(opt2))`.
func WithCacheOptions(opts ...cache.Option[string, *llms.ContentResponse]) Option {
	return func(o *Options) error {
		o.CacheOptions = append(o.CacheOptions, opts...)

		return nil
	}
}

// WithItemOptions specifies the options for the underlying cache for each specific
// item that is added. Note: multiple instances are appended, so
// `New(ctx, WithItemOptions(opt1, opt2))` is the same as
// `New(ctx, WithItemOptions(opt1), WithItemOptions(opt2))`.
func WithItemOptions(opts ...cache.ItemOption) Option {
	return func(o *Options) error {
		o.ItemOptions = append(o.ItemOptions, opts...)

		return nil
	}
}

// WithExpiration specifies the time-to-live for specific items that are added to
// the cache. This is the same as: `WithItemOptions(cache.WithExpiration(expiration))`.
func WithExpiration(expiration time.Duration) Option {
	return func(o *Options) error {
		o.ItemOptions = append(o.ItemOptions, cache.WithExpiration(expiration))

		return nil
	}
}

func applyOptions(opts ...Option) (*Options, error) {
	o := new(Options)

	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	return o, nil
}
