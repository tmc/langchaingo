package llms_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
)

func TestCallOptions(t *testing.T) { //nolint:funlen // comprehensive test
	tests := []struct {
		name   string
		option llms.CallOption
		verify func(t *testing.T, opts llms.CallOptions)
	}{
		{
			name:   "WithModel",
			option: llms.WithModel("gpt-4"),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.Model != "gpt-4" {
					t.Errorf("Model = %v, want %v", opts.Model, "gpt-4")
				}
			},
		},
		{
			name:   "WithMaxTokens",
			option: llms.WithMaxTokens(100),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.MaxTokens != 100 {
					t.Errorf("MaxTokens = %v, want %v", opts.MaxTokens, 100)
				}
			},
		},
		{
			name:   "WithCandidateCount",
			option: llms.WithCandidateCount(3),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.CandidateCount != 3 {
					t.Errorf("CandidateCount = %v, want %v", opts.CandidateCount, 3)
				}
			},
		},
		{
			name:   "WithTemperature",
			option: llms.WithTemperature(0.7),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.Temperature != 0.7 {
					t.Errorf("Temperature = %v, want %v", opts.Temperature, 0.7)
				}
			},
		},
		{
			name:   "WithStopWords",
			option: llms.WithStopWords([]string{"STOP", "END"}),
			verify: func(t *testing.T, opts llms.CallOptions) {
				expected := []string{"STOP", "END"}
				if !reflect.DeepEqual(opts.StopWords, expected) {
					t.Errorf("StopWords = %v, want %v", opts.StopWords, expected)
				}
			},
		},
		{
			name:   "WithTopK",
			option: llms.WithTopK(50),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.TopK != 50 {
					t.Errorf("TopK = %v, want %v", opts.TopK, 50)
				}
			},
		},
		{
			name:   "WithTopP",
			option: llms.WithTopP(0.9),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.TopP != 0.9 {
					t.Errorf("TopP = %v, want %v", opts.TopP, 0.9)
				}
			},
		},
		{
			name:   "WithSeed",
			option: llms.WithSeed(42),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.Seed != 42 {
					t.Errorf("Seed = %v, want %v", opts.Seed, 42)
				}
			},
		},
		{
			name:   "WithMinLength",
			option: llms.WithMinLength(10),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.MinLength != 10 {
					t.Errorf("MinLength = %v, want %v", opts.MinLength, 10)
				}
			},
		},
		{
			name:   "WithMaxLength",
			option: llms.WithMaxLength(200),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.MaxLength != 200 {
					t.Errorf("MaxLength = %v, want %v", opts.MaxLength, 200)
				}
			},
		},
		{
			name:   "WithN",
			option: llms.WithN(5),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.N != 5 {
					t.Errorf("N = %v, want %v", opts.N, 5)
				}
			},
		},
		{
			name:   "WithRepetitionPenalty",
			option: llms.WithRepetitionPenalty(1.2),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.RepetitionPenalty != 1.2 {
					t.Errorf("RepetitionPenalty = %v, want %v", opts.RepetitionPenalty, 1.2)
				}
			},
		},
		{
			name:   "WithFrequencyPenalty",
			option: llms.WithFrequencyPenalty(0.5),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.FrequencyPenalty != 0.5 {
					t.Errorf("FrequencyPenalty = %v, want %v", opts.FrequencyPenalty, 0.5)
				}
			},
		},
		{
			name:   "WithPresencePenalty",
			option: llms.WithPresencePenalty(0.6),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.PresencePenalty != 0.6 {
					t.Errorf("PresencePenalty = %v, want %v", opts.PresencePenalty, 0.6)
				}
			},
		},
		{
			name:   "WithJSONMode",
			option: llms.WithJSONMode(),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if !opts.JSONMode {
					t.Error("JSONMode = false, want true")
				}
			},
		},
		{
			name:   "WithResponseMIMEType",
			option: llms.WithResponseMIMEType("application/json"),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if *opts.ResponseMIMEType != "application/json" {
					t.Errorf("ResponseMIMEType = %v, want %v", opts.ResponseMIMEType, "application/json")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts llms.CallOptions
			tt.option(&opts)
			tt.verify(t, opts)
		})
	}
}

func TestWithOptions(t *testing.T) {
	baseOptions := llms.CallOptions{
		Model:       getRefString("gpt-3.5-turbo"),
		MaxTokens:   getRefInt(150),
		Temperature: getRefFloat64(0.8),
		TopK:        getRefInt(40),
		TopP:        getRefFloat64(0.95),
		Seed:        getRefInt(123),
		N:           getRefInt(2),
	}

	var opts llms.CallOptions
	llms.WithOptions(baseOptions)(&opts)

	if !reflect.DeepEqual(opts, baseOptions) {
		t.Errorf("WithOptions did not copy all fields correctly\ngot:  %+v\nwant: %+v", opts, baseOptions)
	}
}

func TestWithStreamingFunc(t *testing.T) { //nolint:funlen
	var (
		called       int
		gotDone      bool
		gotReasoning *reasoning.ContentReasoning
		gotChunk     string
		gotToolCall  streaming.ToolCall
	)
	testFunc := func(ctx context.Context, chunk streaming.Chunk) error {
		called++
		switch chunk.Type {
		case streaming.ChunkTypeReasoning:
			gotReasoning = chunk.Reasoning
		case streaming.ChunkTypeText:
			gotChunk = chunk.Content
		case streaming.ChunkTypeToolCall:
			gotToolCall = chunk.ToolCall
		case streaming.ChunkTypeDone:
			gotDone = true
		default:
			return fmt.Errorf("unexpected chunk type: %s", chunk.Type)
		}
		return nil
	}

	var opts llms.CallOptions
	llms.WithStreamingFunc(testFunc)(&opts)

	if opts.StreamingFunc == nil {
		t.Error("StreamingFunc was not set")
	}

	ctx := t.Context()

	// Test that the function works
	reasoning := &reasoning.ContentReasoning{Content: "reasoning"}
	if err := opts.StreamingFunc(ctx, streaming.NewReasoningChunk(reasoning)); err != nil {
		t.Errorf("StreamingFunc with reasoning content returned error: %v", err)
	}
	chunk := "chunk"
	if err := opts.StreamingFunc(ctx, streaming.NewTextChunk(chunk)); err != nil {
		t.Errorf("StreamingFunc with text chunk returned error: %v", err)
	}
	toolCall := streaming.ToolCall{
		ID:        "123",
		Name:      "test",
		Arguments: "{}",
	}
	if err := opts.StreamingFunc(ctx, streaming.NewToolCallChunk(toolCall)); err != nil {
		t.Errorf("StreamingFunc with tool call chunk returned error: %v", err)
	}
	if err := opts.StreamingFunc(ctx, streaming.NewDoneChunk()); err != nil {
		t.Errorf("StreamingFunc with done chunk returned error: %v", err)
	}

	if called != 4 {
		t.Errorf("StreamingFunc was not called 4 times, got %d", called)
	}
	if !gotDone {
		t.Error("StreamingFunc was not called with done chunk")
	}
	if gotReasoning.String() != reasoning.String() {
		t.Errorf("StreamingFunc reasoning = %s, want %s", gotReasoning.String(), reasoning.String())
	}
	if gotChunk != chunk {
		t.Errorf("StreamingFunc chunk = %s, want %s", gotChunk, chunk)
	}
	if !reflect.DeepEqual(gotToolCall, toolCall) {
		t.Errorf("StreamingFunc tool call = %v, want %v", gotToolCall, toolCall)
	}
}

func TestWithMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"key4": []string{"a", "b", "c"},
	}

	var opts llms.CallOptions
	llms.WithMetadata(metadata)(&opts)

	if !reflect.DeepEqual(opts.Metadata, metadata) {
		t.Errorf("Metadata = %v, want %v", opts.Metadata, metadata)
	}
}

func TestWithTools(t *testing.T) {
	tools := []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_time",
				Description: "Get the current time",
				Strict:      true,
			},
		},
	}

	var opts llms.CallOptions
	llms.WithTools(tools)(&opts)

	if len(opts.Tools) != len(tools) {
		t.Fatalf("Tools length = %v, want %v", len(opts.Tools), len(tools))
	}

	for i, tool := range tools {
		if opts.Tools[i].Type != tool.Type {
			t.Errorf("Tool[%d].Type = %v, want %v", i, opts.Tools[i].Type, tool.Type)
		}
		if opts.Tools[i].Function.Name != tool.Function.Name {
			t.Errorf("Tool[%d].Function.Name = %v, want %v", i, opts.Tools[i].Function.Name, tool.Function.Name)
		}
		if opts.Tools[i].Function.Description != tool.Function.Description {
			t.Errorf("Tool[%d].Function.Description = %v, want %v", i, opts.Tools[i].Function.Description, tool.Function.Description)
		}
		if opts.Tools[i].Function.Strict != tool.Function.Strict {
			t.Errorf("Tool[%d].Function.Strict = %v, want %v", i, opts.Tools[i].Function.Strict, tool.Function.Strict)
		}
	}
}

func TestWithToolChoice(t *testing.T) {
	tests := []struct {
		name   string
		choice any
	}{
		{
			name:   "string choice",
			choice: "auto",
		},
		{
			name:   "none choice",
			choice: "none",
		},
		{
			name: "specific tool choice",
			choice: llms.ToolChoice{
				Type: "function",
				Function: &llms.FunctionReference{
					Name: "get_weather",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts llms.CallOptions
			llms.WithToolChoice(tt.choice)(&opts)

			if !reflect.DeepEqual(opts.ToolChoice, tt.choice) {
				t.Errorf("ToolChoice = %v, want %v", opts.ToolChoice, tt.choice)
			}
		})
	}
}

func TestDeprecatedFunctionOptions(t *testing.T) {
	// Test WithFunctionCallBehavior
	t.Run("WithFunctionCallBehavior", func(t *testing.T) {
		var opts llms.CallOptions
		llms.WithFunctionCallBehavior(llms.FunctionCallBehaviorAuto)(&opts)

		if opts.FunctionCallBehavior != llms.FunctionCallBehaviorAuto {
			t.Errorf("FunctionCallBehavior = %v, want %v", opts.FunctionCallBehavior, llms.FunctionCallBehaviorAuto)
		}

		// Test with None behavior
		opts = llms.CallOptions{}
		llms.WithFunctionCallBehavior(llms.FunctionCallBehaviorNone)(&opts)

		if opts.FunctionCallBehavior != llms.FunctionCallBehaviorNone {
			t.Errorf("FunctionCallBehavior = %v, want %v", opts.FunctionCallBehavior, llms.FunctionCallBehaviorNone)
		}
	})

	// Test WithFunctions
	t.Run("WithFunctions", func(t *testing.T) {
		functions := []llms.FunctionDefinition{
			{
				Name:        "get_weather",
				Description: "Get weather information",
				Parameters: map[string]interface{}{
					"location": "string",
				},
			},
			{
				Name:        "calculate",
				Description: "Perform calculations",
				Parameters: map[string]interface{}{
					"expression": "string",
				},
			},
		}

		var opts llms.CallOptions
		llms.WithFunctions(functions)(&opts)

		if len(opts.Functions) != len(functions) {
			t.Fatalf("Functions length = %v, want %v", len(opts.Functions), len(functions))
		}

		for i, fn := range functions {
			if opts.Functions[i].Name != fn.Name {
				t.Errorf("Functions[%d].Name = %v, want %v", i, opts.Functions[i].Name, fn.Name)
			}
			if opts.Functions[i].Description != fn.Description {
				t.Errorf("Functions[%d].Description = %v, want %v", i, opts.Functions[i].Description, fn.Description)
			}
		}
	})
}

func TestMultipleOptions(t *testing.T) {
	var opts llms.CallOptions

	// Apply multiple options
	options := []llms.CallOption{
		llms.WithModel("gpt-4"),
		llms.WithMaxTokens(200),
		llms.WithTemperature(0.5),
		llms.WithTopK(30),
		llms.WithTopP(0.8),
		llms.WithStopWords([]string{"END"}),
		llms.WithJSONMode(),
		llms.WithN(3),
	}

	for _, opt := range options {
		opt(&opts)
	}

	// Verify all options were applied
	if *opts.Model != "gpt-4" {
		t.Errorf("Model = %v, want %v", opts.Model, "gpt-4")
	}
	if *opts.MaxTokens != 200 {
		t.Errorf("MaxTokens = %v, want %v", opts.MaxTokens, 200)
	}
	if *opts.Temperature != 0.5 {
		t.Errorf("Temperature = %v, want %v", opts.Temperature, 0.5)
	}
	if *opts.TopK != 30 {
		t.Errorf("TopK = %v, want %v", opts.TopK, 30)
	}
	if *opts.TopP != 0.8 {
		t.Errorf("TopP = %v, want %v", opts.TopP, 0.8)
	}
	if len(opts.StopWords) != 1 || opts.StopWords[0] != "END" {
		t.Errorf("StopWords = %v, want %v", opts.StopWords, []string{"END"})
	}
	if !opts.JSONMode {
		t.Error("JSONMode = false, want true")
	}
	if *opts.N != 3 {
		t.Errorf("N = %v, want %v", opts.N, 3)
	}
}

func TestStreamingFuncError(t *testing.T) {
	testErr := errors.New("streaming error")
	testFunc := func(ctx context.Context, chunk streaming.Chunk) error {
		return testErr
	}

	var opts llms.CallOptions
	llms.WithStreamingFunc(testFunc)(&opts)

	ctx := t.Context()

	err := opts.StreamingFunc(ctx, streaming.NewTextChunk("test"))
	if !errors.Is(err, testErr) {
		t.Errorf("StreamingFunc error = %v, want %v", err, testErr)
	}
}

func TestEmptyOptions(t *testing.T) {
	var opts llms.CallOptions

	// Verify default values
	if opts.Model != nil {
		t.Errorf("Model = %v, want empty string", opts.Model)
	}
	if opts.MaxTokens != nil {
		t.Errorf("MaxTokens = %v, want 0", opts.MaxTokens)
	}
	if opts.Temperature != nil {
		t.Errorf("Temperature = %v, want 0", opts.Temperature)
	}
	if opts.JSONMode {
		t.Error("JSONMode = true, want false")
	}
	if opts.StreamingFunc != nil {
		t.Error("StreamingFunc is not nil")
	}
	if opts.Tools != nil {
		t.Error("Tools is not nil")
	}
	if opts.Functions != nil {
		t.Error("Functions is not nil")
	}
}

func TestReasoningConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *llms.ReasoningConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty config",
			config:   &llms.ReasoningConfig{},
			expected: false,
		},
		{
			name:     "config with tokens only",
			config:   &llms.ReasoningConfig{Tokens: 1000},
			expected: true,
		},
		{
			name:     "config with effort only",
			config:   &llms.ReasoningConfig{Effort: llms.ReasoningLow},
			expected: true,
		},
		{
			name:     "config with both tokens and effort",
			config:   &llms.ReasoningConfig{Effort: llms.ReasoningMedium, Tokens: 2000},
			expected: true,
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.config.IsEnabled()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestReasoningConfig_GetEffort(t *testing.T) {
	t.Parallel()

	maxTokens := 10000

	tests := []struct {
		name      string
		config    *llms.ReasoningConfig
		maxTokens int
		expected  llms.ReasoningEffort
	}{
		{
			name:      "nil config",
			config:    nil,
			maxTokens: maxTokens,
			expected:  llms.ReasoningNone,
		},
		{
			name:      "empty config",
			config:    &llms.ReasoningConfig{},
			maxTokens: maxTokens,
			expected:  llms.ReasoningNone,
		},
		{
			name:      "config with explicit effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningHigh},
			maxTokens: maxTokens,
			expected:  llms.ReasoningHigh,
		},
		{
			name:      "config with low tokens",
			config:    &llms.ReasoningConfig{Tokens: maxTokens / 5},
			maxTokens: maxTokens,
			expected:  llms.ReasoningLow,
		},
		{
			name:      "config with medium tokens",
			config:    &llms.ReasoningConfig{Tokens: maxTokens/4 + 100}, // Just above low threshold
			maxTokens: maxTokens,
			expected:  llms.ReasoningMedium,
		},
		{
			name:      "config with high tokens",
			config:    &llms.ReasoningConfig{Tokens: maxTokens / 2},
			maxTokens: maxTokens,
			expected:  llms.ReasoningHigh,
		},
		{
			name:      "precedence - effort over tokens",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningLow, Tokens: maxTokens / 2},
			maxTokens: maxTokens,
			expected:  llms.ReasoningLow,
		},
		{
			name:      "negative maxTokens uses default 8192",
			config:    &llms.ReasoningConfig{Tokens: 8192/3 + 10}, // Just above medium threshold for 8192
			maxTokens: -1,
			expected:  llms.ReasoningHigh,
		},
		{
			name:      "zero maxTokens uses default 8192",
			config:    &llms.ReasoningConfig{Tokens: 8192 / 5}, // Low for 8192
			maxTokens: 0,
			expected:  llms.ReasoningLow,
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.config.GetEffort(tc.maxTokens)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestReasoningConfig_GetTokens(t *testing.T) {
	t.Parallel()

	maxTokens := 12000

	tests := []struct {
		name      string
		config    *llms.ReasoningConfig
		maxTokens int
		expected  int
	}{
		{
			name:      "nil config",
			config:    nil,
			maxTokens: maxTokens,
			expected:  0,
		},
		{
			name:      "empty config",
			config:    &llms.ReasoningConfig{},
			maxTokens: maxTokens,
			expected:  0,
		},
		{
			name:      "config with explicit tokens",
			config:    &llms.ReasoningConfig{Tokens: 3000},
			maxTokens: maxTokens,
			expected:  3000,
		},
		{
			name:      "config with low effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningLow},
			maxTokens: maxTokens,
			expected:  max(maxTokens/4, 1024),
		},
		{
			name:      "config with medium effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningMedium},
			maxTokens: maxTokens,
			expected:  max(maxTokens/3, 2048),
		},
		{
			name:      "config with high effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningHigh},
			maxTokens: maxTokens,
			expected:  max(maxTokens/2, 4096),
		},
		{
			name:      "tokens exceeding max reasoning tokens",
			config:    &llms.ReasoningConfig{Tokens: llms.MaxReasoningTokens + 1000},
			maxTokens: maxTokens,
			expected:  min(llms.MaxReasoningTokens, maxTokens*2/3),
		},
		{
			name:      "tokens exceeding 2/3 of max tokens",
			config:    &llms.ReasoningConfig{Tokens: maxTokens},
			maxTokens: maxTokens,
			expected:  maxTokens * 2 / 3,
		},
		{
			name:      "invalid effort",
			config:    &llms.ReasoningConfig{Effort: "invalid"},
			maxTokens: maxTokens,
			expected:  -1,
		},
		{
			name:      "negative maxTokens uses default 8192 for low effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningLow},
			maxTokens: -10,
			expected:  max(8192/4, 1024), // Based on default 8192
		},
		{
			name:      "zero maxTokens uses default 8192 for high effort",
			config:    &llms.ReasoningConfig{Effort: llms.ReasoningHigh},
			maxTokens: 0,
			expected:  max(8192/2, 4096), // Based on default 8192
		},
		{
			name:      "negative maxTokens with explicit tokens",
			config:    &llms.ReasoningConfig{Tokens: 7000},
			maxTokens: -5,
			expected:  min(7000, 8192*2/3), // Using default 8192 as maxTokens
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.config.GetTokens(tc.maxTokens)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWithReasoning(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		effort          llms.ReasoningEffort
		tokens          int
		expectedEffort  llms.ReasoningEffort
		expectedTokens  int
		expectedEnabled bool
	}{
		{
			name:            "high effort",
			effort:          llms.ReasoningHigh,
			tokens:          0,
			expectedEffort:  llms.ReasoningHigh,
			expectedTokens:  0,
			expectedEnabled: true,
		},
		{
			name:            "medium effort",
			effort:          llms.ReasoningMedium,
			tokens:          0,
			expectedEffort:  llms.ReasoningMedium,
			expectedTokens:  0,
			expectedEnabled: true,
		},
		{
			name:            "low effort",
			effort:          llms.ReasoningLow,
			tokens:          0,
			expectedEffort:  llms.ReasoningLow,
			expectedTokens:  0,
			expectedEnabled: true,
		},
		{
			name:            "specific tokens",
			effort:          llms.ReasoningNone,
			tokens:          5000,
			expectedEffort:  llms.ReasoningNone,
			expectedTokens:  5000,
			expectedEnabled: true,
		},
		{
			name:            "both effort and tokens",
			effort:          llms.ReasoningHigh,
			tokens:          3000,
			expectedEffort:  llms.ReasoningHigh,
			expectedTokens:  3000,
			expectedEnabled: true,
		},
		{
			name:            "disabled reasoning",
			effort:          llms.ReasoningNone,
			tokens:          0,
			expectedEffort:  llms.ReasoningNone,
			expectedTokens:  0,
			expectedEnabled: false,
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts := &llms.CallOptions{}
			llms.WithReasoning(tc.effort, tc.tokens)(opts)

			assert.NotNil(t, opts.Reasoning)
			assert.Equal(t, tc.expectedEffort, opts.Reasoning.Effort)
			assert.Equal(t, tc.expectedTokens, opts.Reasoning.Tokens)
			assert.Equal(t, tc.expectedEnabled, opts.Reasoning.IsEnabled())
		})
	}
}

func TestReasoningConfig_Integration(t *testing.T) {
	t.Parallel()

	const maxTokens = 16000

	tests := []struct {
		name            string
		config          *llms.ReasoningConfig
		maxTokens       int
		expectedEnabled bool
		expectedEffort  llms.ReasoningEffort
		expectedTokens  int
	}{
		{
			name:            "high effort conversion to tokens",
			config:          &llms.ReasoningConfig{Effort: llms.ReasoningHigh},
			maxTokens:       maxTokens,
			expectedEnabled: true,
			expectedEffort:  llms.ReasoningHigh,
			expectedTokens:  8000, // max(maxTokens/2, 4096)
		},
		{
			name:            "tokens conversion to effort level",
			config:          &llms.ReasoningConfig{Tokens: 2500},
			maxTokens:       maxTokens,
			expectedEnabled: true,
			expectedEffort:  llms.ReasoningLow, // tokens < maxTokens/4 (4000)
			expectedTokens:  2500,
		},
		{
			name:            "cap excessive tokens",
			config:          &llms.ReasoningConfig{Tokens: maxTokens},
			maxTokens:       maxTokens,
			expectedEnabled: true,
			expectedEffort:  llms.ReasoningHigh, // tokens > maxTokens/3
			expectedTokens:  maxTokens * 2 / 3,
		},
		{
			name:            "default maxTokens handling",
			config:          &llms.ReasoningConfig{Effort: llms.ReasoningMedium},
			maxTokens:       0, // Should use default 8192
			expectedEnabled: true,
			expectedEffort:  llms.ReasoningMedium,
			expectedTokens:  max(8192/3, 2048), // Based on default 8192
		},
	}

	for idx := range tests {
		tc := tests[idx]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			isEnabled := tc.config.IsEnabled()
			effort := tc.config.GetEffort(tc.maxTokens)
			tokens := tc.config.GetTokens(tc.maxTokens)

			assert.Equal(t, tc.expectedEnabled, isEnabled)
			assert.Equal(t, tc.expectedEffort, effort)
			assert.Equal(t, tc.expectedTokens, tokens)
		})
	}
}

func getRefString(s string) *string {
	return &s
}

func getRefInt(i int) *int {
	return &i
}

func getRefFloat64(f float64) *float64 {
	return &f
}
