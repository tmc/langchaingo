# Google Cloud SQL Vector Store Example

This example demonstrates how to use [Cloud SQL for Postgres](https://cloud.google.com/products/sql) for vector similarity search with LangChain in Go.

## What This Example Does

1. **Creates a Cloud SQL VectorStore:**
   - Initializes the `cloudsql.PostgresEngine` object to establish a connection to the Cloud SQL database.
   - Initializes a new table to store embeddings.
   - Initializes a `cloudsql.VectorStore` object using a VertexAI model for embeddings.

2. **Initializes VertexAI Embeddings:**
    - Creates an embeddings client using the VertexAI API.

3. **Adds Sample Documents:**
    - Inserts several documents (cities) with metadata into the vector store.
    - Each document includes the city name, population, and area.

4. **Performs Similarity Searches:**
    - Basic search for documents similar to "Japan".
    - Customized search for documents using filters by metadata.

## How to Run the Example

1. Set the following environment variables:
   ```
   export PROJECT_ID=<your project Id>
   export GOOGLE_CLOUD_LOCATION=<your cloud location>
   export POSTGRES_USERNAME=<your user>
   export POSTGRES_PASSWORD=<your password>
   export POSTGRES_REGION=<your region>
   export POSTGRES_INSTANCE=<your instance>
   export POSTGRES_DATABASE=<your database>
   export POSTGRES_TABLE=<your tablename>
   ```

2. Run the Go example:
   ```
   go run google_cloudsql_vectorstore_example.go
   ```

## Key Features

- This example demonstrates how to use `cloudsql.PostgresEngine` for connection pooling.
- It shows how to integrate with VertexAI embeddings models.
- Run the code to add documents and perform a similarity search with `cloudsql.VectorStore`.
- Demonstrates how to filter through the metadata added by using key value pairs.

This example provides a practical demonstration of using vector databases for semantic search and similarity matching, which can be incredibly useful for various AI and machine learning applications.
