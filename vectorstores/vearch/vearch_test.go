package vearch_test

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcvearch "github.com/testcontainers/testcontainers-go/modules/vearch"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/vearch"
)

func getValues(t *testing.T) string {
	t.Helper()
	vearchURL := os.Getenv("VEARCH_URL")
	if vearchURL == "" {
		vearchContainer, err := tcvearch.RunContainer(context.Background(), testcontainers.WithImage("vearch/vearch:latest"))
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, vearchContainer.Terminate(context.Background()))
		})

		vearchURL, err = vearchContainer.RESTEndpoint(context.Background())
		if err != nil {
			t.Skipf("Failed to get vearch container endpoint: %s", err)
		}
	}

	return vearchURL
}

func TestVearchStoreRest(t *testing.T) {
	t.Parallel()
	vearchURL := getValues(t)
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		t.Skip("Must set OPENAI_API_KEY to run test")
	}
	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	parsedURL, err := url.Parse(vearchURL)
	require.NoError(t, err)
	store, err := vearch.New(
		vearch.WithDBName(uuid.New().String()),
		vearch.WithSpaceName(uuid.New().String()),
		vearch.WithURL(*parsedURL),
		vearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}
