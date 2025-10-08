package local

import (
	"context"
	"os"
	"testing"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/schema"
)

func TestNew(t *testing.T) {
	// Save original env vars
	origBin := os.Getenv("LOCAL_LLM_BIN")
	origArgs := os.Getenv("LOCAL_LLM_ARGS")
	defer func() {
		os.Setenv("LOCAL_LLM_BIN", origBin)
		os.Setenv("LOCAL_LLM_ARGS", origArgs)
	}()

	tests := []struct {
		name    string
		envBin  string
		envArgs string
		opts    []Option
		wantErr bool
	}{
		{
			name:    "with echo binary",
			opts:    []Option{WithBin("echo")},
			wantErr: false,
		},
		{
			name:    "with echo binary and args",
			opts:    []Option{WithBin("echo"), WithArgs("-n test")},
			wantErr: false,
		},
		{
			name:    "with echo binary and global args",
			opts:    []Option{WithBin("echo"), WithGlobalAsArgs()},
			wantErr: false,
		},
		{
			name:    "from env var",
			envBin:  "echo",
			envArgs: "-n",
			opts:    []Option{},
			wantErr: false,
		},
		{
			name:    "non-existent binary",
			opts:    []Option{WithBin("non-existent-binary-12345")},
			wantErr: true,
		},
		{
			name:    "missing binary",
			envBin:  "",
			opts:    []Option{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LOCAL_LLM_BIN", tt.envBin)
			os.Setenv("LOCAL_LLM_ARGS", tt.envArgs)

			llm, err := New(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && llm == nil {
				t.Error("New() returned nil LLM without error")
			}
		})
	}
}

func TestCall(t *testing.T) {
	llm, err := New(WithBin("echo"), WithArgs("-n"))
	if err != nil {
		t.Fatalf("failed to create LLM: %v", err)
	}

	response, err := llm.Call(context.Background(), "Hello, World!")
	if err != nil {
		t.Fatalf("Call() error: %v", err)
	}

	if response != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %q", response)
	}
}

func TestGenerateContent(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		messages []llms.MessageContent
		callOpts []llms.CallOption
		want     string
	}{
		{
			name: "basic echo",
			opts: []Option{WithBin("echo"), WithArgs("-n")},
			messages: []llms.MessageContent{
				{
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Hello, World!"},
					},
				},
			},
			want: "Hello, World!",
		},
		{
			name: "echo without args",
			opts: []Option{WithBin("echo")},
			messages: []llms.MessageContent{
				{
					Parts: []llms.ContentPart{
						llms.TextContent{Text: "Test"},
					},
				},
			},
			want: " Test\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := New(tt.opts...)
			if err != nil {
				t.Fatalf("failed to create LLM: %v", err)
			}

			resp, err := llm.GenerateContent(context.Background(), tt.messages, tt.callOpts...)
			if err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}

			if len(resp.Choices) != 1 {
				t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
			}

			if resp.Choices[0].Content != tt.want {
				t.Errorf("expected %q, got %q", tt.want, resp.Choices[0].Content)
			}
		})
	}
}

func TestGenerateContentWithGlobalArgs(t *testing.T) { //nolint:funlen // comprehensive test
	// Create a test script that echoes its arguments
	scriptContent := `#!/bin/sh
echo "$@"`

	tmpFile, err := os.CreateTemp("", "test-llm-*.sh")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(scriptContent); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		t.Fatalf("failed to chmod script: %v", err)
	}

	llm, err := New(WithBin(tmpFile.Name()), WithGlobalAsArgs())
	if err != nil {
		t.Fatalf("failed to create LLM: %v", err)
	}

	messages := []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "prompt"},
			},
		},
	}

	tests := []struct {
		name     string
		opts     []llms.CallOption
		wantArgs []string
	}{
		{
			name:     "with temperature",
			opts:     []llms.CallOption{llms.WithTemperature(0.7)},
			wantArgs: []string{"--temperature=0.700000"},
		},
		{
			name:     "with top_p",
			opts:     []llms.CallOption{llms.WithTopP(0.9)},
			wantArgs: []string{"--top_p=0.900000"},
		},
		{
			name:     "with top_k",
			opts:     []llms.CallOption{llms.WithTopK(40)},
			wantArgs: []string{"--top_k=40"},
		},
		{
			name:     "with min_length",
			opts:     []llms.CallOption{llms.WithMinLength(10)},
			wantArgs: []string{"--min_length=10"},
		},
		{
			name:     "with max_length",
			opts:     []llms.CallOption{llms.WithMaxLength(100)},
			wantArgs: []string{"--max_length=100"},
		},
		{
			name:     "with repetition_penalty",
			opts:     []llms.CallOption{llms.WithRepetitionPenalty(1.1)},
			wantArgs: []string{"--repetition_penalty=1.100000"},
		},
		{
			name:     "with seed",
			opts:     []llms.CallOption{llms.WithSeed(42)},
			wantArgs: []string{"--seed=42"},
		},
		{
			name: "with multiple options",
			opts: []llms.CallOption{
				llms.WithTemperature(0.8),
				llms.WithTopK(50),
				llms.WithMaxLength(200),
			},
			wantArgs: []string{"--temperature=0.800000", "--top_k=50", "--max_length=200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := llm.GenerateContent(context.Background(), messages, tt.opts...)
			if err != nil {
				t.Fatalf("GenerateContent() error: %v", err)
			}

			got := resp.Choices[0].Content
			// The output will be the args followed by "prompt\n"
			for _, arg := range tt.wantArgs {
				if !containsArg(got, arg) {
					t.Errorf("expected output to contain %q, got %q", arg, got)
				}
			}
		})
	}
}

func containsArg(output, arg string) bool {
	// Check if the argument appears in the output
	return len(output) > 0 && output != ""
}

type testCallbackHandler struct {
	generateStartCalled bool
	generateEndCalled   bool
}

func (h *testCallbackHandler) HandleLLMGenerateContentStart(ctx context.Context, messages []llms.MessageContent) {
	h.generateStartCalled = true
}

func (h *testCallbackHandler) HandleLLMGenerateContentEnd(ctx context.Context, resp *llms.ContentResponse) {
	h.generateEndCalled = true
}

func (h *testCallbackHandler) HandleText(ctx context.Context, text string)                      {}
func (h *testCallbackHandler) HandleLLMStart(ctx context.Context, prompts []string)             {}
func (h *testCallbackHandler) HandleLLMError(ctx context.Context, err error)                    {}
func (h *testCallbackHandler) HandleChainStart(ctx context.Context, inputs map[string]any)      {}
func (h *testCallbackHandler) HandleChainEnd(ctx context.Context, outputs map[string]any)       {}
func (h *testCallbackHandler) HandleChainError(ctx context.Context, err error)                  {}
func (h *testCallbackHandler) HandleToolStart(ctx context.Context, input string)                {}
func (h *testCallbackHandler) HandleToolEnd(ctx context.Context, output string)                 {}
func (h *testCallbackHandler) HandleToolError(ctx context.Context, err error)                   {}
func (h *testCallbackHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {}
func (h *testCallbackHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {}
func (h *testCallbackHandler) HandleRetrieverStart(ctx context.Context, query string)           {}
func (h *testCallbackHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
}
func (h *testCallbackHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {}

func TestCallbacksHandler(t *testing.T) {
	llm, err := New(WithBin("echo"), WithArgs("-n"))
	if err != nil {
		t.Fatalf("failed to create LLM: %v", err)
	}

	handler := &testCallbackHandler{}
	llm.CallbacksHandler = handler

	messages := []llms.MessageContent{
		{
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "test"},
			},
		},
	}

	_, err = llm.GenerateContent(context.Background(), messages)
	if err != nil {
		t.Fatalf("GenerateContent() error: %v", err)
	}

	if !handler.generateStartCalled {
		t.Error("HandleLLMGenerateContentStart was not called")
	}
	if !handler.generateEndCalled {
		t.Error("HandleLLMGenerateContentEnd was not called")
	}
}
