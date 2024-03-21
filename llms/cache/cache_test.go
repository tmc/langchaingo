package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestCache_hashKeyForCache(t *testing.T) {
	t.Parallel()

	rq := require.New(t)

	mustHashKeyForCache := func(messages []llms.MessageContent, options ...llms.CallOption) string {
		var opts llms.CallOptions
		for _, opt := range options {
			opt(&opts)
		}

		key, err := hashKeyForCache(messages, opts)
		if err != nil {
			t.Fatal(err)
		}

		return key
	}

	rq.Equal(
		mustHashKeyForCache([]llms.MessageContent{}),
		mustHashKeyForCache([]llms.MessageContent{}),
	)

	rq.Equal(
		mustHashKeyForCache([]llms.MessageContent{}, llms.WithCandidateCount(1)),
		mustHashKeyForCache([]llms.MessageContent{}, llms.WithCandidateCount(1)),
	)

	rq.NotEqual(
		mustHashKeyForCache([]llms.MessageContent{{}}),
		mustHashKeyForCache([]llms.MessageContent{}),
	)

	rq.NotEqual(
		mustHashKeyForCache([]llms.MessageContent{}, llms.WithCandidateCount(1)),
		mustHashKeyForCache([]llms.MessageContent{}),
	)
}

func TestCache_Call(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rq := require.New(t)

	exp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content: "world",
		}},
	}
	mockLLM := newMockLLM(exp, nil)
	mockCache := newMockCache()

	llm := New(mockLLM, mockCache)

	// expect that the value is fetched from the LLM and cached
	act, err := llm.Call(ctx, "hello")
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.False(mockCache.hit)

	// expect that the cached value is returned
	act, err = llm.Call(ctx, "hello")
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.True(mockCache.hit)

	// expect that the value is fetched from the LLM and cached
	act, err = llm.Call(ctx, "goodbye")
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(2, mockLLM.called)
	rq.Equal(2, mockCache.puts)
	rq.False(mockCache.hit)

	// expect that the cached value is returned
	act, err = llm.Call(ctx, "goodbye")
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(2, mockLLM.called)
	rq.Equal(2, mockCache.puts)
	rq.True(mockCache.hit)

	// expect that the cached value is returned
	act, err = llm.Call(ctx, "hello")
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(2, mockLLM.called)
	rq.Equal(2, mockCache.puts)
	rq.True(mockCache.hit)
}

func TestCache_Call_Streaming(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rq := require.New(t)

	exp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{
			Content: "world",
		}},
	}
	mockLLM := newMockLLM(exp, nil)
	mockCache := newMockCache()

	llm := New(mockLLM, mockCache)

	stream := false

	// expect that the value is fetched from the LLM and cached
	act, err := llm.Call(ctx, "hello", llms.WithStreamingFunc(
		func(_ context.Context, bs []byte) error {
			rq.Equal("world", string(bs))

			stream = true

			return nil
		}))
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.False(mockCache.hit)
	rq.True(stream)

	stream = false

	// expect that the cached value is returned
	act, err = llm.Call(ctx, "hello", llms.WithStreamingFunc(
		func(_ context.Context, bs []byte) error {
			rq.Equal("world", string(bs))

			stream = true

			return nil
		}))
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.True(mockCache.hit)
	rq.True(stream)
}
