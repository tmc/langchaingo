package llms_test

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestThinkingModes(t *testing.T) {
	tests := []struct {
		name     string
		mode     llms.ThinkingMode
		expected string
	}{
		{"None", llms.ThinkingModeNone, "none"},
		{"Low", llms.ThinkingModeLow, "low"},
		{"Medium", llms.ThinkingModeMedium, "medium"},
		{"High", llms.ThinkingModeHigh, "high"},
		{"Auto", llms.ThinkingModeAuto, "auto"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.expected {
				t.Errorf("ThinkingMode %s = %s, want %s", tt.name, tt.mode, tt.expected)
			}
		})
	}
}

func TestDefaultThinkingConfig(t *testing.T) {
	config := llms.DefaultThinkingConfig()
	
	if config.Mode != llms.ThinkingModeAuto {
		t.Errorf("Default mode = %s, want %s", config.Mode, llms.ThinkingModeAuto)
	}
	
	if config.ReturnThinking {
		t.Error("Default ReturnThinking should be false")
	}
	
	if config.StreamThinking {
		t.Error("Default StreamThinking should be false")
	}
	
	if config.InterleaveThinking {
		t.Error("Default InterleaveThinking should be false")
	}
}

func TestWithThinking(t *testing.T) {
	config := &llms.ThinkingConfig{
		Mode:           llms.ThinkingModeHigh,
		BudgetTokens:   4096,
		ReturnThinking: true,
		StreamThinking: true,
	}
	
	option := llms.WithThinking(config)
	
	var opts llms.CallOptions
	option(&opts)
	
	if opts.Metadata == nil {
		t.Fatal("metadata should be initialized")
	}
	
	storedConfig, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if storedConfig != config {
		t.Error("stored config should match original")
	}
}

func TestWithThinkingMode(t *testing.T) {
	option := llms.WithThinkingMode(llms.ThinkingModeMedium)
	
	var opts llms.CallOptions
	option(&opts)
	
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if config.Mode != llms.ThinkingModeMedium {
		t.Errorf("mode = %s, want %s", config.Mode, llms.ThinkingModeMedium)
	}
}

func TestWithThinkingBudget(t *testing.T) {
	option := llms.WithThinkingBudget(8192)
	
	var opts llms.CallOptions
	option(&opts)
	
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if config.BudgetTokens != 8192 {
		t.Errorf("BudgetTokens = %d, want 8192", config.BudgetTokens)
	}
}

func TestWithReturnThinking(t *testing.T) {
	option := llms.WithReturnThinking(true)
	
	var opts llms.CallOptions
	option(&opts)
	
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if !config.ReturnThinking {
		t.Error("ReturnThinking should be true")
	}
}

func TestWithStreamThinking(t *testing.T) {
	option := llms.WithStreamThinking(true)
	
	var opts llms.CallOptions
	option(&opts)
	
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if !config.StreamThinking {
		t.Error("StreamThinking should be true")
	}
}

func TestWithInterleaveThinking(t *testing.T) {
	option := llms.WithInterleaveThinking(true)
	
	var opts llms.CallOptions
	option(&opts)
	
	config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig)
	if !ok {
		t.Fatal("thinking_config should be stored in metadata")
	}
	
	if !config.InterleaveThinking {
		t.Error("InterleaveThinking should be true")
	}
}

func TestIsReasoningModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		// OpenAI reasoning models
		{"gpt-5-mini", true},
		{"gpt-5-preview", true},
		{"o1-preview", true},
		{"o1-mini", true},
		{"o3-mini", true},
		{"o3-2025-04-16", true},
		{"o4-mini", true},
		
		// Anthropic extended thinking models
		{"claude-3-7-sonnet", true},
		{"claude-3.7-sonnet", true},
		{"claude-4", true},
		{"claude-opus-4", true},
		{"claude-sonnet-4", true},
		
		// DeepSeek reasoner
		{"deepseek-reasoner", true},
		{"deepseek-r1", true},
		
		// Grok reasoning
		{"grok-reasoning", true},
		
		// Non-reasoning models
		{"gpt-4", false},
		{"gpt-3.5-turbo", false},
		{"claude-3-sonnet", false},
		{"claude-2", false},
		{"llama-2", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := llms.IsReasoningModel(tt.model)
			if result != tt.expected {
				t.Errorf("IsReasoningModel(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}

func TestCalculateThinkingBudget(t *testing.T) {
	tests := []struct {
		mode      llms.ThinkingMode
		maxTokens int
		expected  int
	}{
		{llms.ThinkingModeLow, 1000, 200},      // 20%
		{llms.ThinkingModeMedium, 1000, 500},   // 50%
		{llms.ThinkingModeHigh, 1000, 800},     // 80%
		{llms.ThinkingModeAuto, 1000, 0},       // Let model decide
		{llms.ThinkingModeNone, 1000, 0},       // No thinking
		{llms.ThinkingModeLow, 5000, 1000},     // 20% of 5000
		{llms.ThinkingModeMedium, 5000, 2500},  // 50% of 5000
		{llms.ThinkingModeHigh, 5000, 4000},    // 80% of 5000
	}
	
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			result := llms.CalculateThinkingBudget(tt.mode, tt.maxTokens)
			if result != tt.expected {
				t.Errorf("CalculateThinkingBudget(%s, %d) = %d, want %d",
					tt.mode, tt.maxTokens, result, tt.expected)
			}
		})
	}
}

func TestExtractThinkingTokens(t *testing.T) {
	tests := []struct {
		name          string
		generationInfo map[string]any
		expected      *llms.ThinkingTokenUsage
	}{
		{
			name: "OpenAI reasoning tokens",
			generationInfo: map[string]any{
				"ReasoningTokens":           500,
				"CompletionReasoningTokens": 300,
			},
			expected: &llms.ThinkingTokenUsage{
				ThinkingTokens:       500,
				ThinkingOutputTokens: 300,
			},
		},
		{
			name: "Anthropic thinking tokens",
			generationInfo: map[string]any{
				"ThinkingTokens":          1000,
				"ThinkingInputTokens":     200,
				"ThinkingOutputTokens":    800,
				"ThinkingBudgetUsed":      1000,
				"ThinkingBudgetAllocated": 2000,
			},
			expected: &llms.ThinkingTokenUsage{
				ThinkingTokens:          1000,
				ThinkingInputTokens:     200,
				ThinkingOutputTokens:    800,
				ThinkingBudgetUsed:      1000,
				ThinkingBudgetAllocated: 2000,
			},
		},
		{
			name:           "nil generationInfo",
			generationInfo: nil,
			expected:       nil,
		},
		{
			name:           "empty generationInfo",
			generationInfo: map[string]any{},
			expected:       &llms.ThinkingTokenUsage{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llms.ExtractThinkingTokens(tt.generationInfo)
			
			if tt.expected == nil {
				if result != nil {
					t.Error("expected nil result")
				}
				return
			}
			
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			
			if result.ThinkingTokens != tt.expected.ThinkingTokens {
				t.Errorf("ThinkingTokens = %d, want %d",
					result.ThinkingTokens, tt.expected.ThinkingTokens)
			}
			
			if result.ThinkingInputTokens != tt.expected.ThinkingInputTokens {
				t.Errorf("ThinkingInputTokens = %d, want %d",
					result.ThinkingInputTokens, tt.expected.ThinkingInputTokens)
			}
			
			if result.ThinkingOutputTokens != tt.expected.ThinkingOutputTokens {
				t.Errorf("ThinkingOutputTokens = %d, want %d",
					result.ThinkingOutputTokens, tt.expected.ThinkingOutputTokens)
			}
			
			if result.ThinkingBudgetUsed != tt.expected.ThinkingBudgetUsed {
				t.Errorf("ThinkingBudgetUsed = %d, want %d",
					result.ThinkingBudgetUsed, tt.expected.ThinkingBudgetUsed)
			}
			
			if result.ThinkingBudgetAllocated != tt.expected.ThinkingBudgetAllocated {
				t.Errorf("ThinkingBudgetAllocated = %d, want %d",
					result.ThinkingBudgetAllocated, tt.expected.ThinkingBudgetAllocated)
			}
		})
	}
}

// MockLLMWithThinking is a mock LLM that supports thinking tokens
type MockLLMWithThinking struct {
	thinkingConfig  *llms.ThinkingConfig
	model           string
	supportsThinking bool
}

func (m *MockLLMWithThinking) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "test response with thinking", nil
}

func (m *MockLLMWithThinking) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	var opts llms.CallOptions
	for _, opt := range options {
		opt(&opts)
	}
	
	// Extract thinking config
	if opts.Metadata != nil {
		if config, ok := opts.Metadata["thinking_config"].(*llms.ThinkingConfig); ok {
			m.thinkingConfig = config
		}
	}
	
	generationInfo := map[string]any{
		"CompletionTokens": 100,
		"PromptTokens":     200,
		"TotalTokens":      300,
	}
	
	// Add thinking tokens if configured
	if m.thinkingConfig != nil && m.thinkingConfig.Mode != llms.ThinkingModeNone {
		budget := llms.CalculateThinkingBudget(m.thinkingConfig.Mode, 1000)
		generationInfo["ReasoningTokens"] = budget
		generationInfo["ThinkingBudgetAllocated"] = budget
		generationInfo["ThinkingBudgetUsed"] = budget * 80 / 100 // Use 80% of budget
	}
	
	content := "test response"
	if m.thinkingConfig != nil && m.thinkingConfig.ReturnThinking {
		content = "<thinking>step by step reasoning</thinking>\n" + content
	}
	
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:        content,
				GenerationInfo: generationInfo,
			},
		},
	}, nil
}

// SupportsReasoning implements the ReasoningModel interface.
func (m *MockLLMWithThinking) SupportsReasoning() bool {
	return m.supportsThinking
}

func TestThinkingIntegration(t *testing.T) {
	llm := &MockLLMWithThinking{supportsThinking: true}
	ctx := context.Background()
	
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("test")},
		},
	}
	
	t.Run("with thinking mode high", func(t *testing.T) {
		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithThinkingMode(llms.ThinkingModeHigh),
			llms.WithReturnThinking(true),
		)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if llm.thinkingConfig == nil {
			t.Fatal("thinking config should be set")
		}
		
		if llm.thinkingConfig.Mode != llms.ThinkingModeHigh {
			t.Errorf("mode = %s, want %s", llm.thinkingConfig.Mode, llms.ThinkingModeHigh)
		}
		
		if !llm.thinkingConfig.ReturnThinking {
			t.Error("ReturnThinking should be true")
		}
		
		// Check response includes thinking
		if resp.Choices[0].Content == "" {
			t.Error("content should not be empty")
		}
		
		// Extract thinking tokens
		usage := llms.ExtractThinkingTokens(resp.Choices[0].GenerationInfo)
		if usage == nil {
			t.Fatal("thinking token usage should be extracted")
		}
		
		if usage.ThinkingTokens == 0 {
			t.Error("ThinkingTokens should be > 0")
		}
	})
	
	t.Run("with explicit budget", func(t *testing.T) {
		_, err := llm.GenerateContent(ctx, messages,
			llms.WithThinkingBudget(4096),
		)
		
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		
		if llm.thinkingConfig == nil {
			t.Fatal("thinking config should be set")
		}
		
		if llm.thinkingConfig.BudgetTokens != 4096 {
			t.Errorf("BudgetTokens = %d, want 4096", llm.thinkingConfig.BudgetTokens)
		}
	})
}

func TestSupportsReasoningModel(t *testing.T) {
	tests := []struct {
		name     string
		llm      interface{}
		expected bool
	}{
		{
			name:     "LLM with reasoning support",
			llm:      &MockLLMWithThinking{supportsThinking: true},
			expected: true,
		},
		{
			name:     "LLM without reasoning support",
			llm:      &MockLLMWithThinking{supportsThinking: false},
			expected: false,
		},
		{
			name:     "Non-reasoning LLM",
			llm:      struct{}{},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := llms.SupportsReasoningModel(tt.llm)
			if result != tt.expected {
				t.Errorf("SupportsReasoningModel() = %v, want %v", result, tt.expected)
			}
		})
	}
}