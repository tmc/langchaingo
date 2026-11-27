// Package main demonstrates how to use Langchaingo with Google Gemini
// for text embeddings and Milvus as a vector store. It inserts documents,
// performs a similarity search, and generates an LLM-based response.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	milvus "github.com/tmc/langchaingo/vectorstores/milvus/v2"
)

func main() {
	ctx := context.Background()

	// Initialize the Google Gemini LLM client using the GEMINI_API_KEY
	// environment variable. The same client will be used both for generating
	// embeddings and for LLM text generation.
	llm, err := googleai.New(
		ctx,
		googleai.WithAPIKey(os.Getenv("GEMINI_API_KEY")),
		googleai.WithDefaultEmbeddingModel("gemini-embedding-001"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create an embedder instance using the LLM.
	// The embedder generates vector representations for input text.
	embeddingModel, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}

	// Configure Milvus client connection parameters.
	// Requires MILVUS_URL and optionally MILVUS_API_KEY environment variables.
	config := milvusclient.ClientConfig{
		Address: os.Getenv("MILVUS_URL"),
		APIKey:  os.Getenv("MILVUS_API_KEY"), // Only required for Milvus cloud instances
	}

	// Define an automatic index with L2 (Euclidean) distance metric.
	idx := index.NewAutoIndex(entity.L2)

	// Initialize the Milvus vector store, providing:
	// - the embedder for vectorization
	// - a collection name ("docs")
	// - the index configuration
	store, err := milvus.New(
		ctx,
		config,
		milvus.WithEmbedder(embeddingModel),
		milvus.WithCollectionName("docs"),
		milvus.WithIndex(idx),
	)
	if err != nil {
		log.Fatal(err, store)
	}

	// Prepare example documents for insertion into Milvus.
	docs := []schema.Document{
		{
			PageContent: "The capital of France is Paris. It is known for its culture, art, and cuisine.",
			Metadata:    map[string]any{"topic": "Geography", "source": "notes"},
		},
		{
			PageContent: "Machine learning is a branch of artificial intelligence that enables systems to learn from data and improve over time without being explicitly programmed.",
			Metadata:    map[string]any{"topic": "Technology", "source": "notes"},
		},
		{
			PageContent: "Photosynthesis is the process by which green plants use sunlight to synthesize nutrients from carbon dioxide and water.",
			Metadata:    map[string]any{"topic": "Biology", "source": "notes"},
		},
	}

	// Insert the documents into Milvus.
	ids, err := store.AddDocuments(ctx, docs)
	if err != nil {
		log.Fatal(err, ids)
	}

	// Define a query for similarity search.
	query := "What is machine learning?"

	// Perform similarity search in Milvus to retrieve documents
	// semantically close to the query. Only results with similarity
	// above the threshold (0.85) are considered.
	results, err := store.SimilaritySearch(ctx, query, 5, vectorstores.WithScoreThreshold(0.85))
	if err != nil {
		log.Fatal(err)
	}

	// Concatenate the retrieved content into a single string
	// to provide contextual information to the LLM.
	providedInformation := ""
	for _, result := range results {
		providedInformation += result.PageContent + "\n"
	}

	// Construct a natural-language prompt for the LLM.
	// The model is instructed to answer clearly and concisely based on
	// the provided notes or indicate if no related information is found.
	prompt := fmt.Sprintf(`You are an assistant that answers questions based on provided notes.
	If the information is not related, reply: "No related information."
	Answer clearly and concisely in English.

	Question: %v
	Notes: %v`, query, providedInformation)

	// Generate an answer using the Gemini LLM.
	answer, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		prompt,
		llms.WithModel("gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Print the final answer.
	fmt.Println("Answer:", answer)
}
