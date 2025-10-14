# Pinecone Vector Store Example

Welcome to this exciting example of using Pinecone as a vector store with LangChain in Go! 🚀

## What This Example Does

This example demonstrates how to use Pinecone, a powerful vector database, in conjunction with LangChain to create and query a vector store. Here's a breakdown of the main features:

1. **Setting up OpenAI Embeddings**: The example uses OpenAI's embedding model to convert text into vector representations.

2. **Creating a Pinecone Vector Store**: It shows how to initialize a Pinecone vector store with custom configurations.

3. **Adding Documents**: The code adds several documents (cities) to the vector store, each with its own metadata (population and area).

4. **Performing Similarity Searches**: The example showcases different types of similarity searches:
   - Basic similarity search
   - Search with a score threshold
   - Search with both a score threshold and metadata filters

## Key Points

- The example uses the `github.com/vendasta/langchaingo` library for LangChain functionality in Go.
- It demonstrates how to handle errors and set up the necessary clients and stores.
- The code shows how to use metadata filters to refine search results based on specific criteria.

## Running the Example

To run this example, make sure you have:

1. Set up your OpenAI API key as an environment variable (`OPENAI_API_KEY`).
2. Replaced `"YOUR_API_KEY"` with your actual Pinecone API key.

This example is a great starting point for anyone looking to implement vector search capabilities in their Go applications using Pinecone and LangChain! 🎉

Happy coding! 💻🌟
