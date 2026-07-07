package llmtest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// TestModel tests a Model implementation.
//
// It calls model.GenerateContent with a series of small requests and
// checks that the responses satisfy the [llms.Model] contract: a
// non-nil response with at least one non-nil choice carrying text
// content, tool calls, or typed parts, and honored context
// cancellation.
//
// The capabilities list names behaviors the model must demonstrate.
// TestModel fails if a named capability misbehaves, and does not
// exercise capabilities that are not named. Recognized capabilities:
//
//   - "multiturn": a human/assistant/human conversation generates.
//   - "streaming": [llms.WithStreamingFunc] receives at least one
//     chunk, and an error returned from the callback aborts the
//     request.
//   - "tools": a declared tool is invoked with valid JSON arguments,
//     and the tool loop round-trips: the response's assistant message
//     replays with a tool result and generation succeeds.
//   - "usage": token usage is reported in GenerationInfo under a
//     recognized token-count key.
//
// Unrecognized capability names are reported as errors.
//
// If TestModel finds any misbehaviors, it returns an error reporting
// all of them; the message is a multi-line report.
//
// Against a live provider TestModel performs network calls; record
// them (for example with httprr) to run it offline.
func TestModel(ctx context.Context, model llms.Model, capabilities ...string) error {
	t := &modelTester{ctx: ctx, model: model}
	if model == nil {
		t.errorf("model is nil")
		return errors.Join(t.errs...)
	}

	caps := make(map[string]bool, len(capabilities))
	for _, name := range capabilities {
		switch name {
		case "multiturn", "streaming", "tools", "usage":
			caps[name] = true
		default:
			t.errorf("unrecognized capability %q", name)
		}
	}

	t.testGenerate()
	t.testCancellation()
	if caps["multiturn"] {
		t.testMultiturn()
	}
	if caps["streaming"] {
		t.testStreaming()
	}
	if caps["tools"] {
		t.testTools()
	}
	if caps["usage"] {
		t.testUsage()
	}

	return errors.Join(t.errs...)
}

// modelTester accumulates contract violations across checks.
type modelTester struct {
	ctx   context.Context
	model llms.Model
	errs  []error
}

func (t *modelTester) errorf(format string, args ...any) {
	t.errs = append(t.errs, fmt.Errorf(format, args...))
}

func humanMessage(text string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextPart(text)},
	}
}

// checkResponse verifies the invariants every successful response must
// satisfy and reports whether the response is usable for further checks.
func (t *modelTester) checkResponse(op string, resp *llms.ContentResponse) bool {
	if resp == nil {
		t.errorf("%s: returned nil response with nil error", op)
		return false
	}
	if len(resp.Choices) == 0 {
		t.errorf("%s: response has no choices", op)
		return false
	}
	hasContent := false
	for i, choice := range resp.Choices {
		if choice == nil {
			t.errorf("%s: choice %d is nil", op, i)
			return false
		}
		if choice.Content != "" || len(choice.ToolCalls) > 0 || len(choice.Parts) > 0 {
			hasContent = true
		}
	}
	if !hasContent {
		t.errorf("%s: no choice carries content, tool calls, or parts", op)
		return false
	}
	return true
}

func (t *modelTester) testGenerate() {
	resp, err := t.model.GenerateContent(t.ctx,
		[]llms.MessageContent{humanMessage("Reply with the word OK.")})
	if err != nil {
		t.errorf("GenerateContent: %v", err)
		return
	}
	t.checkResponse("GenerateContent", resp)
}

func (t *modelTester) testCancellation() {
	ctx, cancel := context.WithCancel(t.ctx)
	cancel()
	resp, err := t.model.GenerateContent(ctx,
		[]llms.MessageContent{humanMessage("Reply with the word OK.")})
	if err == nil {
		t.errorf("GenerateContent: no error with canceled context (response %v)", resp)
	}
}

func (t *modelTester) testMultiturn() {
	messages := []llms.MessageContent{
		humanMessage("My name is Grace."),
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Hello Grace.")},
		},
		humanMessage("Reply with my name and nothing else."),
	}
	resp, err := t.model.GenerateContent(t.ctx, messages)
	if err != nil {
		t.errorf("multiturn: GenerateContent: %v", err)
		return
	}
	t.checkResponse("multiturn", resp)
}

func (t *modelTester) testStreaming() {
	var chunks int
	resp, err := t.model.GenerateContent(t.ctx,
		[]llms.MessageContent{humanMessage("Count from 1 to 3.")},
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			chunks++
			return nil
		}))
	if err != nil {
		t.errorf("streaming: GenerateContent: %v", err)
		return
	}
	if !t.checkResponse("streaming", resp) {
		return
	}
	if chunks == 0 {
		t.errorf("streaming: callback received no chunks")
	}

	// An error returned from the callback must abort the request.
	errStop := errors.New("llmtest: stop stream")
	_, err = t.model.GenerateContent(t.ctx,
		[]llms.MessageContent{humanMessage("Count from 1 to 3.")},
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			return errStop
		}))
	if err == nil {
		t.errorf("streaming: no error after callback returned one")
	}
}

func (t *modelTester) testTools() {
	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "get_weather",
			Description: "Get the current weather for a location.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "City name.",
					},
				},
				"required": []string{"location"},
			},
		},
	}}

	messages := []llms.MessageContent{
		humanMessage("What is the weather in Paris? Use the get_weather tool."),
	}
	resp, err := t.model.GenerateContent(t.ctx, messages, llms.WithTools(tools))
	if err != nil {
		t.errorf("tools: GenerateContent: %v", err)
		return
	}
	if !t.checkResponse("tools", resp) {
		return
	}

	assistant := resp.AssistantMessage()
	var calls []llms.ToolCall
	for _, part := range assistant.Parts {
		if call, ok := part.(llms.ToolCall); ok {
			calls = append(calls, call)
		}
	}
	if len(calls) == 0 {
		t.errorf("tools: no tool call in response")
		return
	}
	for _, call := range calls {
		if call.FunctionCall == nil {
			t.errorf("tools: tool call %q has nil FunctionCall", call.ID)
			continue
		}
		if call.FunctionCall.Name != "get_weather" {
			t.errorf("tools: called %q, want %q", call.FunctionCall.Name, "get_weather")
		}
		if !json.Valid([]byte(call.FunctionCall.Arguments)) {
			t.errorf("tools: arguments are not valid JSON: %q", call.FunctionCall.Arguments)
		}
		if !strings.Contains(strings.ToLower(call.FunctionCall.Arguments), "paris") {
			t.errorf("tools: arguments do not mention paris: %q", call.FunctionCall.Arguments)
		}
	}

	// Round-trip: replay the assistant message with a tool result. This
	// exercises multi-part fidelity, including thinking blocks on models
	// that emit them.
	loop := append(messages, assistant)
	for _, call := range calls {
		loop = append(loop, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{llms.ToolCallResponse{
				ToolCallID: call.ID,
				Name:       call.FunctionCall.Name,
				Content:    `{"temperature": 18, "conditions": "cloudy"}`,
			}},
		})
	}
	resp, err = t.model.GenerateContent(t.ctx, loop, llms.WithTools(tools))
	if err != nil {
		t.errorf("tools: round-trip GenerateContent: %v", err)
		return
	}
	t.checkResponse("tools round-trip", resp)
}

// testUsage checks GenerationInfo token keys. When the typed
// llms.Usage field lands (model-clients.md work item 8), it becomes
// the primary check and these keys the fallback.
func (t *modelTester) testUsage() {
	resp, err := t.model.GenerateContent(t.ctx,
		[]llms.MessageContent{humanMessage("Reply with the word OK.")})
	if err != nil {
		t.errorf("usage: GenerateContent: %v", err)
		return
	}
	if !t.checkResponse("usage", resp) {
		return
	}
	// Providers report usage under provider-family-specific keys:
	// Anthropic uses Input/OutputTokens, OpenAI-style providers use
	// Prompt/Completion/TotalTokens.
	keys := []string{
		"InputTokens",
		"OutputTokens",
		"PromptTokens",
		"CompletionTokens",
		"TotalTokens",
	}
	for _, choice := range resp.Choices {
		for _, key := range keys {
			if n, ok := choice.GenerationInfo[key].(int); ok && n > 0 {
				return
			}
		}
	}
	t.errorf("usage: no positive token count in GenerationInfo")
}
