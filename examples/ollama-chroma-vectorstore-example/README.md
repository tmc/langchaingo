# Chroma Vector Store Example with LangChain and Ollama

This example demonstrates how to use the Chroma vector store with LangChain and Ollama to perform similarity searches on a collection of city data. The program showcases various querying techniques, including basic similarity search, filtering, and score thresholding.

## What This Example Does

1. **Setup**: 
   - Initializes an Ollama language model (LLM) with the "llama2" model.
   - Creates an embedder using the Ollama LLM.
   - Sets up a Chroma vector store with custom configurations.

2. **Data Loading**:
   - Adds a collection of city documents to the vector store. Each document contains the city name, population, and area.

3. **Similarity Searches**:
   The example performs three different similarity searches:

   a. "Up to 5 Cities in Japan":
      - Searches for Japanese cities with a score threshold of 0.8.
      - Limits the results to a maximum of 5 cities.

   b. "A City in South America":
      - Looks for a South American city with a score threshold of 0.8.
      - Returns only one result.

   c. "Large Cities in South America":
      - Searches for South American cities with specific filters:
        - Area greater than or equal to 1000
        - Population greater than or equal to 13 million

4. **Results Display**:
   - Prints the results of each search query, showing the matching city names.

## Key Features

- Demonstrates the use of Chroma vector store for similarity searches.
- Shows how to use Ollama for embeddings and as an LLM.
- Illustrates different querying techniques:
  - Basic similarity search
  - Score thresholding
  - Filtering based on metadata

This example is perfect for developers looking to understand how to implement and use vector stores for semantic search applications, especially when working with geographical data.
