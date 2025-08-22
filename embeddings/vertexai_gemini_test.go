package embeddings

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/llms/googleai/vertex"
)

func skipIfNoVertexCredentials(t *testing.T) {
	t.Helper()

	if os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
		t.Skip("GOOGLE_CLOUD_PROJECT not set, skipping Vertex AI test")
	}
	if os.Getenv("GOOGLE_CLOUD_LOCATION") == "" {
		t.Skip("GOOGLE_CLOUD_LOCATION not set, skipping Vertex AI test")
	}
	// Check for Application Default Credentials
	if _, exists := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); !exists {
		// Check if running in a GCP environment
		if os.Getenv("GOOGLE_CLOUD_PROJECT") == "" {
			t.Skip("No Google Cloud credentials available, skipping Vertex AI test")
		}
	}
}

func TestVertexAIGeminiEmbeddings(t *testing.T) {
	skipIfNoVertexCredentials(t)

	ctx := context.Background()

	// Create Vertex AI client with Gemini embeddings
	llm, err := vertex.New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer llm.Close()

	embedder, err := NewEmbedder(llm)
	if err != nil {
		t.Fatalf("Failed to create embedder: %v", err)
	}

	// Test single embedding
	embedding, err := embedder.EmbedQuery(ctx, "Hello world!")
	if err != nil {
		t.Fatalf("Failed to embed query: %v", err)
	}
	if len(embedding) == 0 {
		t.Error("Expected non-empty embedding")
	}

	// Test batch embeddings
	texts := []string{
		"The quick brown fox jumps over the lazy dog",
		"Machine learning is a subset of artificial intelligence",
		"Golang is a statically typed programming language",
	}
	embeddings, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to embed documents: %v", err)
	}
	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}

	// Verify all embeddings have the same dimension
	if len(embeddings) > 0 {
		dim := len(embeddings[0])
		for i, emb := range embeddings {
			if len(emb) != dim {
				t.Errorf("Embedding %d has different dimension: expected %d, got %d", i, dim, len(emb))
			}
		}
	}
}

func TestVertexAIGeminiEmbeddingsWithOptions(t *testing.T) {
	skipIfNoVertexCredentials(t)

	ctx := context.Background()

	// Create Vertex AI client with specific embedding model
	llm, err := vertex.New(ctx)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer llm.Close()

	// Create embedder with options
	embedder, err := NewEmbedder(llm,
		WithBatchSize(5),
		WithStripNewLines(false),
	)
	if err != nil {
		t.Fatalf("Failed to create embedder with options: %v", err)
	}

	// Test with newlines in text
	textWithNewlines := "This is line one.\nThis is line two.\nThis is line three."
	embedding, err := embedder.EmbedQuery(ctx, textWithNewlines)
	if err != nil {
		t.Fatalf("Failed to embed text with newlines: %v", err)
	}
	if len(embedding) == 0 {
		t.Error("Expected non-empty embedding for text with newlines")
	}

	// Test batch with specific size
	texts := []string{
		"First document",
		"Second document",
		"Third document",
		"Fourth document",
		"Fifth document",
		"Sixth document",
	}
	embeddings, err := embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to embed batch of documents: %v", err)
	}
	if len(embeddings) != 6 {
		t.Errorf("Expected 6 embeddings, got %d", len(embeddings))
	}
}
