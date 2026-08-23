package bedrock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/bedrock"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
)

func turnLimitMessages(last llms.ChatMessageType) []llms.MessageContent {
	msgs := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}
	if last == llms.ChatMessageTypeAI {
		msgs = append(msgs, llms.TextParts(llms.ChatMessageTypeAI, "half an "))
	}
	return msgs
}

func TestBedrockRefusesTheSameTurnsAsThePrimaryDoor(t *testing.T) {
	t.Parallel()

	for _, converse := range []bool{false, true} {
		name := "invoke-model"
		opts := []bedrock.Option{}
		if converse {
			name = "converse"
			opts = append(opts, bedrock.WithConverseAPI())
		}

		t.Run(name+"/assistant prefill is refused before the request", func(t *testing.T) {
			t.Parallel()
			llm := truncationLLMWithBody(t, `{}`,
				append([]bedrock.Option{bedrock.WithModel("us.anthropic.claude-opus-4-6-v1:0")}, opts...)...)
			_, err := llm.GenerateContent(context.Background(), turnLimitMessages(llms.ChatMessageTypeAI))
			var target *reasoning.ErrAssistantPrefillUnsupported
			if !errors.As(err, &target) {
				t.Errorf("want ErrAssistantPrefillUnsupported, got %v", err)
			}
		})

		t.Run(name+"/a forced tool with manual thinking is refused", func(t *testing.T) {
			t.Parallel()
			llm := truncationLLMWithBody(t, `{}`,
				append([]bedrock.Option{bedrock.WithModel("us.anthropic.claude-sonnet-4-5-v1:0")}, opts...)...)
			_, err := llm.GenerateContent(context.Background(), turnLimitMessages(llms.ChatMessageTypeHuman),
				llms.WithReasoning(llms.ReasoningMedium, 2048),
				llms.WithToolChoice(llms.ToolChoice{Type: "any"}))
			var target *reasoning.ErrForcedToolUseWithThinking
			if !errors.As(err, &target) {
				t.Errorf("want ErrForcedToolUseWithThinking, got %v", err)
			}
		})
	}
}
