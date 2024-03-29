package openai

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func TestLLM_GenerateContent(t *testing.T) {
	t.Run("Test Generate Content Nvidia", func(t *testing.T) {
		opts := []Option{
			WithModel("meta/llama2-70b"),
			WithAPIType(APITypeNvidia),
			WithToken("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"),
		}
		llm, err := New(opts...)
		if err != nil {
			t.Error(err)
		}

		parts := []llms.ContentPart{
			llms.TextPart("I'm a pomeranian"),
			llms.TextPart("Tell me more about my taxonomy"),
		}
		content := []llms.MessageContent{
			{
				Role:  schema.ChatMessageTypeHuman,
				Parts: parts,
			},
		}

		_, err = llm.GenerateContent(context.Background(), content)
		if err != nil {
			t.Error(err)
		}
	})
}

// Test Embeddings Nvidia
func TestLLM_CreateEmbedding(t *testing.T) {
	t.Run("Test Create Embedding Nvidia", func(t *testing.T) {
		opts := []Option{
			WithModel("meta/llama2-70b"),
			WithAPIType(APITypeNvidia),
			WithToken("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"),
		}

		llm, err := New(opts...)
		if err != nil {
			t.Error(err)
		}

		_, err = llm.CreateEmbedding(context.Background(), []string{"I'm a pomeranian", "Tell me more about my taxonomy"})
		if err != nil {
			t.Error(err)
		}
	})
}
