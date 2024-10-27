package inmemory

import (
	"context"
	"testing"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/fifo"
	"github.com/stretchr/testify/require"
	"github.com/starmvp/langchaingo/llms"
)

func TestInMemory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rq := require.New(t)
	ttl := time.Second / 2

	cache, err := New(ctx,
		WithCacheOptions(
			cache.AsFIFO[string, *llms.ContentResponse](
				fifo.WithCapacity(1),
			),
		),
		WithExpiration(ttl),
	)
	rq.NoError(err)

	rq.Nil(cache.Get(ctx, "key1"), "empty cache should be empty")

	val := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content: "value",
		}},
	}

	cache.Put(ctx, "key1", val)
	rq.Equal(val, cache.Get(ctx, "key1"))

	cache.Put(ctx, "key2", val)
	rq.Nil(cache.Get(ctx, "key1"), "first value should have been evicted")
	rq.NotNil(cache.Get(ctx, "key2"))

	time.Sleep(ttl * 2) // double the ttl to make sure the value has timed out.
	rq.Nil(cache.Get(ctx, "key2"), "second value should have been evicted")
}
