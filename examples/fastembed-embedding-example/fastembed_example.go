package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xDezzy/langchaingo/embeddings/fastembed"
	fastembed_lib "github.com/anush008/fastembed-go"
)

func main() {
	ctx := context.Background()

	embedder, err := fastembed.NewFastEmbed(
		fastembed.WithModel(fastembed_lib.BGESmallENV15),
		fastembed.WithCacheDir("./models"),
		fastembed.WithBatchSize(8),
		fastembed.WithMaxLength(256),
		fastembed.WithDocEmbedType("passage"),
		fastembed.WithShowDownloadProgress(true),
	)
	if err != nil {
		log.Fatalf("Failed to create FastEmbed: %v", err)
	}
	defer func() {
		if err := embedder.Close(); err != nil {
			log.Printf("Error closing embedder: %v", err)
		}
	}()

	documents := []string{
		"FastEmbed is a lightweight, fast, Python library built for embedding generation.",
		"It supports various pre-trained models including BGE and sentence-transformers.",
		"The library runs models locally using ONNX for optimal performance.",
		"FastEmbed can be used for semantic search, document similarity, and RAG applications.",
	}

	fmt.Println("Embedding documents...")
	docEmbeddings, err := embedder.EmbedDocuments(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to embed documents: %v", err)
	}

	fmt.Printf("Created embeddings for %d documents\n", len(docEmbeddings))
	for i, embedding := range docEmbeddings {
		fmt.Printf("Document %d: embedding dimension = %d, first 5 values = %v\n",
			i+1, len(embedding), embedding[:5])
	}

	query := "What is FastEmbed used for?"
	fmt.Printf("\nEmbedding query: '%s'\n", query)

	queryEmbedding, err := embedder.EmbedQuery(ctx, query)
	if err != nil {
		log.Fatalf("Failed to embed query: %v", err)
	}

	fmt.Printf("Query embedding dimension: %d, first 5 values = %v\n",
		len(queryEmbedding), queryEmbedding[:5])

	fmt.Println("\nCalculating similarity with documents...")
	for i, docEmb := range docEmbeddings {
		similarity := cosineSimilarity(queryEmbedding, docEmb)
		fmt.Printf("Document %d similarity: %.4f\n", i+1, similarity)
	}
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

func sqrt(x float32) float32 {
	if x == 0 {
		return 0
	}

	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}
