# Chroma Vector Store Example

This example demonstrates how to use the Chroma vector store with LangChain in Go. It showcases various operations and queries on a vector store containing information about cities.

## What This Example Does

1. **Vector Store Creation**: The example starts by creating a new Chroma vector store using environment variables for configuration.

2. **Adding Documents**: It adds a list of documents to the vector store. Each document represents a city with its name, population, and area.

3. **Similarity Searches**: The example performs three different similarity searches:

   a. **Up to 5 Cities in Japan**: Searches for cities located in Japan, limiting the results to 5 and using a score threshold.
   
   b. **A City in South America**: Looks for a single city in South America, also using a score threshold.
   
   c. **Large Cities in South America**: Searches for large cities in South America, using filters for area and population.

4. **Result Display**: Finally, it prints out the results of each search, showing the matching cities for each query.

## Key Features

- Demonstrates the use of the Chroma vector store in Go
- Shows how to add documents with metadata to a vector store
- Illustrates different types of similarity searches with various options
- Showcases the use of filters in vector store queries
- Provides examples of working with environment variables for configuration

This example is excellent for developers looking to understand how to integrate and use vector stores in their Go applications, particularly for semantic search and similarity matching tasks.
