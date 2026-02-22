package pgvector_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vxcontrol/langchaingo/chains"
	"github.com/vxcontrol/langchaingo/embeddings"
	"github.com/vxcontrol/langchaingo/httputil"
	"github.com/vxcontrol/langchaingo/internal/httprr"
	"github.com/vxcontrol/langchaingo/internal/testutil/testctr"
	"github.com/vxcontrol/langchaingo/llms/googleai"
	"github.com/vxcontrol/langchaingo/llms/openai"
	"github.com/vxcontrol/langchaingo/schema"
	"github.com/vxcontrol/langchaingo/vectorstores"
	"github.com/vxcontrol/langchaingo/vectorstores/pgvector"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/log"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func createOpenAIEmbedder(t *testing.T, rr *httprr.RecordReplay) *embeddings.EmbedderImpl {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	opts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	return e
}

func createOpenAILLMAndEmbedder(t *testing.T, rr *httprr.RecordReplay) (llm *openai.LLM, e *embeddings.EmbedderImpl) {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	embeddingOpts := []openai.Option{
		openai.WithEmbeddingModel("text-embedding-ada-002"),
		openai.WithHTTPClient(rr.Client()),
	}
	if !rr.Recording() {
		embeddingOpts = append(embeddingOpts, openai.WithToken("test-api-key"))
	}

	embeddingLLM, err := openai.New(embeddingOpts...)
	require.NoError(t, err)

	e, err = embeddings.NewEmbedder(embeddingLLM)
	require.NoError(t, err)

	return llm, e
}

func preCheckEnvSetting(t *testing.T) string {
	t.Helper()
	testctr.SkipIfDockerNotAvailable(t)

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	pgvectorURL := os.Getenv("PGVECTOR_CONNECTION_STRING")
	if pgvectorURL == "" {
		ctx := t.Context()
		pgVectorContainer, err := tcpostgres.Run(
			ctx,
			"docker.io/pgvector/pgvector:pg16",
			tcpostgres.WithDatabase("db_test"),
			tcpostgres.WithUsername("user"),
			tcpostgres.WithPassword("passw0rd!"),
			testcontainers.WithLogger(log.TestLogger(t)),
			testcontainers.WithWaitStrategy(
				wait.ForAll(
					wait.ForLog("database system is ready to accept connections").
						WithOccurrence(2).
						WithStartupTimeout(60*time.Second),
					wait.ForListeningPort("5432/tcp").
						WithStartupTimeout(60*time.Second),
				)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)

		t.Cleanup(func() {
			ctx := context.Background() //nolint:usetesting
			if err := pgVectorContainer.Terminate(ctx); err != nil {
				t.Logf("Failed to terminate pgvector container: %v", err)
			}
		})

		str, err := pgVectorContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err)

		pgvectorURL = str

		// Give the container a moment to fully initialize
		time.Sleep(2 * time.Second)
	}

	return pgvectorURL
}

func makeNewCollectionName() string {
	return fmt.Sprintf("test-collection-%s", uuid.New().String())
}

func cleanupTestArtifacts(ctx context.Context, t *testing.T, s pgvector.Store, pgvectorURL string) {
	t.Helper()

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	tx, err := conn.Begin(ctx)
	require.NoError(t, err)

	require.NoError(t, s.RemoveCollection(ctx, tx))

	require.NoError(t, tx.Commit(ctx))
}

func TestPgvectorStoreRest(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}

func TestPgvectorStoreRestWithScoreThreshold(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0.8),
	)
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestPgvectorStoreSimilarityScore(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo is the capital city of Japan."},
		{PageContent: "Paris is the city of love."},
		{PageContent: "I like to visit London."},
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(
		ctx,
		"What is the capital city of Japan?",
		3,
		vectorstores.WithScoreThreshold(0.8),
	)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.True(t, docs[0].Score > 0.9)
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(-0.8),
	)
	require.Error(t, err)

	_, err = store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		10,
		vectorstores.WithScoreThreshold(1.8),
	)
	require.Error(t, err)
}

// note, we can also use same llm to show this test, but need imply
// openai embedding [dimensions](https://platform.openai.com/docs/api-reference/embeddings/create#embeddings-create-dimensions) args.
func TestSimilaritySearchWithDifferentDimensions(t *testing.T) { //nolint:funlen
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "GOOGLE_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	// Avoid issue with different view of request bodies for Google AI SDK
	rr.ScrubReq(httprr.JsonCompactScrubBody)

	if !rr.Recording() {
		t.Parallel()
	}

	ctx := t.Context()
	pgvectorURL := preCheckEnvSetting(t)
	collectionName := makeNewCollectionName()

	// Configure client with httprr - use test credentials when replaying
	var opts []googleai.Option
	opts = append(opts, googleai.WithRest(), googleai.WithHTTPClient(rr.Client()))

	if rr.Replaying() {
		// Use test credentials during replay
		opts = append(opts, googleai.WithAPIKey("test-api-key"))
	} else {
		// Use real API key during recording to ensure consistent request format
		if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
			opts = append(opts, googleai.WithAPIKey(key))
		}
	}

	// use Google embedding (now default model is gemini-embedding-001, with dimensions:3072) to add some data to collection
	googleLLM, err := googleai.New(ctx, opts...)
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(googleLLM)
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Beijing"},
	})
	require.NoError(t, err)

	// use openai embedding (now default model is text-embedding-ada-002, with dimensions:1536) to add some data to same collection (same table)
	e = createOpenAIEmbedder(t, rr)

	store, err = pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(false),
		pgvector.WithCollectionName(collectionName),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London"},
		{PageContent: "New York"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(
		ctx,
		"Which of these are cities in Japan",
		5,
	)
	require.NoError(t, err)
	require.Len(t, docs, 5)
}

func TestPgvectorAsRetriever(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	llm, e := createOpenAILLMAndEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result), "orange", "expected orange in result")
}

func TestPgvectorAsRetrieverWithScoreThreshold(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	llm, e := createOpenAILLMAndEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors are all of the pieces of furniture next to the desk and the desk itself?",
	)
	require.NoError(t, err)

	require.Contains(t, strings.ToLower(result), "orange", "expected orange in result")
	require.Contains(t, strings.ToLower(result), "black", "expected black in result")
	require.Contains(t, strings.ToLower(result), "beige", "expected beige in result")
}

func TestPgvectorAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	llm, e := createOpenAILLMAndEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "in kitchen, The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "in bedroom, The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "in office, The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "in sitting room, The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "in patio, The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	result = strings.ToLower(result)

	require.Contains(t, strings.ToLower(result), "black", "expected black in result")
	require.Contains(t, strings.ToLower(result), "blue", "expected blue in result")
	require.Contains(t, strings.ToLower(result), "orange", "expected orange in result")
	require.Contains(t, strings.ToLower(result), "purple", "expected purple in result")
	require.Contains(t, strings.ToLower(result), "yellow", "expected yellow in result")
}

func TestPgvectorAsRetrieverWithMetadataFilters(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	llm, e := createOpenAILLMAndEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "In office, the color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location":    "office",
					"square_feet": 100,
				},
			},
			{
				PageContent: "in sitting room, the color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 400,
				},
			},
			{
				PageContent: "in patio, the color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location":    "patio",
					"square_feet": 800,
				},
			},
		},
	)
	require.NoError(t, err)

	filter := map[string]any{"location": "sitting room"}

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store,
				5,
				vectorstores.WithFilters(filter))),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result), "purple", "expected purple in result")
	require.NotContains(t, strings.ToLower(result), "orange", "expected not orange in result")
	require.NotContains(t, strings.ToLower(result), "yellow", "expected not yellow in result")
}

func TestDeduplicater(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)

	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"type": "city",
		}},
		{PageContent: "potato", Metadata: map[string]any{
			"type": "vegetable",
		}},
	}, vectorstores.WithDeduplicater(
		func(_ context.Context, doc schema.Document) bool {
			return doc.PageContent == "tokyo"
		},
	))
	require.NoError(t, err)

	docs, err := store.Search(ctx, 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "potato", docs[0].PageContent)
	require.Equal(t, "vegetable", docs[0].Metadata["type"])
}

func TestWithAllOptions(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()
	rr.ScrubResp(httprr.EmbeddingJSONFormatter())

	if !rr.Recording() {
		t.Parallel()
	}

	pgvectorURL := preCheckEnvSetting(t)
	ctx := t.Context()

	e := createOpenAIEmbedder(t, rr)
	conn, err := pgx.Connect(ctx, pgvectorURL)
	require.NoError(t, err)
	defer conn.Close(ctx)

	store, err := pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
		pgvector.WithCollectionTableName("collection_table_name"),
		pgvector.WithEmbeddingTableName("embedding_table_name"),
		pgvector.WithCollectionMetadata(map[string]any{
			"key": "value",
		}),
		pgvector.WithVectorDimensions(1536),
		pgvector.WithHNSWIndex(16, 64, "vector_l2_ops"),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])

	e = createOpenAIEmbedder(t, rr)
	store, err = pgvector.New(
		ctx,
		pgvector.WithConn(conn),
		pgvector.WithEmbedder(e),
		pgvector.WithPreDeleteCollection(true),
		pgvector.WithCollectionName(makeNewCollectionName()),
		pgvector.WithCollectionTableName("collection_table_name1"),
		pgvector.WithEmbeddingTableName("embedding_table_name1"),
		pgvector.WithCollectionMetadata(map[string]any{
			"key": "value",
		}),
		pgvector.WithVectorDimensions(1536),
		pgvector.WithHNSWIndex(16, 64, "vector_l2_ops"),
	)
	require.NoError(t, err)

	defer cleanupTestArtifacts(ctx, t, store, pgvectorURL)

	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err = store.SimilaritySearch(ctx, "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}
