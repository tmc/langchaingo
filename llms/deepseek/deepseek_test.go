package deepseek

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "default configuration",
			opts:    []Option{},
			wantErr: false,
		},
		{
			name: "with token",
			opts: []Option{
				WithToken("test-token"),
			},
			wantErr: false,
		},
		{
			name: "with model",
			opts: []Option{
				WithModel(ModelChat),
			},
			wantErr: false,
		},
		{
			name: "with custom base URL",
			opts: []Option{
				WithBaseURL("https://custom.api.com/v1"),
			},
			wantErr: false,
		},
		{
			name: "with all options",
			opts: []Option{
				WithToken("test-token"),
				WithModel(ModelCoder),
				WithBaseURL("https://custom.api.com/v1"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM")
			}
		})
	}
}

func TestModels(t *testing.T) {
	tests := []struct {
		name  string
		model Model
		want  string
	}{
		{
			name:  "reasoner model",
			model: ModelReasoner,
			want:  "deepseek-reasoner",
		},
		{
			name:  "chat model",
			model: ModelChat,
			want:  "deepseek-chat",
		},
		{
			name:  "coder model",
			model: ModelCoder,
			want:  "deepseek-coder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.model) != tt.want {
				t.Errorf("Model %s = %s, want %s", tt.name, string(tt.model), tt.want)
			}
		})
	}
}

// Mock test for convenience methods (would require actual API key for full testing)
func TestLLM_ConvenienceMethods(t *testing.T) {
	// Skip if no API key is available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// These tests would require actual API credentials
	t.Run("Chat method exists", func(t *testing.T) {
		llm, err := New(WithToken("fake-token"))
		if err != nil {
			t.Fatal(err)
		}

		// Verify method exists (compilation test)
		_ = llm.Chat
		_ = llm.ChatWithReasoning
		_ = llm.GenerateWithReasoning
	})
}

func TestConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := &config{}
		
		// Apply no options - check defaults
		if cfg.model != "" {
			t.Errorf("Expected empty model, got %s", cfg.model)
		}
		if cfg.baseURL != "" {
			t.Errorf("Expected empty baseURL, got %s", cfg.baseURL)
		}
	})

	t.Run("option application", func(t *testing.T) {
		cfg := &config{}
		
		WithToken("test-token")(cfg)
		WithModel(ModelChat)(cfg)
		WithBaseURL("https://custom.com")(cfg)
		
		if cfg.token != "test-token" {
			t.Errorf("Expected token 'test-token', got %s", cfg.token)
		}
		if cfg.model != ModelChat {
			t.Errorf("Expected model %s, got %s", ModelChat, cfg.model)
		}
		if cfg.baseURL != "https://custom.com" {
			t.Errorf("Expected baseURL 'https://custom.com', got %s", cfg.baseURL)
		}
	})
}

// Example test showing how to use the DeepSeek client
func ExampleNew() {
	// Basic usage
	llm, err := New(
		WithToken("your-deepseek-api-key"),
		WithModel(ModelReasoner),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	
	// Simple chat
	response, err := llm.Chat(ctx, "Hello!")
	if err != nil {
		panic(err)
	}
	
	_ = response // Use the response
	
	// Chat with reasoning
	reasoning, answer, err := llm.ChatWithReasoning(ctx, "Explain quantum computing")
	if err != nil {
		panic(err)
	}
	
	_ = reasoning // The model's reasoning process
	_ = answer    // The final answer
}

// Benchmark convenience methods vs direct calls
func BenchmarkDeepSeekMethods(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	llm, err := New(WithToken("fake-token"))
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Hello"),
	}

	b.ResetTimer()

	b.Run("GenerateContent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This would normally call the API
			_ = messages
		}
	})

	b.Run("GenerateWithReasoning", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// This would normally call the convenience method
			_ = messages
		}
	})
}