package weaviate

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcweaviate "github.com/testcontainers/testcontainers-go/modules/weaviate"
	"github.com/vendasta/langchaingo/chains"
	"github.com/vendasta/langchaingo/embeddings"
	"github.com/vendasta/langchaingo/llms/openai"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate/entities/models"
)

func getValues(t *testing.T) (string, string) {
	t.Helper()

	scheme := os.Getenv("WEAVIATE_SCHEME")
	host := os.Getenv("WEAVIATE_HOST")
	if scheme == "" || host == "" {
		weaviateContainer, err := tcweaviate.RunContainer(context.Background(), testcontainers.WithImage("semitechnologies/weaviate:1.25.4"))
		if err != nil && strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			t.Skip("Docker not available")
		}
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, weaviateContainer.Terminate(context.Background()))
		})

		scheme, host, err = weaviateContainer.HttpHostAddress(context.Background())
		if err != nil {
			t.Skipf("Failed to get weaviate container endpoint: %s", err)
		}
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	return scheme, host
}

func randomizedCamelCaseClass() string {
	return "Test" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

func createTestClass(ctx context.Context, s Store) error {
	properties := make([]*models.Property, 0, len(s.queryAttrs))
	for _, v := range s.queryAttrs {
		properties = append(properties, &models.Property{
			Name:        v,
			Description: fmt.Sprintf("test property %s", v),
			DataType:    []string{"text"},
		})
	}
	return s.client.Schema().ClassCreator().WithClass(&models.Class{
		Class:       s.indexName,
		Description: "test class",
		VectorIndexConfig: map[string]any{
			"distance": "cosine",
		},
		ModuleConfig: map[string]any{},
		Properties:   properties,
	}).Do(ctx)
}

func TestWeaviateStoreRest(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"country"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	})
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(context.Background(), "japan", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}

func TestWeaviateStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
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
	})
	require.NoError(t, err)

	// test with a score threshold of 0.8, expected 6 documents
	docs, err := store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.9))
	require.NoError(t, err)
	require.Len(t, docs, 6)

	// test with a score threshold of 0, expected all 10 documents
	docs, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0))
	require.NoError(t, err)
	require.Len(t, docs, 10)
}

func TestMetadataSearch(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"type"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"type": "city",
		}},
		{PageContent: "potato", Metadata: map[string]any{
			"type": "vegetable",
		}},
	})
	require.NoError(t, err)

	docs, err := store.MetadataSearch(context.Background(), 2,
		vectorstores.WithFilters(
			filters.Where().
				WithPath([]string{"type"}).
				WithOperator(filters.Equal).
				WithValueString("city"),
		))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "city", docs[0].Metadata["type"])
}

func TestDeduplicater(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"type"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
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

	docs, err := store.MetadataSearch(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "potato", docs[0].PageContent)
	require.Equal(t, "vegetable", docs[0].Metadata["type"])
}

func TestSimilaritySearchWithInvalidScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)
	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
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
	})
	require.NoError(t, err)

	_, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(-0.8))
	require.Error(t, err)

	_, err = store.SimilaritySearch(context.Background(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(1.8))
	require.Error(t, err)
}

func TestWeaviateAsRetriever(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	nameSpace := uuid.New().String()

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 1, vectorstores.WithNameSpace(nameSpace)),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.True(t, strings.Contains(result, "orange"), "expected orange in result")
}

func TestWeaviateAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	nameSpace := randomizedCamelCaseClass()
	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(
				nameSpace), vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "beige", "expected beige in result")
}

func TestWeaviateAsRetrieverWithMetadataFilterEqualsClause(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"location"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	nameSpace := randomizedCamelCaseClass()

	_, err = store.AddDocuments(
		context.Background(),
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
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	filter := filters.Where().
		WithPath([]string{"location"}).
		WithOperator(filters.Equal).
		WithValueString("patio")

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store,
				5,
				vectorstores.WithNameSpace(nameSpace),
				vectorstores.WithFilters(filter)),
		),
		"What colors is the lamp?",
	)
	require.NoError(t, err)

	require.NotContains(t, result, "black", "expected black not in result")
	require.NotContains(t, result, "blue", "expected blue not in result")
	require.NotContains(t, result, "orange", "expected orange not in result")
	require.NotContains(t, result, "purple", "expected purple not in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestWeaviateAsRetrieverWithMetadataFilterNotSelected(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"location"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	nameSpace := randomizedCamelCaseClass()

	_, err = store.AddDocuments(
		context.Background(),
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
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithNameSpace(nameSpace)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)

	require.Contains(t, result, "black", "expected black in result")
	require.Contains(t, result, "blue", "expected blue in result")
	require.Contains(t, result, "orange", "expected orange in result")
	require.Contains(t, result, "purple", "expected purple in result")
	require.Contains(t, result, "yellow", "expected yellow in result")
}

func TestWeaviateAsRetrieverWithMetadataFilters(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"location"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	nameSpace := randomizedCamelCaseClass()

	_, err = store.AddDocuments(
		context.Background(),
		[]schema.Document{
			{
				PageContent: "The color of the lamp beside the desk is orange.",
				Metadata: map[string]any{
					"location":    "office",
					"square_feet": 100,
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is purple.",
				Metadata: map[string]any{
					"location":    "sitting room",
					"square_feet": 400,
				},
			},
			{
				PageContent: "The color of the lamp beside the desk is yellow.",
				Metadata: map[string]any{
					"location":    "patio",
					"square_feet": 800,
				},
			},
		},
		vectorstores.WithNameSpace(nameSpace),
	)
	require.NoError(t, err)

	filter := filters.Where().WithOperator(filters.And).WithOperands([]*filters.WhereBuilder{
		filters.Where().WithOperator(filters.Or).WithOperands([]*filters.WhereBuilder{
			filters.Where().WithPath([]string{"location"}).
				WithOperator(filters.Equal).WithValueString("office"),
			filters.Where().WithPath([]string{"location"}).
				WithOperator(filters.Equal).WithValueString("sitting room"),
		}),
		filters.Where().WithPath([]string{"square_feet"}).
			WithOperator(filters.GreaterThanEqual).WithValueNumber(300),
	})

	result, err := chains.Run(
		context.TODO(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store,
				5,
				vectorstores.WithFilters(filter),
				vectorstores.WithNameSpace(nameSpace)),
		),
		"What color is the lamp in each room?",
	)
	require.NoError(t, err)
	require.Contains(t, result, "purple", "expected purple in result")
	require.NotContains(t, result, "orange", "expected not orange in result")
	require.NotContains(t, result, "yellow", "expected not yellow in result")
}

func TestWeaviateStoreAdditionalFieldsDefaults(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Foo"},
	})
	require.NoError(t, err)

	// Check if the default additional fields are present in the result
	docs, err := store.SimilaritySearch(context.Background(),
		"Foo", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	additional, ok := docs[0].Metadata["_additional"].(map[string]any)
	require.True(t, ok, "expected '_additional' to be present in the metadata and parsable as 'map[string]any'")
	require.Len(t, additional, 1)

	certainty, _ := additional["certainty"].(float64)
	require.InDelta(t, docs[0].Score, float32(certainty), 0, "expect score to be equal to the certainty")
}

func TestWeaviateStoreAdditionalFieldsAdded(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(e),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithAdditionalFields([]string{"id", "vector", "certainty", "distance"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "Foo"},
	})
	require.NoError(t, err)

	// Check if all the additional fields are present in the result
	docs, err := store.SimilaritySearch(context.Background(),
		"Foo", 1)
	require.NoError(t, err)
	require.Len(t, docs, 1)

	additional, ok := docs[0].Metadata["_additional"].(map[string]any)
	require.True(t, ok, "expected '_additional' to be present in the metadata and parsable as 'map[string]any'")
	require.Len(t, additional, 4)

	require.NotEmpty(t, additional["id"], "expected the id to be present")
	require.NotEmpty(t, additional["vector"], "expected the vector to be present")
	require.NotEmpty(t, additional["certainty"], "expected the certainty to be present")
	require.NotEmpty(t, additional["distance"], "expected the distance to be present")
}

// TestWeaviateWithOptionEmbedder ensures that the embedder provided as an option to either
// `AddDocuments` or `SimilaritySearch` takes precedence over the one provided when creating
// the `Store`.
func TestWeaviateWithOptionEmbedder(t *testing.T) {
	t.Parallel()

	scheme, host := getValues(t)

	llm, err := openai.New()
	require.NoError(t, err)

	notme, err := embeddings.NewEmbedder(
		embeddings.EmbedderClientFunc(func(context.Context, []string) ([][]float32, error) {
			require.FailNow(t, "wrong embedder was called")
			return nil, nil
		}),
	)
	require.NoError(t, err)

	butme, err := embeddings.NewEmbedder(
		embeddings.EmbedderClientFunc(func(ctx context.Context, texts []string) ([][]float32, error) {
			return llm.CreateEmbedding(ctx, texts)
		}),
	)
	require.NoError(t, err)

	store, err := New(
		WithScheme(scheme),
		WithHost(host),
		WithEmbedder(notme),
		WithNameSpace(uuid.New().String()),
		WithIndexName(randomizedCamelCaseClass()),
		WithQueryAttrs([]string{"location"}),
	)
	require.NoError(t, err)

	err = createTestClass(context.Background(), store)
	require.NoError(t, err)

	_, err = store.AddDocuments(context.Background(), []schema.Document{
		{PageContent: "tokyo", Metadata: map[string]any{
			"country": "japan",
		}},
		{PageContent: "potato"},
	}, vectorstores.WithEmbedder(butme))
	require.NoError(t, err)

	docs, err := store.SimilaritySearch(context.Background(), "japan", 1,
		vectorstores.WithEmbedder(butme))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
	require.Equal(t, "japan", docs[0].Metadata["country"])
}
