package kronk

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		name string
		opts llms.CallOptions
		want model.D
	}{
		{
			name: "defaults",
			want: model.D{},
		},
		{
			name: "call sampling options",
			opts: llms.CallOptions{
				MaxTokens:         128,
				Temperature:       0.2,
				TopK:              20,
				TopP:              0.95,
				Seed:              42,
				StopWords:         []string{"END"},
				RepetitionPenalty: 1.1,
				FrequencyPenalty:  0.3,
				PresencePenalty:   0.4,
			},
			want: model.D{
				"max_tokens":        128,
				"temperature":       0.2,
				"top_k":             20,
				"top_p":             0.95,
				"seed":              42,
				"stop":              []string{"END"},
				"repeat_penalty":    1.1,
				"frequency_penalty": 0.3,
				"presence_penalty":  0.4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyOptions(model.D{}, tt.opts)
			if err != nil {
				t.Fatalf("applyOptions: got error %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyOptions: got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestModelOptionsPreserveOrder(t *testing.T) {
	var opts options
	for _, opt := range []Option{
		WithAutoTune(true),
		WithConfig(model.Config{}),
		WithAutoTune(true),
		WithContextWindow(8192),
	} {
		opt(&opts)
	}

	cfg := model.NewConfig(opts.modelOptions...)
	if !cfg.AutoTune {
		t.Error("AutoTune: got false, want true")
	}
	if cfg.ContextWindow() != 8192 {
		t.Errorf("ContextWindow: got %d, want 8192", cfg.ContextWindow())
	}
}

func TestMessageContentToDocs(t *testing.T) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hello"),
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.ToolCall{
				ID:   "call-1",
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      "weather",
					Arguments: `{"city":"Austin"}`,
				},
			}},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: "call-1",
				Name:       "weather",
				Content:    "sunny",
			}},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("describe this"),
				llms.BinaryPart("image/png", []byte("image")),
			},
		},
	}

	docs, err := messageContentToDocs(messages)
	if err != nil {
		t.Fatalf("messageContentToDocs: got error %v, want nil", err)
	}
	if got, want := docs[0]["content"], "hello"; got != want {
		t.Errorf("text content: got %v, want %v", got, want)
	}
	toolCalls := docs[1]["tool_calls"].([]model.D)
	if got, want := toolCalls[0]["id"], "call-1"; got != want {
		t.Errorf("tool call ID: got %v, want %v", got, want)
	}
	if got, want := docs[2]["tool_call_id"], "call-1"; got != want {
		t.Errorf("tool response ID: got %v, want %v", got, want)
	}
	content := docs[3]["content"].([]model.D)
	if got, want := content[1]["type"], "image_url"; got != want {
		t.Errorf("media type: got %v, want %v", got, want)
	}
}

func TestApplyOptionsToolsAndJSON(t *testing.T) {
	opts := llms.CallOptions{
		JSONMode: true,
		Tools: []llms.Tool{{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "weather",
				Description: "Get weather",
				Parameters:  json.RawMessage(`{"type":"object"}`),
			},
		}},
		ToolChoice: "required",
	}

	d, err := applyOptions(model.D{}, opts)
	if err != nil {
		t.Fatalf("applyOptions: got error %v, want nil", err)
	}
	if got, want := d["tool_choice"], "required"; got != want {
		t.Errorf("tool choice: got %v, want %v", got, want)
	}
	tools := d["tools"].([]model.D)
	function := tools[0]["function"].(model.D)
	if got, want := function["name"], "weather"; got != want {
		t.Errorf("tool name: got %v, want %v", got, want)
	}
	responseFormat := d["response_format"].(model.D)
	if got, want := responseFormat["type"], "json_object"; got != want {
		t.Errorf("response format: got %v, want %v", got, want)
	}
}

type callbackRecorder struct {
	callbacks.SimpleHandler
	starts int
	errors int
}

func (cr *callbackRecorder) HandleLLMGenerateContentStart(_ context.Context, _ []llms.MessageContent) {
	cr.starts++
}

func (cr *callbackRecorder) HandleLLMError(_ context.Context, _ error) {
	cr.errors++
}

func TestGenerateContentCallbacksOnConversionError(t *testing.T) {
	recorder := callbackRecorder{}
	client := Client{callbacks: &recorder}

	_, err := client.GenerateContent(t.Context(), []llms.MessageContent{{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.BinaryPart("application/pdf", []byte("pdf"))},
	}})
	if err == nil {
		t.Fatal("GenerateContent: got nil error, want unsupported media error")
	}
	if recorder.starts != 1 {
		t.Errorf("start callbacks: got %d, want 1", recorder.starts)
	}
	if recorder.errors != 1 {
		t.Errorf("error callbacks: got %d, want 1", recorder.errors)
	}
}

func TestProcessContentStream(t *testing.T) {
	finishReason := model.FinishReasonLength
	ch := make(chan model.ChatResponse, 3)
	ch <- model.ChatResponse{Choices: []model.Choice{{Delta: &model.ResponseMessage{
		Content:   "answer",
		Reasoning: "thinking",
		ToolCallDeltas: []model.ResponseToolCallDelta{{
			ID:    "call-1",
			Index: 0,
			Type:  "function",
			Function: model.ResponseToolCallDeltaFunction{
				Name:      "weather",
				Arguments: `{"city":`,
			},
		}},
	}}}}
	ch <- model.ChatResponse{Choices: []model.Choice{{Delta: &model.ResponseMessage{
		ToolCallDeltas: []model.ResponseToolCallDelta{{
			Index:    0,
			Function: model.ResponseToolCallDeltaFunction{Arguments: `"Austin"}`},
		}},
	}}}}
	ch <- model.ChatResponse{
		Choices: []model.Choice{{
			Delta:           &model.ResponseMessage{},
			FinishReasonPtr: &finishReason,
		}},
		Usage: &model.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	}
	close(ch)

	var streamedContent strings.Builder
	var streamedReasoning strings.Builder
	client := Client{}
	response, err := client.processContentStream(t.Context(), ch, llms.CallOptions{
		StreamingFunc: func(_ context.Context, chunk []byte) error {
			streamedContent.Write(chunk)
			return nil
		},
		StreamingReasoningFunc: func(_ context.Context, reasoning, _ []byte) error {
			streamedReasoning.Write(reasoning)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("processContentStream: got error %v, want nil", err)
	}

	choice := response.Choices[0]
	if got, want := choice.StopReason, model.FinishReasonLength; got != want {
		t.Errorf("stop reason: got %q, want %q", got, want)
	}
	if got, want := streamedContent.String(), "answer"; got != want {
		t.Errorf("streamed content: got %q, want %q", got, want)
	}
	if got, want := streamedReasoning.String(), "thinking"; got != want {
		t.Errorf("streamed reasoning: got %q, want %q", got, want)
	}
	if got, want := choice.ToolCalls[0].FunctionCall.Arguments, `{"city":"Austin"}`; got != want {
		t.Errorf("tool arguments: got %q, want %q", got, want)
	}
	if got, want := choice.GenerationInfo["total_tokens"], 15; got != want {
		t.Errorf("total tokens: got %v, want %v", got, want)
	}
}

func TestProcessContentStreamCallbackError(t *testing.T) {
	ch := make(chan model.ChatResponse, 1)
	ch <- model.ChatResponse{Choices: []model.Choice{{Delta: &model.ResponseMessage{Content: "chunk"}}}}
	close(ch)

	wantErr := errors.New("stop streaming")
	client := Client{}
	response, err := client.processContentStream(t.Context(), ch, llms.CallOptions{
		StreamingFunc: func(context.Context, []byte) error { return wantErr },
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("processContentStream: got error %v, want %v", err, wantErr)
	}
	if got, want := response.Choices[0].Content, "chunk"; got != want {
		t.Errorf("partial content: got %q, want %q", got, want)
	}
}

func TestProcessContentStreamRequiresFinishReason(t *testing.T) {
	ch := make(chan model.ChatResponse)
	close(ch)

	client := Client{}
	_, err := client.processContentStream(t.Context(), ch, llms.CallOptions{})
	if err == nil || !strings.Contains(err.Error(), "without a finish reason") {
		t.Errorf("processContentStream: got error %v, want missing finish reason", err)
	}
}
