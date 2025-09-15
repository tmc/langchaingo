package v2

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	oldentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	tcmilvus "github.com/testcontainers/testcontainers-go/modules/milvus"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
	"github.com/tmc/langchaingo/internal/testutil/testctr"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// createOpenAIEmbedder creates an OpenAI embedder with httprr support for testing.
func createOpenAIEmbedder(t *testing.T) (llms.Model, *embeddings.EmbedderImpl) {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	llm, err := openai.New(openai.WithHTTPClient(rr.Client()))
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	return llm, e
}

func getTestStore(t *testing.T, ctx context.Context, e *embeddings.EmbedderImpl, opts ...Option) (Store, error) {
	t.Helper()
	testctr.SkipIfDockerNotAvailable(t)

	url := os.Getenv("MILVUS_URL")
	if url == "" {
		milvusContainer, err := tcmilvus.Run(ctx, "milvusdb/milvus:v2.4.0", testcontainers.WithLogger(tclog.TestLogger(t)))
		if err != nil {
			t.Skipf("Failed to start milvus container: %s", err)
		}
		t.Cleanup(func() {
			if err := milvusContainer.Terminate(context.Background()); err != nil {
				t.Logf("Failed to terminate milvus container: %v", err)
			}
		})

		url, err = milvusContainer.ConnectionString(ctx)
		if err != nil {
			t.Skipf("Failed to get milvus container endpoint: %s", err)
		}
	}

	// Test both v1 and v2 config compatibility
	var config interface{}
	if strings.Contains(t.Name(), "V1Config") {
		// Test v1 config compatibility
		config = client.Config{
			Address: url,
		}
	} else {
		// Test v2 config
		config = milvusclient.ClientConfig{
			Address: url,
		}
	}

	idx := index.NewAutoIndex(entity.L2)
	opts = append(
		opts,
		WithEmbedder(e),
		WithIndex(idx))
	return New(
		ctx,
		config,
		opts...,
	)
}

func TestV2ConfigCompatibility(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	store, err := getTestStore(t, ctx, e, WithCollectionName("test_v2_config"))
	require.NoError(t, err)
	require.NotNil(t, store.client)
}

func TestV1ConfigCompatibility(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	store, err := getTestStore(t, ctx, e, WithCollectionName("test_v1_config"))
	require.NoError(t, err)
	require.NotNil(t, store.client)
}

func TestV1IndexCompatibility(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	// Create v1 index and test conversion
	v1Index, err := oldentity.NewIndexAUTOINDEX(oldentity.L2)
	require.NoError(t, err)

	store, err := getTestStore(t, ctx, e,
		WithCollectionName("test_v1_index"),
		WithIndexV1(v1Index),
	)
	require.NoError(t, err)
	require.NotNil(t, store.client)
	require.NotNil(t, store.index)
}

func TestOptionFunctionsCompatibility(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	// Test all option functions work
	store, err := getTestStore(t, ctx, e,
		WithCollectionName("test_options"),
		WithPartitionName("test_partition"),
		WithTextField("content"),
		WithMetaField("metadata"),
		WithVectorField("embedding"),
		WithPrimaryField("id"),
		WithMaxTextLength(1000),
		WithShards(2),
		WithEF(20),
		WithMetricType(entity.L2),
		WithDropOld(),
		WithSkipFlushOnWrite(),
	)
	require.NoError(t, err)
	require.NotNil(t, store.client)

	// Verify options were applied
	require.Equal(t, "test_options", store.collectionName)
	require.Equal(t, "test_partition", store.partitionName)
	require.Equal(t, "content", store.textField)
	require.Equal(t, "metadata", store.metaField)
	require.Equal(t, "embedding", store.vectorField)
	require.Equal(t, "id", store.primaryField)
	require.Equal(t, 1000, store.maxTextLength)
	require.Equal(t, int32(2), store.shardNum)
	require.Equal(t, 20, store.ef)
	require.Equal(t, entity.L2, store.metricType)
	require.True(t, store.dropOld)
	require.True(t, store.skipFlushOnWrite)
}

func TestAddDocuments(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	store, err := getTestStore(t, ctx, e, WithCollectionName("test_add_docs"))
	require.NoError(t, err)

	docs := []schema.Document{
		{PageContent: "Document 1", Metadata: map[string]any{"source": "test1"}},
		{PageContent: "Document 2", Metadata: map[string]any{"source": "test2"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)
}

func TestSimilaritySearch(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	store, err := getTestStore(t, ctx, e, WithCollectionName("test_search"))
	require.NoError(t, err)

	// Add some documents first
	docs := []schema.Document{
		{PageContent: "The quick brown fox", Metadata: map[string]any{"type": "animal"}},
		{PageContent: "Jumps over the lazy dog", Metadata: map[string]any{"type": "action"}},
	}

	_, err = store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Perform similarity search
	results, err := store.SimilaritySearch(ctx, "fox", 1)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Contains(t, results[0].PageContent, "fox")
}

func TestVectorStoreInterface(t *testing.T) {
	ctx := context.Background()
	_, e := createOpenAIEmbedder(t)

	store, err := getTestStore(t, ctx, e, WithCollectionName("test_interface"))
	require.NoError(t, err)

	// Verify it implements the VectorStore interface
	var _ vectorstores.VectorStore = store
}

func TestConfigAdapterToV2Config(t *testing.T) {
	adapter := ConfigAdapter{}

	// Test v2 config pass-through
	v2Config := milvusclient.ClientConfig{Address: "localhost:19530"}
	result, err := adapter.ToV2Config(v2Config)
	require.NoError(t, err)
	require.Equal(t, v2Config, result)

	// Test v1 config conversion
	v1Config := client.Config{Address: "localhost:19530"}
	result, err = adapter.ToV2Config(v1Config)
	require.NoError(t, err)
	require.Equal(t, "localhost:19530", result.Address)

	// Test string address
	result, err = adapter.ToV2Config("localhost:19530")
	require.NoError(t, err)
	require.Equal(t, "localhost:19530", result.Address)

	// Test unsupported type
	_, err = adapter.ToV2Config(123)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported config type")
}