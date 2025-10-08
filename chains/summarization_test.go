package chains

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

func loadTestData(t *testing.T) []schema.Document {
	t.Helper()

	file, err := os.Open("./testdata/mouse_story.txt")
	require.NoError(t, err)

	docs, err := documentloaders.NewText(file).LoadAndSplit(
		context.Background(),
		textsplitter.NewRecursiveCharacter(),
	)
	require.NoError(t, err)

	return docs
}

func TestStuffSummarization(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)

	chain := LoadStuffSummarization(llm)
	_, err = Call(
		context.Background(),
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestRefineSummarization(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)

	chain := LoadRefineSummarization(llm)
	_, err = Call(
		context.Background(),
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestMapReduceSummarization(t *testing.T) {
	t.Parallel()

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	llm, err := openai.New()
	require.NoError(t, err)

	docs := loadTestData(t)

	chain := LoadMapReduceSummarization(llm)
	_, err = Run(
		context.Background(),
		chain,
		docs,
	)
	require.NoError(t, err)
}
