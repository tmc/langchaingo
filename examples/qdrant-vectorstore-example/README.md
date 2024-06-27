# Qdrant Vector Store Example with LangChain Go

Welcome to this cheerful example of using Qdrant vector store with LangChain Go! 🎉

This example demonstrates how to use the Qdrant vector store to store and search for similar documents using embeddings. It's a great way to get started with vector databases and semantic search in your Go applications!

## What This Example Does

1. **Sets up OpenAI Embeddings**: 
   - Creates an embeddings client using the OpenAI API.
   - Make sure you have your `OPENAI_API_KEY` environment variable set!

2. **Creates a Qdrant Vector Store**:
   - Connects to your Qdrant instance.
   - Don't forget to replace `YOUR_QDRANT_URL` and `YOUR_COLLECTION_NAME` with your actual Qdrant details!

3. **Adds Documents**:
   - Adds a variety of documents about different locations to the vector store.
   - Each document has some content and metadata (like area).

4. **Performs Similarity Searches**:
   - Searches for documents similar to "england".
   - Searches for "american places" with a score threshold.
   - Searches for "cities in south america" with both a score threshold and metadata filter.

## Cool Features Demonstrated

- **Similarity Search**: Find documents that are semantically similar to a query.
- **Score Threshold**: Filter results based on a minimum similarity score.
- **Metadata Filtering**: Use additional metadata to refine your search results.

## How to Run

1. Make sure you have Go installed and your `OPENAI_API_KEY` set.
2. Replace `YOUR_QDRANT_URL` and `YOUR_COLLECTION_NAME` with your Qdrant details.
3. Run the example with `go run qdrant_vectorstore_example.go`.

Have fun exploring the world of vector databases and semantic search with LangChain Go and Qdrant! 🚀🔍
