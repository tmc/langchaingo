package opensearch_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"

	opensearchgo "github.com/opensearch-project/opensearch-go"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"

	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func getEnvVariables(t *testing.T) (string, string, string) {
	t.Helper()

	opensearchEndpoint := os.Getenv("OPENSEARCH_ENDPOINT")
	if opensearchEndpoint == "" {
		t.Fatalf("Must set %s to run test", "OPENSEARCH_ENDPOINT")
	}

	opensearchUser := os.Getenv("OPENSEARCH_USER")
	if opensearchUser == "" {
		t.Fatalf("Must set %s to run test", "OPENSEARCH_USER")
	}

	opensearchPassword := os.Getenv("OPENSEARCH_PASSWORD")
	if opensearchPassword == "" {
		t.Fatalf("Must set %s to run test", "OPENSEARCH_PASSWORD")
	}
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Fatal("OPENAI_API_KEY not set")
	}

	return opensearchEndpoint, opensearchUser, opensearchPassword
}

func setIndex(t *testing.T, storer opensearch.Store, indexName string) {
	t.Helper()
	if _, err := storer.CreateIndex(context.TODO(), indexName); err != nil {
		t.Fatalf("error creating index: %v\n", err)
	}
}

func removeIndex(t *testing.T, storer opensearch.Store, indexName string) {
	t.Helper()
	if _, err := storer.DeleteIndex(context.TODO(), indexName); err != nil {
		t.Fatalf("error deleting index: %v\n", err)
	}
}

func setLLM(t *testing.T) *openai.LLM {
	t.Helper()
	openaiOpts := []openai.Option{}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		openaiOpts = append(openaiOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
			openai.WithEmbeddingModel("text-embedding-ada-002"),
			openai.WithModel("gpt-4"),
		)
	}

	llm, err := openai.New(openaiOpts...)
	if err != nil {
		t.Fatalf("error setting openAI embedded: %v\n", err)
	}

	return llm
}

func setOpensearchClient(t *testing.T, opensearchEndpoint, opensearchUser, opensearchPassword string) *opensearchgo.Client {
	client, err := opensearchgo.NewClient(opensearchgo.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{opensearchEndpoint},
		Username:  opensearchUser,
		Password:  opensearchPassword,
	})
	if err != nil {
		t.Fatalf("cannot initialize opensearch client: %v\n", err)
	}
	return client
}

func TestOpensearchStoreRest(t *testing.T) {
	t.Parallel()
	opensearchEndpoint, opensearchUser, opensearchPassword := getEnvVariables(t)
	indexName := uuid.New().String()
	llm := setLLM(t)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := opensearch.New(
		setOpensearchClient(t, opensearchEndpoint, opensearchUser, opensearchPassword),
		opensearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	err = storer.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)

	docs, err := storer.SimilaritySearch(context.Background(), "japan", 1, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}
