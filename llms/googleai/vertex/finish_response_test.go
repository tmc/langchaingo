package vertex

import (
	"encoding/json"
	"errors"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestFinishVertexResponseRefusesATruncatedAnswer(t *testing.T) {
	t.Parallel()

	resp, err := convertCandidates([]*genai.Candidate{{
		Content:      &genai.Content{Parts: []genai.Part{genai.Text("half an ans")}},
		FinishReason: genai.FinishReasonMaxTokens,
	}}, nil)
	if err != nil {
		t.Fatalf("convertCandidates: %v", err)
	}

	if err := finishVertexResponse(&llms.CallOptions{}, resp); err != nil {
		t.Errorf("without the option a truncated answer must pass through, got %v", err)
	}

	err = finishVertexResponse(&llms.CallOptions{FailOnTruncation: true}, resp)
	if err == nil {
		t.Fatal("WithFailOnTruncation must refuse a truncated answer")
	}
	var typed *llms.Error
	if !errors.As(err, &typed) || typed.Code != llms.ErrCodeTruncated {
		t.Errorf("want a typed truncation error, got %v", err)
	}
	if resp.Choices[0].Content != "half an ans" {
		t.Error("the partial answer must survive the refusal")
	}
}

func TestToolCallsBelongToTheCandidateThatMadeThem(t *testing.T) {
	t.Parallel()

	call := func(name string) genai.Part {
		return genai.FunctionCall{Name: name, Args: map[string]any{"a": 1}}
	}
	resp, err := convertCandidates([]*genai.Candidate{
		{Content: &genai.Content{Parts: []genai.Part{call("alpha")}}, FinishReason: genai.FinishReasonStop},
		{Content: &genai.Content{Parts: []genai.Part{call("beta")}}, FinishReason: genai.FinishReasonStop},
	}, nil)
	if err != nil {
		t.Fatalf("convertCandidates: %v", err)
	}

	for i, want := range []string{"alpha", "beta"} {
		got := resp.Choices[i].ToolCalls
		if len(got) != 1 || got[0].FunctionCall.Name != want {
			names := make([]string, 0, len(got))
			for _, c := range got {
				names = append(names, c.FunctionCall.Name)
			}
			b, _ := json.Marshal(names)
			t.Errorf("choice %d carries %s, want exactly [%q]", i, b, want)
		}
	}
}
