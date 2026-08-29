package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/require"
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
	ctx := t.Context()
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
	ctx := t.Context()
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
	streamDone := false
	streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
		switch chunk.Type { //nolint:exhaustive
		case streaming.ChunkTypeText:
			rq.Equal("world", chunk.Content)
			stream = true
		case streaming.ChunkTypeDone:
			streamDone = true
		default:
			// skip other chunks
		}
		return nil
	}

	// expect that the value is fetched from the LLM and cached
	act, err := llm.Call(ctx, "hello", llms.WithStreamingFunc(streamingFunc))
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.False(mockCache.hit)
	rq.True(stream)
	rq.True(streamDone)

	stream = false
	streamDone = false

	// expect that the cached value is returned
	act, err = llm.Call(ctx, "hello", llms.WithStreamingFunc(streamingFunc))
	rq.NoError(err)
	rq.Equal("world", act)
	rq.Equal(1, mockLLM.called)
	rq.Equal(1, mockCache.puts)
	rq.True(mockCache.hit)
	rq.True(stream)
	rq.True(streamDone)
}

func TestAWarmCacheReturnsItsAnswerWhenTheConsumerGivesUp(t *testing.T) {
	t.Parallel()

	answer := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "sixty rooms are free"}},
	}
	llm := New(newMockLLM(answer, nil), newMockCache())
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")}

	_, err := llm.GenerateContent(context.Background(), msgs)
	require.NoError(t, err, "the first call fills the cache")

	gaveUp := errors.New("consumer gave up")
	resp, err := llm.GenerateContent(context.Background(), msgs,
		llms.WithStreamingFunc(func(context.Context, streaming.Chunk) error { return gaveUp }))

	require.ErrorIs(t, err, gaveUp)
	require.NotNil(t, resp, "the cached answer exists whether or not the consumer wanted it streamed")
	require.Len(t, resp.Choices, 1)
	require.Equal(t, "sixty rooms are free", resp.Choices[0].Content)
}
