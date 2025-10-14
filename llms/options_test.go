package llms_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/vendasta/langchaingo/llms"
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
				if opts.Model != "gpt-4" {
					t.Errorf("Model = %v, want %v", opts.Model, "gpt-4")
				}
			},
		},
		{
			name:   "WithMaxTokens",
			option: llms.WithMaxTokens(100),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.MaxTokens != 100 {
					t.Errorf("MaxTokens = %v, want %v", opts.MaxTokens, 100)
				}
			},
		},
		{
			name:   "WithCandidateCount",
			option: llms.WithCandidateCount(3),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.CandidateCount != 3 {
					t.Errorf("CandidateCount = %v, want %v", opts.CandidateCount, 3)
				}
			},
		},
		{
			name:   "WithTemperature",
			option: llms.WithTemperature(0.7),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.Temperature != 0.7 {
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
				if opts.TopK != 50 {
					t.Errorf("TopK = %v, want %v", opts.TopK, 50)
				}
			},
		},
		{
			name:   "WithTopP",
			option: llms.WithTopP(0.9),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.TopP != 0.9 {
					t.Errorf("TopP = %v, want %v", opts.TopP, 0.9)
				}
			},
		},
		{
			name:   "WithSeed",
			option: llms.WithSeed(42),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.Seed != 42 {
					t.Errorf("Seed = %v, want %v", opts.Seed, 42)
				}
			},
		},
		{
			name:   "WithMinLength",
			option: llms.WithMinLength(10),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.MinLength != 10 {
					t.Errorf("MinLength = %v, want %v", opts.MinLength, 10)
				}
			},
		},
		{
			name:   "WithMaxLength",
			option: llms.WithMaxLength(200),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.MaxLength != 200 {
					t.Errorf("MaxLength = %v, want %v", opts.MaxLength, 200)
				}
			},
		},
		{
			name:   "WithN",
			option: llms.WithN(5),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.N != 5 {
					t.Errorf("N = %v, want %v", opts.N, 5)
				}
			},
		},
		{
			name:   "WithRepetitionPenalty",
			option: llms.WithRepetitionPenalty(1.2),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.RepetitionPenalty != 1.2 {
					t.Errorf("RepetitionPenalty = %v, want %v", opts.RepetitionPenalty, 1.2)
				}
			},
		},
		{
			name:   "WithFrequencyPenalty",
			option: llms.WithFrequencyPenalty(0.5),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.FrequencyPenalty != 0.5 {
					t.Errorf("FrequencyPenalty = %v, want %v", opts.FrequencyPenalty, 0.5)
				}
			},
		},
		{
			name:   "WithPresencePenalty",
			option: llms.WithPresencePenalty(0.6),
			verify: func(t *testing.T, opts llms.CallOptions) {
				if opts.PresencePenalty != 0.6 {
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
				if opts.ResponseMIMEType != "application/json" {
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
		Model:       "gpt-3.5-turbo",
		MaxTokens:   150,
		Temperature: 0.8,
		TopK:        40,
		TopP:        0.95,
		Seed:        123,
		N:           2,
	}

	var opts llms.CallOptions
	llms.WithOptions(baseOptions)(&opts)

	if !reflect.DeepEqual(opts, baseOptions) {
		t.Errorf("WithOptions did not copy all fields correctly\ngot:  %+v\nwant: %+v", opts, baseOptions)
	}
}

func TestWithStreamingFunc(t *testing.T) {
	called := false
	testFunc := func(ctx context.Context, chunk []byte) error {
		called = true
		return nil
	}

	var opts llms.CallOptions
	llms.WithStreamingFunc(testFunc)(&opts)

	if opts.StreamingFunc == nil {
		t.Error("StreamingFunc was not set")
	}

	// Test that the function works
	err := opts.StreamingFunc(context.Background(), []byte("test"))
	if err != nil {
		t.Errorf("StreamingFunc returned error: %v", err)
	}
	if !called {
		t.Error("StreamingFunc was not called")
	}
}

func TestWithStreamingReasoningFunc(t *testing.T) {
	called := false
	var gotReasoning, gotChunk []byte
	testFunc := func(ctx context.Context, reasoningChunk, chunk []byte) error {
		called = true
		gotReasoning = reasoningChunk
		gotChunk = chunk
		return nil
	}

	var opts llms.CallOptions
	llms.WithStreamingReasoningFunc(testFunc)(&opts)

	if opts.StreamingReasoningFunc == nil {
		t.Error("StreamingReasoningFunc was not set")
	}

	// Test that the function works
	reasoning := []byte("reasoning")
	chunk := []byte("chunk")
	err := opts.StreamingReasoningFunc(context.Background(), reasoning, chunk)
	if err != nil {
		t.Errorf("StreamingReasoningFunc returned error: %v", err)
	}
	if !called {
		t.Error("StreamingReasoningFunc was not called")
	}
	if !reflect.DeepEqual(gotReasoning, reasoning) {
		t.Errorf("StreamingReasoningFunc reasoning = %v, want %v", gotReasoning, reasoning)
	}
	if !reflect.DeepEqual(gotChunk, chunk) {
		t.Errorf("StreamingReasoningFunc chunk = %v, want %v", gotChunk, chunk)
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
	if opts.Model != "gpt-4" {
		t.Errorf("Model = %v, want %v", opts.Model, "gpt-4")
	}
	if opts.MaxTokens != 200 {
		t.Errorf("MaxTokens = %v, want %v", opts.MaxTokens, 200)
	}
	if opts.Temperature != 0.5 {
		t.Errorf("Temperature = %v, want %v", opts.Temperature, 0.5)
	}
	if opts.TopK != 30 {
		t.Errorf("TopK = %v, want %v", opts.TopK, 30)
	}
	if opts.TopP != 0.8 {
		t.Errorf("TopP = %v, want %v", opts.TopP, 0.8)
	}
	if len(opts.StopWords) != 1 || opts.StopWords[0] != "END" {
		t.Errorf("StopWords = %v, want %v", opts.StopWords, []string{"END"})
	}
	if !opts.JSONMode {
		t.Error("JSONMode = false, want true")
	}
	if opts.N != 3 {
		t.Errorf("N = %v, want %v", opts.N, 3)
	}
}

func TestStreamingFuncError(t *testing.T) {
	testErr := errors.New("streaming error")
	testFunc := func(ctx context.Context, chunk []byte) error {
		return testErr
	}

	var opts llms.CallOptions
	llms.WithStreamingFunc(testFunc)(&opts)

	err := opts.StreamingFunc(context.Background(), []byte("test"))
	if err != testErr {
		t.Errorf("StreamingFunc error = %v, want %v", err, testErr)
	}
}

func TestEmptyOptions(t *testing.T) {
	var opts llms.CallOptions

	// Verify default values
	if opts.Model != "" {
		t.Errorf("Model = %v, want empty string", opts.Model)
	}
	if opts.MaxTokens != 0 {
		t.Errorf("MaxTokens = %v, want 0", opts.MaxTokens)
	}
	if opts.Temperature != 0 {
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
