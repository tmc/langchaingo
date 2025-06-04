package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestCache_hashKeyForCache(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		v1          []llms.MessageContent
		v1opt       []llms.CallOption
		v2          []llms.MessageContent
		shouldMatch bool
	}{
		{
			name:        "empty",
			v1:          []llms.MessageContent{},
			v2:          []llms.MessageContent{},
			shouldMatch: true,
		},
		{
			name:        "empty vs non-empty",
			v1:          []llms.MessageContent{},
			v2:          []llms.MessageContent{{}},
			shouldMatch: false,
		},
		{
			name:        "different options",
			v1:          []llms.MessageContent{{}},
			v1opt:       []llms.CallOption{llms.WithCandidateCount(1)},
			v2:          []llms.MessageContent{{}},
			shouldMatch: false,
		},
	}
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
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v1hash := mustHashKeyForCache(tc.v1, tc.v1opt...)
			v2hash := mustHashKeyForCache(tc.v2)
			if (v1hash == v2hash) != tc.shouldMatch {
				t.Fatalf("expected %v, got %v", tc.shouldMatch, v1hash == v2hash)
			}
		})
	}
}

func TestCache_Call(t *testing.T) {
	ctx := context.Background()
	t.Parallel()
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
	ctx := context.Background()
	t.Parallel()
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
