# pgvector store with OpenAI embeddings

Start Postgres:

```shell
docker-compose up -d
```

`postgres.Dockerfile` extends the official Postgres image and installs the pgvector extension.

`create_extension.sql` enables the pgvector extension and should run automatically when the container starts for the
first time.

Run the example:

```shell
export OPENAI_API_KEY=<your key>
go run pgvector_vectorstore_example.go
```

# PGVector Store with OpenAI Embeddings Example

This example demonstrates how to use pgvector, a PostgreSQL extension for vector similarity search, with OpenAI embeddings in a Go application. It showcases the integration of langchain-go, OpenAI's API, and pgvector to create a powerful vector database for similarity searches.

## What This Example Does

1. **Sets up a PostgreSQL Database with pgvector:**
   - Uses Docker to run a PostgreSQL instance with the pgvector extension installed.
   - Automatically creates and enables the vector extension when the container starts.

2. **Initializes OpenAI Embeddings:**
   - Creates an embeddings client using the OpenAI API.
   - Requires an OpenAI API key to be set as an environment variable.

3. **Creates a PGVector Store:**
   - Establishes a connection to the PostgreSQL database.
   - Initializes a vector store using pgvector and OpenAI embeddings.

4. **Adds Sample Documents:**
   - Inserts several documents (cities) with metadata into the vector store.
   - Each document includes the city name, population, and area.

5. **Performs Similarity Searches:**
   - Demonstrates various types of similarity searches:
     a. Basic search for documents similar to "japan".
     b. Search for South American cities with a score threshold.
     c. Search with both score threshold and metadata filtering.

## How to Run the Example

1. Start the PostgreSQL database:
   ```
   docker-compose up -d
   ```

2. Set your OpenAI API key:
   ```
   export OPENAI_API_KEY=<your key>
   ```

3. Run the Go example:
   ```
   go run pgvector_vectorstore_example.go
   ```

## Key Features

- Integration of pgvector with OpenAI embeddings
- Similarity search with score thresholds
- Metadata filtering in vector searches
- Dockerized PostgreSQL setup for easy deployment

This example provides a practical demonstration of using vector databases for semantic search and similarity matching, which can be incredibly useful for various AI and machine learning applications.
