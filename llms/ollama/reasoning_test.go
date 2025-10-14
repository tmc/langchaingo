package ollama

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vendasta/langchaingo/llms"
)

func TestOllama_SupportsReasoning(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "DeepSeek R1 supports reasoning",
			model:    "deepseek-r1:latest",
			expected: true,
		},
		{
			name:     "DeepSeek R1 32b supports reasoning",
			model:    "deepseek-r1:32b",
			expected: true,
		},
		{
			name:     "QwQ model supports reasoning",
			model:    "qwq:32b",
			expected: true,
		},
		{
			name:     "Model with reasoning in name supports reasoning",
			model:    "custom-reasoning:latest",
			expected: true,
		},
		{
			name:     "Model with thinking in name supports reasoning",
			model:    "my-thinking-model:v1",
			expected: true,
		},
		{
			name:     "Llama does not support reasoning",
			model:    "llama3:latest",
			expected: false,
		},
		{
			name:     "Mistral does not support reasoning",
			model:    "mistral:latest",
			expected: false,
		},
		{
			name:     "Phi does not support reasoning",
			model:    "phi:latest",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := &LLM{
				options: options{
					model: tt.model,
				},
			}

			got := llm.SupportsReasoning()
			if got != tt.expected {
				t.Errorf("SupportsReasoning() for model %s = %v, want %v", tt.model, got, tt.expected)
			}
		})
	}
}

func TestOllama_ContextCache(t *testing.T) {
	// Create a context cache with 10 entries and 5 minute TTL
	cache := NewContextCache(10, 5*time.Minute)

	// Test messages
	messages1 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	}

	messages2 := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of Germany?"),
			},
		},
	}

	// Test Put and Get
	cache.Put(messages1, 100)

	entry, hit := cache.Get(messages1)
	if !hit {
		t.Error("Expected cache hit for messages1")
	}
	if entry == nil || entry.ContextTokens != 100 {
		t.Error("Invalid cache entry returned")
	}

	// Test cache miss
	_, hit = cache.Get(messages2)
	if hit {
		t.Error("Expected cache miss for messages2")
	}

	// Test multiple accesses
	cache.Get(messages1)
	cache.Get(messages1)

	entries, totalHits, avgTokensSaved := cache.Stats()
	if entries != 1 {
		t.Errorf("Expected 1 entry, got %d", entries)
	}
	if totalHits != 3 { // 3 additional gets after initial put
		t.Errorf("Expected 3 total hits, got %d", totalHits)
	}
	if avgTokensSaved != 100 {
		t.Errorf("Expected 100 average tokens saved, got %d", avgTokensSaved)
	}

	// Test Clear
	cache.Clear()
	entries, _, _ = cache.Stats()
	if entries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", entries)
	}
}

func TestOllama_ReasoningIntegration(t *testing.T) {
	// Skip if Ollama is not available
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	// Try to create client
	llm, err := New(
		WithServerURL(serverURL),
		WithModel("deepseek-r1:latest"), // Use a reasoning model if available
	)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	ctx := context.Background()

	// Check if it implements ReasoningModel
	if _, ok := interface{}(llm).(llms.ReasoningModel); !ok {
		t.Error("Ollama LLM should implement ReasoningModel interface")
	}

	// Test with thinking mode enabled
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 15 + 27? Show your thinking."),
			},
		},
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(100),
		llms.WithThinkingMode(llms.ThinkingModeMedium),
	)
	if err != nil {
		// If the model isn't available, skip
		if strings.Contains(err.Error(), "model") || strings.Contains(err.Error(), "pull") {
			t.Skip("Reasoning model not available")
		}
		t.Fatalf("Failed to generate content: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No response choices")
	}

	content := resp.Choices[0].Content
	if !strings.Contains(content, "42") {
		t.Logf("Response might not contain correct answer: %s", content)
	}

	// Check that thinking was enabled
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		if enabled, ok := genInfo["ThinkingEnabled"].(bool); ok && enabled {
			t.Log("Thinking mode was enabled")
		}
	}
}

func TestOllama_CachingIntegration(t *testing.T) {
	// Skip if Ollama is not available
	serverURL := os.Getenv("OLLAMA_HOST")
	if serverURL == "" {
		serverURL = "http://localhost:11434"
	}

	llm, err := New(
		WithServerURL(serverURL),
		WithModel("llama3:latest"), // Use any available model
	)
	if err != nil {
		t.Skipf("Ollama not available: %v", err)
	}

	ctx := context.Background()

	// Create context cache
	cache := NewContextCache(5, 10*time.Minute)

	// Test messages
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextPart("You are a helpful assistant."),
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 2+2?"),
			},
		},
	}

	// First request (cache miss)
	resp1, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(50),
		WithContextCache(cache),
	)
	if err != nil {
		if strings.Contains(err.Error(), "model") || strings.Contains(err.Error(), "pull") {
			t.Skip("Model not available")
		}
		t.Fatalf("First request failed: %v", err)
	}

	// Check cache miss
	if genInfo := resp1.Choices[0].GenerationInfo; genInfo != nil {
		if hit, ok := genInfo["CacheHit"].(bool); ok && hit {
			t.Error("Expected cache miss on first request")
		}
	}

	// Second request with same messages (cache hit)
	resp2, err := llm.GenerateContent(ctx, messages,
		llms.WithMaxTokens(50),
		WithContextCache(cache),
	)
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}

	// Check cache hit
	if genInfo := resp2.Choices[0].GenerationInfo; genInfo != nil {
		if hit, ok := genInfo["CacheHit"].(bool); ok && !hit {
			t.Error("Expected cache hit on second request")
		}
		if cached, ok := genInfo["CachedTokens"].(int); ok && cached > 0 {
			t.Logf("Reused %d cached tokens", cached)
		}
	}

	// Check cache stats
	entries, hits, avgSaved := cache.Stats()
	t.Logf("Cache stats: %d entries, %d hits, %d avg tokens saved", entries, hits, avgSaved)
}
