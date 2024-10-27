package cache

import (
	"context"

	"github.com/starmvp/langchaingo/llms"
)

// === Mock for llms.Model

func newMockLLM(response *llms.ContentResponse, err error) *mockLLM {
	return &mockLLM{
		called:   0,
		response: response,
		error:    err,
	}
}

// not synchronized, don't use concurrently!
type mockLLM struct {
	called   int
	response *llms.ContentResponse
	error    error
}

func (m *mockLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, m, prompt, options...)
}

func (m *mockLLM) GenerateContent(ctx context.Context, _ []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	var opts llms.CallOptions
	for _, opt := range options {
		opt(&opts)
	}

	if opts.StreamingFunc != nil && len(m.response.Choices) > 0 {
		if err := opts.StreamingFunc(ctx, []byte(m.response.Choices[0].Content)); err != nil {
			return nil, err
		}
	}

	m.called++

	return m.response, m.error
}

// === Mock for cache.Cacher

func newMockCache() *mockCache {
	return &mockCache{
		entries: make(map[string]*llms.ContentResponse),
		puts:    0,
		hit:     false,
	}
}

// not synchronized, don't use concurrently!
type mockCache struct {
	entries map[string]*llms.ContentResponse
	puts    int
	hit     bool
}

func (m *mockCache) Get(_ context.Context, key string) *llms.ContentResponse {
	v, ok := m.entries[key]
	m.hit = ok

	return v
}

func (m *mockCache) Put(_ context.Context, key string, response *llms.ContentResponse) {
	m.entries[key] = response
	m.puts++
}
