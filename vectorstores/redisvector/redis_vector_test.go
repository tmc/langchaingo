package redisvector_test

import (
	"context"
	_ "embed"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/redisvector"
)

const (
	ollamaEmbeddingModel = "nomic-embed-text"
	ollamaChatModel      = "gemma3:1b"
)

func getValues(t *testing.T) (string, string) {
	t.Helper()

	// Default to localhost if OLLAMA_HOST not set
	ollamaURL := os.Getenv("OLLAMA_HOST")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	uri := os.Getenv("REDIS_URL")
	if uri == "" {
		ctx := context.Background()

		redisContainer, err := tcredis.Run(ctx,
			"docker.io/redis/redis-stack:7.2.0-v10",
			testcontainers.WithLogger(tclog.TestLogger(t)),
		)
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)

		t.Cleanup(func() {
			if err := redisContainer.Terminate(context.Background()); err != nil {
				t.Logf("Failed to terminate redis container: %v", err)
			}
		})

		url, err := redisContainer.ConnectionString(ctx)
		if err != nil {
			log.Fatalf("failed to start container: %s", err)
		}
		uri = url
	}

	return uri, ollamaURL
}

//go:embed testdata/schema.json
var jsonSchemaData string

//go:embed testdata/schema.yml
var yamlSchemaData string

func TestCreateRedisVectorOptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisURL, ollamaURL := getValues(t)
	_, e := getOllamaClient(ollamaEmbeddingModel, ollamaChatModel, ollamaURL)
	index := "test_case1"

	_, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithEmbedder(e),
	)
	assert.Equal(t, "invalid options: missing index name", err.Error())

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, false),
	)
	assert.Equal(t, "invalid options: missing embedder", err.Error())

	_, err = redisvector.New(ctx,
		redisvector.WithIndexName(index, false),
		redisvector.WithEmbedder(e),
	)
	assert.Equal(t, "redis: invalid URL scheme: ", err.Error())

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, false),
		redisvector.WithEmbedder(e),
	)
	assert.Equal(t, "redis index name does not exist", err.Error())

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "./testdata/not_exists.yml", nil),
	)
	assert.Equal(t, "open ./testdata/not_exists.yml: no such file or directory", err.Error())

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "", nil),
	)
	assert.Equal(t, redisvector.ErrEmptySchemaContent, err)

	// create redis vector with file
	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "./testdata/schema.yml", nil),
	)
	require.NoError(t, err)

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.JSONSchemaFormat, "./testdata/schema.json", nil),
	)
	require.NoError(t, err)

	// create redis vector with string
	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.JSONSchemaFormat, "", []byte(jsonSchemaData)),
	)
	require.NoError(t, err)

	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "", []byte(yamlSchemaData)),
	)
	require.NoError(t, err)
}

func TestAddDocuments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisURL, ollamaURL := getValues(t)
	_, e := getOllamaClient(ollamaEmbeddingModel, ollamaChatModel, ollamaURL)

	index := "test_add_document"
	prefix := "doc:"

	_, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, false),
		redisvector.WithEmbedder(e),
	)
	assert.Equal(t, "redis index name does not exist", err.Error())

	vector, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

	err = vector.DropIndex(ctx, index, false)
	assert.Equal(t, "redis index name does not exist", err.Error())

	//nolint: dupl
	data := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"population": 1.46, "area": 828}},
		{PageContent: "Hiroshima", Metadata: map[string]any{"population": 1.2, "area": 905}},
		{PageContent: "Kazuno", Metadata: map[string]any{"population": 0.04, "area": 707}},
		{PageContent: "Nagoya", Metadata: map[string]any{"population": 2.3, "area": 326}},
		{PageContent: "Toyota", Metadata: map[string]any{"population": 0.42, "area": 918}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"population": 1.59, "area": 341}},
		{PageContent: "Paris", Metadata: map[string]any{"population": 11, "area": 105}},
		{PageContent: "London", Metadata: map[string]any{"population": 9.5, "area": 1572}},
		{PageContent: "Santiago", Metadata: map[string]any{"population": 6.9, "area": 641}},
		{PageContent: "Buenos Aires", Metadata: map[string]any{"population": 15.5, "area": 203}},
		{PageContent: "Rio de Janeiro", Metadata: map[string]any{"population": 13.7, "area": 1200}},
		{PageContent: "Sao Paulo", Metadata: map[string]any{"population": 22.6, "area": 1523}},
	}
	// create redis vector with not existed index, creating index when adding docs
	docIDs, err := vector.AddDocuments(ctx, data)
	require.NoError(t, err)
	assert.Equal(t, len(data), len(docIDs))
	assert.True(t, strings.HasPrefix(docIDs[0], prefix+index))

	// create data with ids or keys
	dataWithIDOrKeys := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"ids": "id1", "population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"keys": "key1", "population": 1.46, "area": 828}},
	}

	docIDs, err = vector.AddDocuments(ctx, dataWithIDOrKeys)
	require.NoError(t, err)
	assert.Equal(t, len(dataWithIDOrKeys), len(docIDs))
	assert.Equal(t, prefix+index+":id1", docIDs[0])
	assert.Equal(t, prefix+index+":key1", docIDs[1])

	// create vector with existed index & index schema, will not create new index
	_, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "./testdata/schema.yml", nil),
	)
	require.NoError(t, err)

	// create vector with not existed index & index schema, will create new index with schema
	newIndex := index + "_new"
	vector, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(newIndex, true),
		redisvector.WithEmbedder(e),
		redisvector.WithIndexSchema(redisvector.YAMLSchemaFormat, "./testdata/schema.yml", nil),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = vector.DropIndex(ctx, index, true)
		require.NoError(t, err)
		err = vector.DropIndex(ctx, newIndex, true)
		require.NoError(t, err)
	})
}

func TestSimilaritySearch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisURL, ollamaURL := getValues(t)
	_, e := getOllamaClient(ollamaEmbeddingModel, ollamaChatModel, ollamaURL)

	index := "test_similarity_search"

	store, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

	//nolint: dupl
	data := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"population": 1.46, "area": 828}},
		{PageContent: "Hiroshima", Metadata: map[string]any{"population": 1.2, "area": 905}},
		{PageContent: "Kazuno", Metadata: map[string]any{"population": 0.04, "area": 707}},
		{PageContent: "Nagoya", Metadata: map[string]any{"population": 2.3, "area": 326}},
		{PageContent: "Toyota", Metadata: map[string]any{"population": 0.42, "area": 918}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"population": 1.59, "area": 341}},
		{PageContent: "Paris", Metadata: map[string]any{"population": 11, "area": 105}},
		{PageContent: "London", Metadata: map[string]any{"population": 9.5, "area": 1572}},
		{PageContent: "Santiago", Metadata: map[string]any{"population": 6.9, "area": 641}},
		{PageContent: "Buenos Aires", Metadata: map[string]any{"population": 15.5, "area": 203}},
		{PageContent: "Rio de Janeiro", Metadata: map[string]any{"population": 13.7, "area": 1200}},
		{PageContent: "Sao Paulo", Metadata: map[string]any{"population": 22.6, "area": 1523}},
	}
	// create index and add test data
	docIDs, err := store.AddDocuments(ctx, data)
	require.NoError(t, err)
	assert.Equal(t, len(data), len(docIDs))

	// create vector with existed index
	store, err = redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, false),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(ctx, "Tokyo", 5)
	require.NoError(t, err)
	assert.Len(t, docs, 5)
	assert.Len(t, docs[0].Metadata, 3)

	// search with score threshold
	docs, err = store.SimilaritySearch(ctx, "Tokyo", 10,
		vectorstores.WithScoreThreshold(0.5),
	)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(docs), 1) // At least Tokyo itself should match
	assert.LessOrEqual(t, len(docs), 10)   // But not more than requested
	assert.Len(t, docs[0].Metadata, 3)

	// search with filter area>1000 or area < 300
	docs, err = store.SimilaritySearch(ctx, "Tokyo", 10,
		vectorstores.WithFilters("(@area:[(1000 +inf] | @area:[-inf (300])"),
	)
	require.NoError(t, err)
	assert.Len(t, docs, 5)
	assert.Len(t, docs[0].Metadata, 3)

	// search with filter area=622
	docs, err = store.SimilaritySearch(ctx, "Tokyo", 10,
		vectorstores.WithFilters("(@area:[622 622])"),
	)
	require.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Len(t, docs[0].Metadata, 3)

	// search with filter & score threshold
	docs, err = store.SimilaritySearch(ctx, "Tokyo", 2,
		vectorstores.WithFilters("(@area:[(1000 +inf] | @area:[-inf (300])"),
		vectorstores.WithScoreThreshold(0.5),
	)
	require.NoError(t, err)
	assert.Len(t, docs, 2)
	assert.Len(t, docs[0].Metadata, 3)

	t.Cleanup(func() {
		err = store.DropIndex(ctx, index, true)
		require.NoError(t, err)
	})
}

func TestRedisVectorAsRetriever(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisURL, ollamaURL := getValues(t)
	llm, e := getOllamaClient(ollamaEmbeddingModel, ollamaChatModel, ollamaURL)
	index := "test_redis_vector_as_retriever"

	store, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

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
			vectorstores.ToRetriever(store, 3),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	// The LLM should provide some response (not error) - exact content may vary
	require.NotEmpty(t, result, "expected non-empty result from LLM")

	result, err = chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	// The LLM should provide some response (not error) - exact content may vary
	require.NotEmpty(t, result, "expected non-empty result from LLM for furniture question")

	t.Cleanup(func() {
		err = store.DropIndex(ctx, index, true)
		require.NoError(t, err)
	})
}

func TestRedisVectorAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisURL, ollamaURL := getValues(t)
	_, e := getOllamaClient(ollamaEmbeddingModel, ollamaChatModel, ollamaURL)
	index := "test_redis_vector_as_retriever_with_metadata_filters"

	store, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	require.NoError(t, err)

	_, err = store.AddDocuments(
		ctx,
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is black.",
				Metadata: map[string]any{
					"location": "kitchen",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is blue.",
				Metadata: map[string]any{
					"location": "bedroom",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location": "office",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location": "sitting room",
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location": "patio",
				},
			},
		},
	)
	require.NoError(t, err)
	defer t.Cleanup(func() {
		err = store.DropIndex(ctx, index, true)
		require.NoError(t, err)
	})

	// Test that retrieval with filters works correctly (without LLM dependency)
	docs, err := store.SimilaritySearch(ctx, "lamp", 5,
		vectorstores.WithFilters("@location:(patio)"),
	)
	require.NoError(t, err)
	require.Len(t, docs, 1, "should find exactly one document with patio filter")
	require.Contains(t, docs[0].PageContent, "yellow", "the patio document should contain yellow")
	require.Equal(t, "patio", docs[0].Metadata["location"], "document should be from patio")
}

// nolint: unparam
func getOllamaClient(embeddingModel, chatModel string, connectionStr ...string) (llms.Model, *embeddings.EmbedderImpl) {
	// Create embedding LLM
	embOpts := []ollama.Option{ollama.WithModel(embeddingModel)}
	if len(connectionStr) > 0 {
		embOpts = append(embOpts, ollama.WithServerURL(connectionStr[0]))
	}
	embLlm, err := ollama.New(embOpts...)
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(embLlm)
	if err != nil {
		log.Fatal(err)
	}

	// Create chat LLM
	chatOpts := []ollama.Option{ollama.WithModel(chatModel)}
	if len(connectionStr) > 0 {
		chatOpts = append(chatOpts, ollama.WithServerURL(connectionStr[0]))
	}
	chatLlm, err := ollama.New(chatOpts...)
	if err != nil {
		log.Fatal(err)
	}

	return llms.Model(chatLlm), e
}

/**
func runOllamaTestContainer(model string) (*tcollama.OllamaContainer, string) {
	ctx := context.Background()

	ollamaContainer, err := tcollama.RunContainer(
		ctx,
		testcontainers.WithImage("ollama/ollama:0.1.31"),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	_, _, err = ollamaContainer.Exec(ctx, []string{"ollama", "pull", model})
	if err != nil {
		log.Fatalf("failed to pull model %s: %s", model, err)
	}

	connectionStr, err := ollamaContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err) // nolint:gocritic
	}
	return ollamaContainer, connectionStr
}
*/
