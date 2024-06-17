package mongodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/vectorstores/mongodb"
)

// TODO(ishan): use testcontainers to start a mongodb instance

func getOAIKey(t *testing.T) string {
	t.Helper()

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		t.Skip("OPENAI_API_KEY not set")
		return ""
	}

	return openaiAPIKey
}

func TestMongoDBStore(t *testing.T) {
	t.Parallel()

	_ = getOAIKey(t)

	llm, err := openai.New(openai.WithEmbeddingModel("text-embedding-ada-002"))
	require.NoError(t, err)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	store, err := mongodb.New(
		context.Background(),
		mongodb.WithEmbedder(e),
		mongodb.WithConnectionUri("mongodb+srv://langchaingotest:PVgzJSoESmj0Hzpv@langchaingotest.us7qrm4.mongodb.net/?retryWrites=true&w=majority&appName=langchaingotest"),
		mongodb.WithDatabaseName("langchaingo"),
		mongodb.WithCollectionName("test"),
		mongodb.WithIndexName("test"),
		mongodb.WithTextKey("text"),
		mongodb.WithRelevanceScoreFn("cosine"),
		mongodb.WithEmbeddingKey("embedding"),
	)
	if err != nil {
		t.Fatal(err)
	}

	docs := []mongodb.Document{
		{Text: "Tokyo", Metadata: map[string]interface{}{"country": "japan"}},
		{Text: "Potato", Metadata: map[string]interface{}{"country": "usa"}},
		{Text: "Paris", Metadata: map[string]interface{}{"country": "france"}},
		{Text: "London", Metadata: map[string]interface{}{"country": "uk"}},
		{Text: "New York", Metadata: map[string]interface{}{"country": "usa"}},
	}

	ids, err := store.AddDocuments(context.Background(), docs)
	require.NoError(t, err)

	fmt.Print(ids)
}
