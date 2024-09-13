package mongovector_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/mongovector"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	testIndexSize1536 = 1536
	testIndexSize3    = 3
)

func setupMongoDBContainer(ctx context.Context) (string, func(), error) {
	mongodbContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:4.4"),
	)
	if err != nil {
		return "", nil, err
	}

	connectionString, err := mongodbContainer.ConnectionString(ctx)
	if err != nil {
		return "", nil, err
	}

	cleanup := func() {
		if err := mongodbContainer.Terminate(ctx); err != nil {
			panic(err)
		}
	}

	return connectionString, cleanup, nil
}

func setupTest(t *testing.T, dim int) (mongovector.Store, func()) {
	t.Helper()

	ctx := context.Background()
	connectionString, cleanup, err := setupMongoDBContainer(ctx)
	require.NoError(t, err)

	client, err := mongo.Connect(options.Client().ApplyURI(connectionString))
	require.NoError(t, err)

	coll := client.Database("test").Collection("test_collection")
	emb := newMockEmbedder(dim)
	store := mongovector.New(coll, emb)

	return store, func() {
		cleanup()
		client.Disconnect(ctx)
	}
}

func TestStore_AddDocuments(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		docs    []schema.Document
		options []vectorstores.Option
		wantErr bool
	}{
		{
			name: "add single document",
			docs: []schema.Document{
				{PageContent: "test document"},
			},
			wantErr: false,
		},
		{
			name: "add multiple documents",
			docs: []schema.Document{
				{PageContent: "test document 1"},
				{PageContent: "test document 2"},
			},
			wantErr: false,
		},
		{
			name:    "add empty document list",
			docs:    []schema.Document{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids, err := store.AddDocuments(ctx, tt.docs, tt.options...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, ids, len(tt.docs))
			}
		})
	}
}

func TestStore_SimilaritySearch(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	// Add test documents
	docs := []schema.Document{
		{PageContent: "The quick brown fox jumps over the lazy dog", Metadata: map[string]interface{}{"animal": "fox"}},
		{PageContent: "The lazy dog is sleeping", Metadata: map[string]interface{}{"animal": "dog"}},
		{PageContent: "The quick brown fox is running", Metadata: map[string]interface{}{"animal": "fox"}},
	}
	_, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	// Wait for documents to be indexed
	time.Sleep(2 * time.Second)

	tests := []struct {
		name           string
		query          string
		numDocuments   int
		options        []vectorstores.Option
		expectedCount  int
		expectedAnimal string
		wantErr        bool
	}{
		{
			name:           "basic search",
			query:          "quick fox",
			numDocuments:   2,
			expectedCount:  2,
			expectedAnimal: "fox",
			wantErr:        false,
		},
		{
			name:           "search with score threshold",
			query:          "lazy dog",
			numDocuments:   3,
			options:        []vectorstores.Option{vectorstores.WithScoreThreshold(0.8)},
			expectedCount:  1,
			expectedAnimal: "dog",
			wantErr:        false,
		},
		{
			name:           "search with filter",
			query:          "animal",
			numDocuments:   3,
			options:        []vectorstores.Option{vectorstores.WithFilters(bson.M{"animal": "fox"})},
			expectedCount:  2,
			expectedAnimal: "fox",
			wantErr:        false,
		},
		{
			name:         "search with invalid score threshold",
			query:        "test",
			numDocuments: 1,
			options:      []vectorstores.Option{vectorstores.WithScoreThreshold(2.0)},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.SimilaritySearch(ctx, tt.query, tt.numDocuments, tt.options...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.expectedCount)
				if tt.expectedAnimal != "" {
					assert.Equal(t, tt.expectedAnimal, results[0].Metadata["animal"])
				}
			}
		})
	}
}

func TestStore_SimilaritySearch_ExactQuery(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize3)
	defer cleanup()

	ctx := context.Background()

	docs := []schema.Document{
		{PageContent: "Document 1", Metadata: map[string]interface{}{"score": 1.0}},
		{PageContent: "Document 2", Metadata: map[string]interface{}{"score": 0.9}},
		{PageContent: "Document 3", Metadata: map[string]interface{}{"score": 0.8}},
	}
	_, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	results, err := store.SimilaritySearch(ctx, "Document 1", 1)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Document 1", results[0].PageContent)
	assert.InDelta(t, 1.0, results[0].Metadata["score"], 0.01)
}

func TestStore_SimilaritySearch_EmptyResult(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	results, err := store.SimilaritySearch(ctx, "Non-existent document", 1)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestStore_SimilaritySearch_DifferentNamespaces(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	// Add documents to different namespaces
	docs1 := []schema.Document{
		{PageContent: "Document in namespace 1"},
	}
	docs2 := []schema.Document{
		{PageContent: "Document in namespace 2"},
	}

	_, err := store.AddDocuments(ctx, docs1, vectorstores.WithNameSpace("namespace1"))
	require.NoError(t, err)
	_, err = store.AddDocuments(ctx, docs2, vectorstores.WithNameSpace("namespace2"))
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	results1, err := store.SimilaritySearch(ctx, "Document", 1, vectorstores.WithNameSpace("namespace1"))
	require.NoError(t, err)
	assert.Len(t, results1, 1)
	assert.Contains(t, results1[0].PageContent, "namespace 1")

	results2, err := store.SimilaritySearch(ctx, "Document", 1, vectorstores.WithNameSpace("namespace2"))
	require.NoError(t, err)
	assert.Len(t, results2, 1)
	assert.Contains(t, results2[0].PageContent, "namespace 2")
}

func TestStore_SimilaritySearch_MaxDocuments(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	// Add 10 documents
	docs := make([]schema.Document, 10)
	for i := 0; i < 10; i++ {
		docs[i] = schema.Document{PageContent: fmt.Sprintf("Document %d", i+1)}
	}

	_, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Try to retrieve more documents than exist
	results, err := store.SimilaritySearch(ctx, "Document", 15)
	require.NoError(t, err)
	assert.Len(t, results, 10) // Should only return the 10 existing documents
}

func TestStore_SimilaritySearch_Concurrent(t *testing.T) {
	store, cleanup := setupTest(t, testIndexSize1536)
	defer cleanup()

	ctx := context.Background()

	// Add some documents
	docs := []schema.Document{
		{PageContent: "Document 1"},
		{PageContent: "Document 2"},
		{PageContent: "Document 3"},
	}
	_, err := store.AddDocuments(ctx, docs)
	require.NoError(t, err)

	time.Sleep(2 * time.Second)

	// Perform concurrent searches
	concurrency := 10
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			results, err := store.SimilaritySearch(ctx, "Document", 1)
			assert.NoError(t, err)
			assert.Len(t, results, 1)
		}()
	}

	wg.Wait()
}
