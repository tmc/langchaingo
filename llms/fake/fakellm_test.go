package fake

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

func TestFakeLLM(t *testing.T) {
	responses := []string{
		"Resposta 1",
		"Resposta 2",
		"Resposta 3",
	}

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()

	t.Run("Test Call Method", func(t *testing.T) {
		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 1" {
			t.Errorf("Expected 'Resposta 1', got '%s'", output)
		}

		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 2" {
			t.Errorf("Expected 'Resposta 2', got '%s'", output)
		}

		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 3" {
			t.Errorf("Expected 'Resposta 3', got '%s'", output)
		}

		// Testa reinicialização automática
		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 1" {
			t.Errorf("Expected 'Resposta 1', got '%s'", output)
		}
	})

	t.Run("Test GenerateContent Method", func(t *testing.T) {
		msg := llms.MessageContent{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: "Teste"}},
		}

		resp, err := fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(resp.Choices) < 1 || resp.Choices[0].Content != "Resposta 2" {
			t.Errorf("Expected 'Resposta 2', got '%s'", resp.Choices[0].Content)
		}

		resp, err = fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(resp.Choices) < 1 || resp.Choices[0].Content != "Resposta 3" {
			t.Errorf("Expected 'Resposta 3', got '%s'", resp.Choices[0].Content)
		}

		resp, err = fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(resp.Choices) < 1 || resp.Choices[0].Content != "Resposta 1" {
			t.Errorf("Expected 'Resposta 1', got '%s'", resp.Choices[0].Content)
		}
	})

	t.Run("Test Reset Method", func(t *testing.T) {
		fakeLLM.Reset()
		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 1" {
			t.Errorf("Expected 'Resposta 1', got '%s'", output)
		}
	})

	t.Run("Test AddResponse Method", func(t *testing.T) {
		fakeLLM.AddResponse("Resposta 4")
		fakeLLM.Reset()
		fakeLLM.Call(ctx, "Teste") // Descartar "Resposta 1"
		fakeLLM.Call(ctx, "Teste") // Descartar "Resposta 2"
		fakeLLM.Call(ctx, "Teste") // Descartar "Resposta 3"
		if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 4" {
			t.Errorf("Expected 'Resposta 4', got '%s'", output)
		}
	})

	t.Run("Test with Chain", func(t *testing.T) {
		fakeLLM.AddResponse("My name is Alexandre")
		llmChain := chains.NewConversation(fakeLLM, memory.NewConversationBuffer())
		out, err := chains.Run(ctx, llmChain, "What's my name? How many times did I ask this?")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if out != "My name is Alexandre" {
			t.Errorf("Expected 'My name is Alexandre', got '%s'", out)
		}

	})
}
