package yzma

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

// To run tests, set YZMA_TEST_MODEL to the location with the test model.
// For example:
// 		export YZMA_TEST_MODEL=/home/username/models/SmolLM2-135M-Instruct.Q2_K.gguf
// https://huggingface.co/QuantFactory/SmolLM2-135M-Instruct-GGUF/resolve/main/SmolLM2-135M-Instruct.Q2_K.gguf

func TestNew(t *testing.T) {
	if testModel := os.Getenv("YZMA_TEST_MODEL"); testModel == "" {
		t.Fatal("YZMA_TEST_MODEL not set to point to test model")
	}

	tests := []struct {
		name    string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "with model path",
			opts:    []Option{WithModel("some-model-path")},
			wantErr: false,
		},
		{
			name:    "without model",
			opts:    []Option{},
			wantErr: false,
		},
		{
			name:    "with system prompt",
			opts:    []Option{WithSystemPrompt("some-system-prompt")},
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
				t.Error("New() returned nil LLM without error")
			}
			if llm != nil {
				llm.Close()
			}
		})
	}
}

func TestCall(t *testing.T) {
	testModel := os.Getenv("YZMA_TEST_MODEL")
	if testModel == "" {
		t.Fatal("YZMA_TEST_MODEL not set to point to test model")
	}

	llm, err := New(WithModel(testModel))
	if err != nil {
		t.Fatalf("failed to create LLM: %v", err)
	}
	defer llm.Close()

	response, err := llm.Call(context.Background(), "How many feet are in a nautical mile?",
		llms.WithTemperature(0.8),
		llms.WithTopK(40),
		llms.WithTopP(0.9),
		llms.WithMaxTokens(50))

	if err != nil {
		t.Fatalf("Call() error: %v", err)
	}

	expected := "To find out how many feet are in a nautical mile, you can use the formula: 1 nautical mile = 1.875 nautical miles"
	if !strings.Contains(response, expected) {
		t.Errorf("expected %s, got %q", expected, response)
	}
}
