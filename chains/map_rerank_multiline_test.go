package chains

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func TestMapRerankPreservesMultilineAnswer(t *testing.T) {
	t.Parallel()
	for _, answer := range []string{
		"First sentence.\nSecond sentence.",
		"Steps:\n\n- first\n- second",
		"First sentence.\r\nSecond sentence.",
	} {
		t.Run(answer, func(t *testing.T) {
			c := NewMapRerankDocuments(NewLLMChain(
				&testLanguageModel{},
				prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
			))
			c.MaxConcurrentWorkers = 2
			docs := []schema.Document{
				{PageContent: "Incomplete answer\nScore: 10"},
				{PageContent: answer + "\nScore: 100\n"},
			}
			result, err := Run(context.Background(), c, docs)
			require.NoError(t, err)
			require.Equal(t, answer, result)
		})
	}
}

func TestMapRerankCRLFScoreLine(t *testing.T) {
	t.Parallel()
	c := NewMapRerankDocuments(NewLLMChain(
		&testLanguageModel{},
		prompts.NewPromptTemplate("{{.context}}", []string{"context"}),
	))
	c.ReturnIntermediateSteps = true
	result, err := Call(context.Background(), c, map[string]any{
		c.InputKey: []schema.Document{
			{PageContent: "Incomplete answer\r\nScore: 10\r\n"},
			{PageContent: "First line\r\nSecond line\r\nScore: 100\r\n"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "First line\r\nSecond line", result[c.OutputKey])
	steps, ok := result[_intermediateStepsOutputKey].([]map[string]any)
	require.True(t, ok)
	require.Equal(t, "100", steps[0][c.RankKey])
}
