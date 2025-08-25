package opensearch_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	opensearchgo "github.com/opensearch-project/opensearch-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	tcopensearch "github.com/testcontainers/testcontainers-go/modules/opensearch"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/internal/testutil/testctr"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/opensearch"
)

func getEnvVariables(t *testing.T) (string, string, string) {
	t.Helper()
	testctr.SkipIfDockerNotAvailable(t)

	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	var osUser string
	var osPassword string

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skipf("Must set %s to run test", "OPENAI_API_KEY")
	}

	ctx := context.Background()
	opensearchEndpoint := os.Getenv("OPENSEARCH_ENDPOINT")
	if opensearchEndpoint == "" {
		openseachContainer, err := tcopensearch.Run(ctx, "opensearchproject/opensearch:2.11.1", testcontainers.WithLogger(log.TestLogger(t)))
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			if err := openseachContainer.Terminate(context.Background()); err != nil {
				t.Logf("Failed to terminate opensearch container: %v", err)
			}
		})

		address, err := openseachContainer.Address(ctx)
		if err != nil {
			t.Skipf("cannot get address of opensearch container: %v\n", err)
		}

		opensearchEndpoint = address
		osUser = openseachContainer.User
		osPassword = openseachContainer.Password
	}

	opensearchUser := os.Getenv("OPENSEARCH_USER")
	if opensearchUser == "" {
		opensearchUser = osUser
		if opensearchUser == "" {
			t.Skipf("Must set %s to run test", "OPENSEARCH_USER")
		}
	}

	opensearchPassword := os.Getenv("OPENSEARCH_PASSWORD")
	if opensearchPassword == "" {
		opensearchPassword = osPassword
		if opensearchPassword == "" {
			t.Skipf("Must set %s to run test", "OPENSEARCH_PASSWORD")
		}
	}

	return opensearchEndpoint, opensearchUser, opensearchPassword
}

func setIndex(t *testing.T, storer opensearch.Store, indexName string) {
	t.Helper()
	ctx := context.Background()
	_, err := storer.CreateIndex(ctx, indexName)
	if err != nil {
		t.Fatalf("error creating index: %v\n", err)
	}
}

func removeIndex(t *testing.T, storer opensearch.Store, indexName string) {
	t.Helper()
	ctx := context.Background()
	_, err := storer.DeleteIndex(ctx, indexName)
	if err != nil {
		t.Fatalf("error deleting index: %v\n", err)
	}
}

// createOpenAIEmbedder creates an OpenAI embedder using the provided httprr client.
func createOpenAIEmbedder(t *testing.T, httpClient *http.Client) *embeddings.EmbedderImpl {
	t.Helper()

	openaiOpts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(httpClient),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	// When httpClient is not DefaultClient, we need to check if we're recording
	// If we're replaying (not recording), use fake token
	// When recording, openai.New() will read OPENAI_API_KEY from environment
	if httpClient != http.DefaultClient {
		// This is during test - but we need to know if we're recording or replaying
		// For now, assume if no OPENAI_API_KEY is set, we're replaying
		if os.Getenv("OPENAI_API_KEY") == "" {
			openaiOpts = append(openaiOpts, openai.WithToken("test-api-key"))
		}
	}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		openaiOpts = append(openaiOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
		)
	}

	llm, err := openai.New(openaiOpts...)
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)
	return e
}

// createOpenAILLMAndEmbedder creates both LLM and embedder using the provided httprr client.
func createOpenAILLMAndEmbedder(t *testing.T, httpClient *http.Client, recording bool) (*openai.LLM, *embeddings.EmbedderImpl) {
	t.Helper()

	llmOpts := []openai.Option{
		openai.WithHTTPClient(httpClient),
	}
	// Only add fake token when NOT recording (i.e., during replay)
	if !recording {
		llmOpts = append(llmOpts, openai.WithToken("test-api-key"))
	}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		llmOpts = append(llmOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
			openai.WithModel("gpt-4"),
		)
	}

	llm, err := openai.New(llmOpts...)
	require.NoError(t, err)

	embeddingOpts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(httpClient),
	}
	// Only add fake token when NOT recording (i.e., during replay)
	if !recording {
		embeddingOpts = append(embeddingOpts, openai.WithToken("test-api-key"))
	}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		embeddingOpts = append(embeddingOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
		)
	}

	embeddingLLM, err := openai.New(embeddingOpts...)
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(embeddingLLM)
	require.NoError(t, err)
	return llm, e
}

func setOpensearchClient(
	t *testing.T,
	opensearchEndpoint,
	opensearchUser,
	opensearchPassword string,
) *opensearchgo.Client {
	t.Helper()
	client, err := opensearchgo.NewClient(opensearchgo.Config{
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
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENSEARCH_ENDPOINT", "OPENSEARCH_USER", "OPENSEARCH_PASSWORD", "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()
	opensearchEndpoint, opensearchUser, opensearchPassword := getEnvVariables(t)
	indexName := uuid.New().String()
	e := createOpenAIEmbedder(t, rr.Client())

	storer, err := opensearch.New(
		setOpensearchClient(t, opensearchEndpoint, opensearchUser, opensearchPassword),
		opensearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	time.Sleep(time.Second)
	docs, err := storer.SimilaritySearch(ctx, "japan", 1, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}

func TestOpensearchStoreRestWithScoreThreshold(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENSEARCH_ENDPOINT", "OPENSEARCH_USER", "OPENSEARCH_PASSWORD", "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()
	opensearchEndpoint, opensearchUser, opensearchPassword := getEnvVariables(t)
	indexName := uuid.New().String()

	e := createOpenAIEmbedder(t, rr.Client())

	storer, err := opensearch.New(
		setOpensearchClient(t, opensearchEndpoint, opensearchUser, opensearchPassword),
		opensearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London "},
		{PageContent: "New York"},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	time.Sleep(time.Second)
	// test with a score threshold of 0.72, expected 6 documents
	docs, err := storer.SimilaritySearch(ctx,
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.72),
		vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	require.Len(t, docs, 6)
}

func TestOpensearchAsRetriever(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENSEARCH_ENDPOINT", "OPENSEARCH_USER", "OPENSEARCH_PASSWORD", "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()
	opensearchEndpoint, opensearchUser, opensearchPassword := getEnvVariables(t)
	indexName := uuid.New().String()

	llm, e := createOpenAILLMAndEmbedder(t, rr.Client(), rr.Recording())

	storer, err := opensearch.New(
		setOpensearchClient(t, opensearchEndpoint, opensearchUser, opensearchPassword),
		opensearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(indexName),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(storer, 1, vectorstores.WithNameSpace(indexName)),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestOpensearchAsRetrieverWithScoreThreshold(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENSEARCH_ENDPOINT", "OPENSEARCH_USER", "OPENSEARCH_PASSWORD", "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()
	opensearchEndpoint, opensearchUser, opensearchPassword := getEnvVariables(t)
	indexName := uuid.New().String()

	llm, e := createOpenAILLMAndEmbedder(t, rr.Client(), rr.Recording())

	storer, err := opensearch.New(
		setOpensearchClient(t, opensearchEndpoint, opensearchUser, opensearchPassword),
		opensearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
		vectorstores.WithNameSpace(indexName),
	)
	require.NoError(t, err)
	time.Sleep(time.Second)
	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(storer, 5,
				vectorstores.WithNameSpace(indexName),
				vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}
