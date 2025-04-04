package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/averikitsch/langchaingo/embeddings"
	"github.com/averikitsch/langchaingo/embeddings/cybertron"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/vectorstores"
	"github.com/averikitsch/langchaingo/vectorstores/weaviate"
	"github.com/chewxy/math32"
	"github.com/google/uuid"
)

func cosineSimilarity(x, y []float32) float32 {
	if len(x) != len(y) {
		log.Fatal("x and y have different lengths")
	}

	var dot, nx, ny float32

	for i := range x {
		nx += x[i] * x[i]
		ny += y[i] * y[i]
		dot += x[i] * y[i]
	}

	return dot / (math32.Sqrt(nx) * math32.Sqrt(ny))
}

func randomIndexName() string {
	return "Test" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

func exampleInMemory(ctx context.Context, emb embeddings.Embedder) {
	// We're going to create embeddings for the following strings, then calculate the similarity
	// between them using cosine-simularity.
	docs := []string{
		"tokyo",
		"japan",
		"potato",
	}

	vecs, err := emb.EmbedDocuments(ctx, docs)
	if err != nil {
		log.Fatal("embed query", err)
	}

	fmt.Println("Similarities:")

	for i := range docs {
		for j := range docs {
			fmt.Printf("%6s ~ %6s = %0.2f\n", docs[i], docs[j], cosineSimilarity(vecs[i], vecs[j]))
		}
	}
}

func exampleWeaviate(ctx context.Context, emb embeddings.Embedder) {
	scheme := os.Getenv("WEAVIATE_SCHEME")
	host := os.Getenv("WEAVIATE_HOST")

	if scheme == "" || host == "" {
		log.Print("Set WEAVIATE_HOST and WEAVIATE_SCHEME to run the weaviate example")

		return
	}

	// Create a new Weaviate vector store with the Cybertron Embedder to generate embeddings.
	store, err := weaviate.New(
		weaviate.WithEmbedder(emb),
		weaviate.WithScheme(scheme),
		weaviate.WithHost(host),
		weaviate.WithIndexName(randomIndexName()),
	)
	if err != nil {
		log.Fatal("create weaviate store", err)
	}

	// Add some documents to the vector store. This will use the Cybertron Embedder to create
	// embeddings for the documents.
	_, err = store.AddDocuments(ctx, []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "japan"},
		{PageContent: "potato"},
	})
	if err != nil {
		log.Fatal("add documents", err)
	}

	// Perform a similarity search, returning at most three results with similarity scores of
	// at least 0.8. This again uses the Cybertron Embedder to create an embedding for the
	// search query.
	matches, err := store.SimilaritySearch(ctx, "japan", 3,
		vectorstores.WithScoreThreshold(0.8),
	)
	if err != nil {
		log.Fatal("similarity search", err)
	}

	fmt.Println("Matches:")
	for _, match := range matches {
		fmt.Printf(" japan ~ %6s = %0.2f\n", match.PageContent, match.Score)
	}
}

func main() {
	ctx := context.Background()

	// Create an embedder client that uses the "BAAI/bge-small-en-v1.5" model and caches it in
	// the "models" directory. Cybertron will automatically download the model from HuggingFace
	// and convert it when needed.
	//
	// Note that not all models are supported and that Cybertron executes the model locally on
	// the CPU, so larger models will be quite slow!
	emc, err := cybertron.NewCybertron(
		cybertron.WithModelsDir("models"),
		cybertron.WithModel("BAAI/bge-small-en-v1.5"),
	)
	if err != nil {
		log.Fatal("create embedder client", err)
	}

	// Create an embedder from the previously created client.
	emb, err := embeddings.NewEmbedder(emc,
		embeddings.WithStripNewLines(false),
	)
	if err != nil {
		log.Fatal("create embedder", err)
	}

	// Example: use the Embedder to do an in-memory comparison between some documents.
	exampleInMemory(ctx, emb)

	// Example: use the Embedder together with a Vector Store.
	exampleWeaviate(ctx, emb)
}
