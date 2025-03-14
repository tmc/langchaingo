# Google AlloyDB Vector Store Example

This example demonstrates how to use [AlloyDB for Postgres](https://cloud.google.com/products/alloydb) for vector similarity search with LangChain in Go.

## What This Example Does

1. **Creates a AlloyDB VectorStore:**
   - Initializes the `alloydb.PostgresEngine` object to establish a connection to the AlloyDB database.
   - Initializes a new table to store embeddings.
   - Initializes a `alloydb.VectorStore` object using a VertexAI model for embeddings.

2. **Initializes VertexAI Embeddings:**
    - Creates an embeddings client using the VertexAI API.

3. **Adds Sample Documents:**
    - Inserts several documents (cities) with metadata into the vector store.
    - Each document includes the city name, population, and area.

4. **Performs Similarity Searches:**
    - Basic search for documents similar to "Japan".
    - Customized search for documents using filters by metadata.

## How to Run the Example

1. Set the following environment variables. Your AlloyDB values can be found in the [Google Cloud Console](https://console.cloud.google.com/alloydb/clusters):
   ```
   export PROJECT_ID=<your project Id>
   export VERTEX_LOCATION=<your vertex location>
   export ALLOYDB_USERNAME=<your user>
   export ALLOYDB_PASSWORD=<your password>
   export ALLOYDB_REGION=<your region>
   export ALLOYDB_CLUSTER=<your cluster>
   export ALLOYDB_INSTANCE=<your instance>
   export ALLOYDB_DATABASE=<your database>
   export ALLOYDB_TABLE=<your tablename>
   ```

2. Run the Go example:
   ```
   go run alloydb_vectorstore_example.go
   ```

## Key Features

- This example demonstrates how to use `alloydb.PostgresEngine` for connection pooling.
- It shows how to integrate with VertexAI embeddings models.
- Run the code to add documents and perform a similarity search with `alloydb.VectorStore`.
- Demonstrates how to filter through the metadata added by using key value pairs.

This example provides a practical demonstration of using vector databases for semantic search and similarity matching, which can be incredibly useful for various AI and machine learning applications.
