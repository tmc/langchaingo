# Cybertron Embedding Example

Hello there! 👋 This example demonstrates how to use the Cybertron embedding model with LangChain in Go. It's a fun and practical way to explore document embeddings and similarity searches. Let's break down what this example does!

## What Does This Example Do?

This example showcases two main features:

1. In-memory document similarity comparison
2. Vector store integration with Weaviate

### In-Memory Document Similarity

The `exampleInMemory` function does the following:

- Creates embeddings for three words: "tokyo", "japan", and "potato"
- Calculates the cosine similarity between each pair of words
- Prints out the similarity scores

This helps you understand how semantically related different words are in the embedding space.

### Weaviate Vector Store Integration

The `exampleWeaviate` function demonstrates how to use the Cybertron embeddings with a Weaviate vector store:

- Creates a Weaviate vector store using the Cybertron embedder
- Adds three documents to the store: "tokyo", "japan", and "potato"
- Performs a similarity search for the query "japan"
- Prints out the matching results and their similarity scores

This shows how you can use embeddings for more advanced document retrieval tasks.

## Key Components

1. **Cybertron Embedder**: The example uses the "BAAI/bge-small-en-v1.5" model to generate embeddings. This model is automatically downloaded and cached.

2. **Cosine Similarity**: A custom function is implemented to calculate the similarity between embeddings.

3. **Weaviate Integration**: The example shows how to set up and use a Weaviate vector store with the Cybertron embeddings.

## How to Run

To run this example:

1. Ensure you have Go installed on your system.
2. Set up the required environment variables for Weaviate (if you want to run the Weaviate example):
   - `WEAVIATE_SCHEME`
   - `WEAVIATE_HOST`
3. Run the example using `go run cybertron-embedding.go`

## Note

The Cybertron model runs locally on your CPU, so larger models might be slow. The example uses a smaller model for better performance.

Have fun exploring embeddings and semantic similarity with this example! 🚀🔍
