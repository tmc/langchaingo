package llmtest_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/testing/llmtest"
)

func TestTestModelFake(t *testing.T) {
	model := fake.NewFakeLLM([]string{"OK", "Grace", "1 2 3"})
	if err := llmtest.TestModel(context.Background(), model, "multiturn", "streaming"); err != nil {
		t.Fatal(err)
	}
}

func TestTestModelNil(t *testing.T) {
	if err := llmtest.TestModel(context.Background(), nil); err == nil {
		t.Fatal("TestModel(nil) succeeded")
	}
}

func TestTestModelUnknownCapability(t *testing.T) {
	model := fake.NewFakeLLM([]string{"OK"})
	err := llmtest.TestModel(context.Background(), model, "telepathy")
	if err == nil || !strings.Contains(err.Error(), "telepathy") {
		t.Fatalf("unrecognized capability not reported: %v", err)
	}
}

// scriptModel is a configurable test double for exercising TestModel's
// capability checks and failure detection.
type scriptModel struct {
	generate func(ctx context.Context, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error)
}

func (m *scriptModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}
	return m.generate(ctx, messages, opts)
}

func (m *scriptModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	resp, err := m.GenerateContent(ctx, []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, prompt)}, options...)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}

// conformingModel behaves correctly for every capability TestModel knows.
func conformingModel() *scriptModel {
	return &scriptModel{generate: func(ctx context.Context, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
		if opts.StreamingFunc != nil {
			for _, chunk := range []string{"O", "K"} {
				if err := opts.StreamingFunc(ctx, []byte(chunk)); err != nil {
					return nil, err
				}
			}
		}
		usage := map[string]any{"TotalTokens": 3}
		if len(opts.Tools) > 0 {
			if last := messages[len(messages)-1]; last.Role == llms.ChatMessageTypeTool {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
					Content:        "It is cloudy in Paris.",
					GenerationInfo: usage,
				}}}, nil
			}
			call := llms.ToolCall{
				ID:   "call_1",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      opts.Tools[0].Function.Name,
					Arguments: `{"location": "Paris"}`,
				},
			}
			return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
				ToolCalls:      []llms.ToolCall{call},
				Parts:          []llms.ContentPart{call},
				GenerationInfo: usage,
			}}}, nil
		}
		return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
			Content:        "OK",
			GenerationInfo: usage,
		}}}, nil
	}}
}

func TestTestModelAllCapabilities(t *testing.T) {
	err := llmtest.TestModel(context.Background(), conformingModel(),
		"multiturn", "streaming", "tools", "usage")
	if err != nil {
		t.Fatal(err)
	}
}

// misbehaviorCase pairs a model that violates the TestModel contract
// with a substring of the report that violation should produce.
type misbehaviorCase struct {
	name     string
	model    llms.Model
	expected []string
	want     string // substring of the reported error
}

// misbehaviorCases returns one case per class of contract violation.
func misbehaviorCases() []misbehaviorCase {
	return []misbehaviorCase{
		{
			name: "nil response",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return nil, nil
			}},
			want: "nil response",
		},
		{
			name: "no choices",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{}, nil
			}},
			want: "no choices",
		},
		{
			name: "empty choice",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{}}}, nil
			}},
			want: "no choice carries content",
		},
		{
			name:  "ignores cancellation",
			model: ignoresCancel{},
			want:  "canceled context",
		},
		{
			name: "silent streaming",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "OK"}}}, nil
			}},
			expected: []string{"streaming"},
			want:     "no chunks",
		},
		{
			name: "swallows callback error",
			model: &scriptModel{generate: func(ctx context.Context, _ []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
				if opts.StreamingFunc != nil {
					_ = opts.StreamingFunc(ctx, []byte("OK")) // error ignored
				}
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "OK"}}}, nil
			}},
			expected: []string{"streaming"},
			want:     "callback returned one",
		},
		{
			name: "no tool call",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "the weather is nice"}}}, nil
			}},
			expected: []string{"tools"},
			want:     "no tool call",
		},
		{
			name: "invalid tool arguments",
			model: &scriptModel{generate: func(_ context.Context, _ []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
				call := llms.ToolCall{
					ID:           "call_1",
					FunctionCall: &llms.FunctionCall{Name: "get_weather", Arguments: `{"location": paris`},
				}
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{
					ToolCalls: []llms.ToolCall{call},
					Parts:     []llms.ContentPart{call},
				}}}, nil
			}},
			expected: []string{"tools"},
			want:     "not valid JSON",
		},
		{
			name: "no usage",
			model: &scriptModel{generate: func(context.Context, []llms.MessageContent, *llms.CallOptions) (*llms.ContentResponse, error) {
				return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "OK"}}}, nil
			}},
			expected: []string{"usage"},
			want:     "no positive token count",
		},
	}
}

// TestTestModelCatchesMisbehavior verifies that TestModel reports each
// class of contract violation.
func TestTestModelCatchesMisbehavior(t *testing.T) {
	for _, tt := range misbehaviorCases() {
		t.Run(tt.name, func(t *testing.T) {
			err := llmtest.TestModel(context.Background(), tt.model, tt.expected...)
			if err == nil {
				t.Fatalf("TestModel did not report misbehavior %q", tt.name)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error does not mention %q:\n%v", tt.want, err)
			}
		})
	}
}

// ignoresCancel returns a response regardless of context state.
type ignoresCancel struct{}

func (ignoresCancel) GenerateContent(context.Context, []llms.MessageContent, ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: "OK"}}}, nil
}

func (ignoresCancel) Call(context.Context, string, ...llms.CallOption) (string, error) {
	return "OK", nil
}
