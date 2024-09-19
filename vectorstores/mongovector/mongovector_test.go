package mongovector

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	testURI                   = "MONGODB_VECTOR_TEST_URI"
	testDB                    = "langchaingo-test"
	testColl                  = "vstore"
	testIndexDP1536           = "vector_index_dotProduct_1536"
	testIndexDP1536WithFilter = "vector_index_dotProduct_1536_w_filters"
	testIndexDP3              = "vector_index_dotProduct_3"
	testIndexSize1536         = 1536
	testIndexSize3            = 3
)

type atlasContainer struct {
	testcontainers.Container
	URI string
}

func setupAtlas(ctx context.Context) (*atlasContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongodb/mongodb-atlas-local",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(1 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	var atlasC *atlasContainer
	if container != nil {
		atlasC = &atlasContainer{Container: container}
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return atlasC, err
	}

	mappedPort, err := container.MappedPort(ctx, "27017")
	if err != nil {
		return atlasC, err
	}

	uri := &url.URL{
		Scheme:   "mongodb",
		Host:     net.JoinHostPort(ip, mappedPort.Port()),
		Path:     "/",
		RawQuery: "directConnection=true",
	}

	atlasC.URI = uri.String()

	return atlasC, nil
}

// resetVectorStore will reset the vector space defined by the given collection.
func resetVectorStore(t *testing.T, coll *mongo.Collection) {
	t.Helper()

	filter := bson.D{{Key: pageContentName, Value: bson.D{{Key: "$exists", Value: true}}}}

	_, err := coll.DeleteMany(context.Background(), filter)
	assert.NoError(t, err, "failed to reset vector store")
}

// setupTest will prepare the Atlas vector search for adding to and searching
// a vector space.
func setupTest(t *testing.T, dim int, index string) Store {
	t.Helper()

	uri := os.Getenv(testURI)
	if uri == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		container, err := setupAtlas(ctx)
		require.NoError(t, err)

		uri = container.URI
		os.Setenv(testURI, uri)
	}

	require.NotEmpty(t, uri, "URI required")

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	require.NoError(t, err, "failed to connect to MongoDB server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	require.NoError(t, err, "failed to ping server")

	time.Sleep(10 * time.Second) // Let the container warm up

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = resetForE2E(ctx, client, testIndexDP1536, testIndexSize1536, nil)
	require.NoError(t, err)

	filters := []string{"pageContent"}
	err = resetForE2E(ctx, client, testIndexDP1536WithFilter, testIndexSize1536, filters)
	require.NoError(t, err)

	err = resetForE2E(ctx, client, testIndexDP3, testIndexSize3, nil)
	require.NoError(t, err)

	// Create the vectorstore collection
	err = client.Database(testDB).CreateCollection(context.Background(), testColl)
	require.NoError(t, err, "failed to create collection")

	coll := client.Database(testDB).Collection(testColl)
	resetVectorStore(t, coll)

	emb := newMockEmbedder(dim)
	store := New(coll, emb, WithIndex(index))

	return store
}

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
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			embedder, err := embeddings.NewEmbedder(&mockLLM{})
			require.NoError(t, err, "failed to construct embedder")

			store := New(&mongo.Collection{}, embedder, test.opts...)

			assert.Equal(t, test.wantIndex, store.index)
			assert.Equal(t, test.wantPath, store.path)
		})
	}
}

//nolint:paralleltest
func TestStore_AddDocuments(t *testing.T) {
	store := setupTest(t, testIndexSize1536, testIndexDP1536)

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
			resetVectorStore(t, store.coll)

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
				require.NoError(t, err)
			}

			assert.Equal(t, len(test.docs), len(ids))
		})
	}
}

type simSearchTest struct {
	ctx          context.Context //nolint:containedctx
	seed         []schema.Document
	numDocuments int                   // Number of documents to return
	options      []vectorstores.Option // Search query options
	want         []schema.Document
	wantErr      string
}

func runSimilaritySearchTest(t *testing.T, store Store, test simSearchTest) {
	t.Helper()

	resetVectorStore(t, store.coll)

	// Merge options
	opts := vectorstores.Options{}
	for _, opt := range test.options {
		opt(&opts)
	}

	var emb *mockEmbedder
	if opts.Embedder != nil {
		var ok bool

		emb, ok = opts.Embedder.(*mockEmbedder)
		require.True(t, ok)
	} else {
		semb, ok := store.embedder.(*mockEmbedder)
		require.True(t, ok)

		emb = newMockEmbedder(len(semb.queryVector))
		emb.mockDocuments(test.seed...)

		test.options = append(test.options, vectorstores.WithEmbedder(emb))
	}

	err := flushMockDocuments(context.Background(), store, emb)
	require.NoError(t, err, "failed to flush mock embedder")

	raw, err := store.SimilaritySearch(test.ctx, "", test.numDocuments, test.options...)
	if test.wantErr != "" {
		require.Error(t, err)
		require.ErrorContains(t, err, test.wantErr)
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

//nolint:paralleltest
func TestStore_SimilaritySearch_ExactQuery(t *testing.T) {
	store := setupTest(t, testIndexSize3, testIndexDP3)

	seed := []schema.Document{
		{PageContent: "v1", Score: 1},
		{PageContent: "v090", Score: 0.90},
		{PageContent: "v051", Score: 0.51},
		{PageContent: "v0001", Score: 0.001},
	}

	t.Run("numDocuments=1 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         seed,
				want: []schema.Document{
					{PageContent: "v1", Score: 1},
				},
			})
	})

	t.Run("numDocuments=3 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
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

//nolint:funlen,paralleltest
func TestStore_SimilaritySearch_NonExactQuery(t *testing.T) {
	store := setupTest(t, testIndexSize1536, testIndexDP1536)

	seed := []schema.Document{
		{PageContent: "v090", Score: 0.90},
		{PageContent: "v051", Score: 0.51},
		{PageContent: "v0001", Score: 0.001},
	}

	t.Run("numDocuments=1 of 3", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         seed,
				want:         seed[:1],
			})
	})

	t.Run("numDocuments=3 of 4", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				want:         seed,
			})
	})

	t.Run("with score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 3,
				seed:         seed,
				options:      []vectorstores.Option{vectorstores.WithScoreThreshold(0.50)},
				want:         seed[:2],
			})
	})

	t.Run("with invalid score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
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
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 3,
				seed:         metadataSeed,
				want:         metadataSeed,
			})
	})

	t.Run("with metadata and score threshold", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 3,
				seed:         metadataSeed,
				want:         metadataSeed[:2],
				options:      []vectorstores.Option{vectorstores.WithScoreThreshold(0.50)},
			})
	})

	t.Run("with namespace", func(t *testing.T) {
		emb := newMockEmbedder(testIndexSize3)

		doc := schema.Document{PageContent: "v090", Score: 0.90, Metadata: map[string]any{"phi": 1.618}}
		emb.mockDocuments(doc)

		runSimilaritySearchTest(t, store,
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

	t.Run("with non-existent namespace", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithNameSpace("some-non-existent-index-name"),
				},
			})
	})

	t.Run("with filter", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				want:         metadataSeed[len(metadataSeed)-1:],
				options: []vectorstores.Option{
					vectorstores.WithFilters(bson.D{{Key: "pageContent", Value: "v0001"}}),
					vectorstores.WithNameSpace(testIndexDP1536WithFilter),
				},
			})
	})

	t.Run("with non-tokenized filter", func(t *testing.T) {
		emb := newMockEmbedder(testIndexSize1536)

		doc := schema.Document{PageContent: "v090", Score: 0.90, Metadata: map[string]any{"phi": 1.618}}
		emb.mockDocuments(doc)

		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithFilters(bson.D{{Key: "pageContent", Value: "v0001"}}),
					vectorstores.WithEmbedder(emb),
				},
				wantErr: "'pageContent' needs to be indexed as token",
			})
	})

	t.Run("with deduplicator", func(t *testing.T) {
		runSimilaritySearchTest(t, store,
			simSearchTest{
				numDocuments: 1,
				seed:         metadataSeed,
				options: []vectorstores.Option{
					vectorstores.WithDeduplicater(func(context.Context, schema.Document) bool { return true }),
				},
				wantErr: ErrUnsupportedOptions.Error(),
			})
	})
}

// vectorField defines the fields of an index used for vector search.
type vectorField struct {
	Type          string `bson:"type,omitempty"`
	Path          string `bson:"path,omityempty"`
	NumDimensions int    `bson:"numDimensions,omitempty"`
	Similarity    string `bson:"similarity,omitempty"`
}

// createVectorSearchIndex will create a vector search index on the "db.vstore"
// collection named "vector_index" with the provided field. This function blocks
// until the index has been created.
func createVectorSearchIndex(
	ctx context.Context,
	coll *mongo.Collection,
	idxName string,
	fields ...vectorField,
) (string, error) {
	def := struct {
		Fields []vectorField `bson:"fields"`
	}{
		Fields: fields,
	}

	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName(idxName).SetType("vectorSearch")
	searchName, err := view.CreateOne(ctx, mongo.SearchIndexModel{Definition: def, Options: siOpts})
	if err != nil {
		return "", fmt.Errorf("failed to create the search index: %w", err)
	}

	// Await the creation of the index.
	var doc bson.Raw
	for doc == nil {
		cursor, err := view.List(ctx, options.SearchIndexes().SetName(searchName))
		if err != nil {
			return "", fmt.Errorf("failed to list search indexes: %w", err)
		}

		if !cursor.Next(ctx) {
			break
		}

		name := cursor.Current.Lookup("name").StringValue()
		queryable := cursor.Current.Lookup("queryable").Boolean()
		if name == searchName && queryable {
			doc = cursor.Current
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	return searchName, nil
}

func searchIndexExists(ctx context.Context, coll *mongo.Collection, idx string) (bool, error) {
	view := coll.SearchIndexes()

	siOpts := options.SearchIndexes().SetName(idx).SetType("vectorSearch")
	cursor, err := view.List(ctx, siOpts)
	if err != nil {
		return false, fmt.Errorf("failed to list search indexes: %w", err)
	}

	if cursor == nil || cursor.Current == nil {
		return false, nil
	}

	name := cursor.Current.Lookup("name").StringValue()
	queryable := cursor.Current.Lookup("queryable").Boolean()

	return name == idx && queryable, nil
}

func resetForE2E(ctx context.Context, client *mongo.Client, idx string, dim int, filters []string) error {
	// Create the vectorstore collection
	err := client.Database(testDB).CreateCollection(ctx, testColl)
	if err != nil {
		return fmt.Errorf("failed to create vector store collection: %w", err)
	}

	coll := client.Database(testDB).Collection(testColl)

	if ok, _ := searchIndexExists(ctx, coll, idx); ok {
		return nil
	}

	fields := []vectorField{}

	fields = append(fields, vectorField{
		Type:          "vector",
		Path:          "plot_embedding",
		NumDimensions: dim,
		Similarity:    "dotProduct",
	})

	for _, filter := range filters {
		fields = append(fields, vectorField{
			Type: "filter",
			Path: filter,
		})
	}

	_, err = createVectorSearchIndex(ctx, coll, idx, fields...)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
