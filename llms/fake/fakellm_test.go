package fake

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
)

var responses = []string{
	"Resposta 1",
	"Resposta 2",
	"Resposta 3",
}

func TestFakeLLM_CallMethod(t *testing.T) {

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()

	if output, _ := fakeLLM.Call(ctx, "Teste"); output != responses[0] {
		t.Errorf("Expected 'Resposta 1', got '%s'", output)
	}

	if output, _ := fakeLLM.Call(ctx, "Teste"); output != responses[1] {
		t.Errorf("Expected 'Resposta 2', got '%s'", output)
	}

	if output, _ := fakeLLM.Call(ctx, "Teste"); output != responses[2] {
		t.Errorf("Expected 'Resposta 3', got '%s'", output)
	}

	// Testa reinicialização automática
	if output, _ := fakeLLM.Call(ctx, "Teste"); output != responses[0] {
		t.Errorf("Expected 'Resposta 1', got '%s'", output)
	}
}

func TestFakeLLM_GenerateContentMethod(t *testing.T) {

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()
	msg := llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: "Teste"}},
	}

	resp, err := fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(resp.Choices) < 1 || resp.Choices[0].Content != responses[0] {
		t.Errorf("Expected 'Resposta 1', got '%s'", resp.Choices[0].Content)
	}

	resp, err = fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(resp.Choices) < 1 || resp.Choices[0].Content != responses[1] {
		t.Errorf("Expected 'Resposta 2', got '%s'", resp.Choices[0].Content)
	}

	resp, err = fakeLLM.GenerateContent(ctx, []llms.MessageContent{msg})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(resp.Choices) < 1 || resp.Choices[0].Content != responses[2] {
		t.Errorf("Expected 'Resposta 1', got '%s'", resp.Choices[0].Content)
	}
}

func TestFakeLLM_ResetMethod(t *testing.T) {

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()

	fakeLLM.Reset()
	if output, _ := fakeLLM.Call(ctx, "Teste"); output != responses[0] {
		t.Errorf("Expected 'Resposta 1', got '%s'", output)
	}
}

func TestFakeLLM_AddResponseMethod(t *testing.T) {

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()

	fakeLLM.AddResponse("Resposta 4")
	fakeLLM.Reset()
	_, err := fakeLLM.Call(ctx, "Teste") // Descarta responses[0]
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = fakeLLM.Call(ctx, "Teste") // Descarta responses[1]
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	_, err = fakeLLM.Call(ctx, "Teste") // Descarta responses[2]

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if output, _ := fakeLLM.Call(ctx, "Teste"); output != "Resposta 4" {
		t.Errorf("Expected 'Resposta 4', got '%s'", output)
	}
}

func TestFakeLLM_WithChain(t *testing.T) {

	fakeLLM := NewFakeLLM(responses)
	ctx := context.Background()

	fakeLLM.AddResponse("My name is Alexandre")

	NextToResponse(fakeLLM, 4)
	llmChain := chains.NewConversation(fakeLLM, memory.NewConversationBuffer())
	out, err := chains.Run(ctx, llmChain, "What's my name? How many times did I ask this?")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if out != "My name is Alexandre" {
		t.Errorf("Expected 'My name is Alexandre', got '%s'", out)
	}
}

// jump some responses
func NextToResponse(fakeLLM *LLM, n int) {
	for i := 1; i < n; i++ {
		_, err := fakeLLM.Call(context.Background(), "Teste")
		if err != nil {
			panic(err)
		}
	}
}
