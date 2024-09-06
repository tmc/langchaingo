package mongovector

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	testDB                   = "langchaingo-test"
	testColl                 = "vstore"
	testIndexDP1536          = "vector_index_dotProduct_1536"
	testIndexDP1536NoFilters = "vector_index_dotProduct_1536_no_filters"
	testIndexSize1536        = 1536
	testIndexDP3             = "vector_index_dotProduct_3"
	testIndexSize3           = 3
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		opts                []Option
		wantIndex           string
		wantPageContentName string
		wantPath            string
	}{
		{
			name:                "nil options",
			opts:                nil,
			wantIndex:           "vector_index",
			wantPageContentName: "page_content",
			wantPath:            "plot_embedding",
		},
		{
			name:                "no options",
			opts:                []Option{},
			wantIndex:           "vector_index",
			wantPageContentName: "page_content",
			wantPath:            "plot_embedding",
		},
		{
			name:                "mixed custom options",
			opts:                []Option{WithIndex("custom_vector_index")},
			wantIndex:           "custom_vector_index",
			wantPageContentName: "page_content",
			wantPath:            "plot_embedding",
		},
		{
			name: "all custom options",
			opts: []Option{
				WithIndex("custom_vector_index"),
				WithPath("custom_plot_embedding"),
			},
			wantIndex: "custom_vector_index",
			wantPath:  "custom_plot_embedding",
		},
	}

	for _, test := range tests {
		test := test // Capture the range variable.

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			embedder, err := embeddings.NewEmbedder(&mockLLM{})
			assert.NoError(t, err, "failed to construct embedder")

			store := New(mongo.Collection{}, embedder, test.opts...)

			assert.Equal(t, test.wantIndex, store.index)
			assert.Equal(t, test.wantPath, store.path)
		})
	}
}

// resetVectorStore will reset the vector space defined by the given collection.
func resetVectorStore(t *testing.T, coll mongo.Collection, pageContentName string) {
	t.Helper()

	filter := bson.D{{Key: pageContentName, Value: bson.D{{Key: "$exists", Value: true}}}}

	_, err := coll.DeleteMany(context.Background(), filter)
	assert.NoError(t, err, "failed to reset vector store")
}

// setupTest will prepare the Atlas vector search for adding to and searching
// a vector space.
func setupTest(t *testing.T, dim int, index string) (Store, *mockEmbedder) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("Must set MONGODB_URI to run test")
	}

	require.NotEmpty(t, uri, "MONGODB_URI required")

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err, "failed to connect to MongoDB server")

	// Create the vectorstore collection
	err = client.Database(testDB).CreateCollection(context.Background(), testColl)
	assert.NoError(t, err, "failed to create collection")

	coll := client.Database(testDB).Collection(testColl)
	resetVectorStore(t, *coll, pageContentName)

	emb := newMockEmbedder(dim, "")
	store := New(*coll, emb, WithIndex(index))

	return store, emb
}

func TestStore_AddDocuments(t *testing.T) {
	store, _ := setupTest(t, 0, testIndexDP1536)

	tests := []struct {
		name    string
		docs    []schema.Document
		options []vectorstores.Option
		wantErr []string
	}{
		{
			name:    "nil docs",
			docs:    nil,
			wantErr: []string{"must provide at least one element in input slice"},
			options: []vectorstores.Option{},
		},
		{
			name:    "no docs",
			docs:    []schema.Document{},
			wantErr: []string{"must provide at least one element in input slice"},
			options: []vectorstores.Option{},
		},
		{
			name:    "single empty doc",
			docs:    []schema.Document{{}},
			wantErr: []string{}, // May vary by embedder
			options: []vectorstores.Option{},
		},
		{
			name:    "single non-empty doc",
			docs:    []schema.Document{{PageContent: "foo"}},
			wantErr: []string{},
			options: []vectorstores.Option{},
		},
		{
			name:    "one non-empty doc and one empty doc",
			docs:    []schema.Document{{PageContent: "foo"}, {}},
			wantErr: []string{}, // May vary by embedder
			options: []vectorstores.Option{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetVectorStore(t, store.coll, pageContentName)

			ids, err := store.AddDocuments(context.Background(), test.docs, test.options...)
			if len(test.wantErr) > 0 {
				require.Error(t, err)
				for _, want := range test.wantErr {
					if strings.Contains(err.Error(), want) {
						return
					}
				}

				t.Errorf("expected error %q to contain of %v", err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, len(test.docs), len(ids))
		})
	}

}

type simSearchTest struct {
	ctx          context.Context
	seed         []schema.Document
	numDocuments int                   // Number of documents to return
	options      []vectorstores.Option // Search query options
	want         []schema.Document
	wantErr      string
}

func runSimilaritySearchTest(t *testing.T, test simSearchTest) {
	t.Helper()

	store, emb := setupTest(t, testIndexSize1536, testIndexDP1536)
	for _, doc := range test.seed {
		emb.addDocument(doc)
	}

	err := emb.flush(context.Background(), store)
	require.NoError(t, err)

	// Merge options
	opts := vectorstores.Options{}
	for _, opt := range test.options {
		opt(&opts)
	}

	if opts.Embedder != nil {
		err = opts.Embedder.(*mockEmbedder).flush(context.Background(), store)
		require.NoError(t, err, "failed to flush custom embedder")
	} else {
		test.options = append(test.options, vectorstores.WithEmbedder(emb))
	}

	raw, err := store.SimilaritySearch(test.ctx, "", test.numDocuments, test.options...)
	if test.wantErr != "" {
		assert.Error(t, err)
		assert.ErrorContains(t, err, test.wantErr)
	} else {
		require.NoError(t, err)
	}

	assert.Len(t, raw, len(test.want))

	got := make(map[string]schema.Document)
	for _, g := range raw {
		got[g.PageContent] = g
	}

	for _, w := range test.want {
		got := got[w.PageContent]
		if w.Score != 0 {
			assert.InDelta(t, w.Score, got.Score, 1e-4, "score out of bounds for %w", w.PageContent)
		}

		assert.Equal(t, w.PageContent, got.PageContent, "page contents differ")
		assert.Equal(t, w.Metadata, got.Metadata, "metadata differs")
	}
}

func TestStore_SimilaritySearch_ExactQuery(t *testing.T) {
	seed := []schema.Document{
		{PageContent: "v1", Score: 1},
		{PageContent: "v090", Score: 0.90},
		{PageContent: "v051", Score: 0.51},
		{PageContent: "v0001", Score: 0.001},
	}

	t.Run("numDocuments=1 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         seed,
				want: []schema.Document{
					{PageContent: "v1", Score: 1},
				},
			})
	})

	t.Run("numDocuments=3 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				want: []schema.Document{
					{PageContent: "v1", Score: 1},
					{PageContent: "v090", Score: 0.90},
					{PageContent: "v051", Score: 0.51},
				},
			})
	})
}

func TestStore_SimilaritySearch_NonExactQuery(t *testing.T) {
	seed := []schema.Document{
		{PageContent: "v090", Score: 0.90},
		{PageContent: "v051", Score: 0.51},
		{PageContent: "v0001", Score: 0.001},
	}

	t.Run("numDocuments=1 of 3", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         seed,
				want:         seed[:1],
			})
	})

	t.Run("numDocuments=3 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				want:         seed,
			})
	})

	t.Run("with score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				options:      []vectorstores.Option{vectorstores.WithScoreThreshold(0.50)},
				want:         seed[:2],
			})
	})

	t.Run("with invalid score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				options:      []vectorstores.Option{vectorstores.WithScoreThreshold(-0.50)},
				wantErr:      ErrInvalidScoreThreshold.Error(),
			})
	})

	metadataSeed := []schema.Document{
		{PageContent: "v090", Score: 0.90},
		{PageContent: "v051", Score: 0.51, Metadata: map[string]any{"pi": 3.14}},
		{PageContent: "v0001", Score: 0.001},
	}

	t.Run("with metadata", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         metadataSeed,
				want:         metadataSeed,
			})
	})

	t.Run("with metadata and score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 3,
				seed:         metadataSeed,
				want:         metadataSeed[:2],
				options:      []vectorstores.Option{vectorstores.WithScoreThreshold(0.50)},
			})
	})

	t.Run("with namespace", func(t *testing.T) {
		emb := newMockEmbedder(testIndexSize3, "")

		doc := schema.Document{PageContent: "v090", Score: 0.90, Metadata: map[string]any{"phi": 1.618}}
		emb.addDocument(doc)

		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         []schema.Document{doc},
				want:         []schema.Document{doc},
				options: []vectorstores.Option{
					vectorstores.WithNameSpace(testIndexDP3),
					vectorstores.WithEmbedder(emb),
				},
			})
	})

	t.Run("with non-existant namespace", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithNameSpace("some-nonexistant-index-name"),
				},
			})
	})

	t.Run("with filter", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				want:         metadataSeed[len(metadataSeed)-1:],
				options: []vectorstores.Option{
					vectorstores.WithFilters(bson.D{{Key: "pageContent", Value: "v0001"}}),
				},
			})
	})

	t.Run("with non-tokenized filter", func(t *testing.T) {
		emb := newMockEmbedder(testIndexSize1536, "")

		doc := schema.Document{PageContent: "v090", Score: 0.90, Metadata: map[string]any{"phi": 1.618}}
		emb.addDocument(doc)

		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithFilters(bson.D{{Key: "pageContent", Value: "v0001"}}),
					vectorstores.WithNameSpace(testIndexDP1536NoFilters),
					vectorstores.WithEmbedder(emb),
				},
				wantErr: "'pageContent' needs to be indexed as token",
			})
	})

	t.Run("with deduplicator", func(t *testing.T) {
		runSimilaritySearchTest(t,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithDeduplicater(func(ctx context.Context, doc schema.Document) bool { return true }),
				},
				wantErr: ErrUnsupportedOptions.Error(),
			})
	})
}
