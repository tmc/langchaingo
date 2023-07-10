package chains

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/openai"
)

func TestRefineQA(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)
	qaChain := LoadRefineQA(llm)

	results, err := Call(
		context.Background(),
		qaChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)
	require.NoError(t, err)

	_, ok := results["text"].(string)
	require.True(t, ok, "result does not contain text key")
}

func TestMapReduceQA(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)
	qaChain := LoadMapReduceQA(llm)

	result, err := Predict(
		context.Background(),
		qaChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)

	require.NoError(t, err)
	require.True(t, strings.Contains(result, "Leo"), "result does not contain correct answer Leo")
}

func TestMapRerankQA(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)
	mapRerankChain := LoadMapRerankQA(llm)

	results, err := Call(
		context.Background(),
		mapRerankChain,
		map[string]any{
			"input_documents": docs,
			"question":        "What is the name of the lion?",
		},
	)

	answer, ok := results["answer"].(string)
	require.True(t, ok, "result does not contain answer key")
	require.True(t, strings.Contains(answer, "Leo"), "result does not contain correct answer Leo")

	score, ok := results["score"].(string)
	require.True(t, ok, "result does not contain score key")
	require.True(t, strings.Contains(score, "100"), "result does not score answer as 100")
}
