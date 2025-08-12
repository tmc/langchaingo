package bedrock_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/embeddings/bedrock"
)

func TestEmbedQuery(t *testing.T) {
	t.Parallel()
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}
	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelTitanEmbedG1))
	require.NoError(t, err)
	_, err = model.EmbedQuery(context.Background(), "hello world")

	require.NoError(t, err)
}

func TestEmbedDocuments(t *testing.T) {
	t.Parallel()
	if os.Getenv("TEST_AWS") != "true" {
		t.Skip("Skipping test, requires AWS access")
	}
	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelCohereEn))
	require.NoError(t, err)

	embeddings, err := model.EmbedDocuments(context.Background(), []string{"hello world", "goodbye world"})

	require.NoError(t, err)
	require.Len(t, embeddings, 2)
}
