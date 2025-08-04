package agents

import (
	"context"
	"os"
	"regexp"
	"testing"
	
	"github.com/stretchr/testify/require"
	"github.com/yincongcyincong/langchaingo/chains"
	"github.com/yincongcyincong/langchaingo/llms/openai"
	"github.com/yincongcyincong/langchaingo/memory"
	"github.com/yincongcyincong/langchaingo/tools"
)

func TestConversationalWithMemory(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	
	llm, err := openai.New(openai.WithModel("gpt-4"))
	require.NoError(t, err)
	
	executor, err := Initialize(
		llm,
		[]tools.Tool{tools.Calculator{}},
		ConversationalReactDescription,
		WithMemory(memory.NewConversationBuffer()),
	)
	require.NoError(t, err)
	
	_, err = chains.Run(context.Background(), executor, "Hi! my name is Bob and the year I was born is 1987")
	require.NoError(t, err)
	
	res, err := chains.Run(context.Background(), executor, "What is the year I was born times 34")
	require.NoError(t, err)
	expectedRe := "67,?558"
	if !regexp.MustCompile(expectedRe).MatchString(res) {
		t.Errorf("result does not contain the crrect answer '67558', got: %s", res)
	}
}
