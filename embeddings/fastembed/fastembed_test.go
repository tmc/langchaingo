package fastembed

import (
	"context"
	"math"
	"testing"

	fastembed "github.com/anush008/fastembed-go"
)

func TestFastEmbed_Integration(t *testing.T) {
	embedder := createTestEmbedder(t)
	defer closeEmbedder(t, embedder)

	ctx := context.Background()

	t.Run("EmbedQuery", func(t *testing.T) {
		testEmbedQuery(t, embedder, ctx)
	})

	t.Run("EmbedDocuments", func(t *testing.T) {
		testEmbedDocuments(t, embedder, ctx)
	})

	t.Run("EmbedDocuments with passage type", func(t *testing.T) {
		testEmbedDocumentsPassage(t, ctx)
	})

	t.Run("StripNewLines", func(t *testing.T) {
		testStripNewLines(t, ctx)
	})
}

func createTestEmbedder(t *testing.T) *FastEmbed {
	embedder, err := NewFastEmbed(
		WithModel(fastembed.BGESmallENV15),
		WithBatchSize(2),
		WithMaxLength(128),
	)
	if err != nil {
		t.Fatalf("Failed to create FastEmbed: %v", err)
	}
	return embedder
}

func closeEmbedder(t *testing.T, embedder *FastEmbed) {
	if err := embedder.Close(); err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func testEmbedQuery(t *testing.T, embedder *FastEmbed, ctx context.Context) {
	query := "hello world"
	embedding, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		t.Fatalf("EmbedQuery() error: %v", err)
	}

	if len(embedding) == 0 {
		t.Errorf("EmbedQuery() returned empty embedding")
	}

	if len(embedding) != 384 {
		t.Errorf("EmbedQuery() expected embedding dimension 384, got %d", len(embedding))
	}

	expectedStart := []float32{0.01522374, -0.02271799, 0.00860278, -0.07424029, 0.00386434}
	epsilon := float64(1e-4)
	for i, expected := range expectedStart {
		if i >= len(embedding) {
			break
		}
		if math.Abs(float64(embedding[i]-expected)) > epsilon {
			t.Logf("Element %d: expected %.6f, got %.6f", i, expected, embedding[i])
		}
	}
}

func testEmbedDocuments(t *testing.T, embedder *FastEmbed, ctx context.Context) {
	docs := []string{"hello world", "goodbye world", "testing fastembed"}
	embeddings, err := embedder.EmbedDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("EmbedDocuments() error: %v", err)
	}

	if len(embeddings) != len(docs) {
		t.Errorf("EmbedDocuments() expected %d embeddings, got %d", len(docs), len(embeddings))
	}

	for i, emb := range embeddings {
		if len(emb) == 0 {
			t.Errorf("EmbedDocuments() embedding %d is empty", i)
		}
		if len(emb) != 384 {
			t.Errorf("EmbedDocuments() embedding %d expected dimension 384, got %d", i, len(emb))
		}
	}
}

func testEmbedDocumentsPassage(t *testing.T, ctx context.Context) {
	passageEmbedder, err := NewFastEmbed(
		WithModel(fastembed.BGESmallENV15),
		WithDocEmbedType("passage"),
		WithBatchSize(2),
	)
	if err != nil {
		t.Fatalf("Failed to create passage FastEmbed: %v", err)
	}
	defer passageEmbedder.Close()

	docs := []string{"This is a document", "Another document"}
	embeddings, err := passageEmbedder.EmbedDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("EmbedDocuments() with passage type error: %v", err)
	}

	if len(embeddings) != len(docs) {
		t.Errorf("EmbedDocuments() with passage type expected %d embeddings, got %d", len(docs), len(embeddings))
	}
}

func testStripNewLines(t *testing.T, ctx context.Context) {
	stripEmbedder, err := NewFastEmbed(
		WithModel(fastembed.BGESmallENV15),
		WithStripNewLines(true),
	)
	if err != nil {
		t.Fatalf("Failed to create strip newlines FastEmbed: %v", err)
	}
	defer stripEmbedder.Close()

	query := "hello\nworld"
	_, err = stripEmbedder.EmbedQuery(ctx, query)
	if err != nil {
		t.Fatalf("EmbedQuery() with newlines error: %v", err)
	}

	docs := []string{"hello\nworld", "goodbye\nworld"}
	_, err = stripEmbedder.EmbedDocuments(ctx, docs)
	if err != nil {
		t.Fatalf("EmbedDocuments() with newlines error: %v", err)
	}
}
