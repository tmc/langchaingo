# Using MongoDB Atlas as a Vector Store with OpenAI Embeddings

This project illustrates how to leverage MongoDB as a vector store for performing similarity searches, utilizing OpenAI embeddings within a Go application. It integrates the LangChainGo library, OpenAI's API, and MongoDB to create an efficient vector database for semantic search.


For more information on getting started with MongoDB Atlas, visit the [MongoDB Atlas Getting Started Guide](https://www.mongodb.com/products/platform/atlas-database). You can also use the following Docker image to containerize a free (M0) tier: [MongoDB Atlas Local](https://hub.docker.com/r/mongodb/mongodb-atlas-local).

## What This Tutorial Covers

1. **MongoDB Setup:**
   - Connects to a MongoDB Atlas instance using a specified connection string.
   - Automatically checks for and creates a vector search index on the collection if it is not already present, ensuring compatibility with OpenAI's embedding model.

2. **OpenAI Embeddings Initialization:**
   - Establishes an embeddings client through the OpenAI API.
   - Requires the OpenAI API key to be set as an environment variable for authentication.

3. **Creating the Vector Store:**
   - Connects to the MongoDB database and sets up a vector store that utilizes OpenAI embeddings for document representation.

4. **Inserting Sample Data:**
   - Adds a collection of documents (cities) along with their metadata into the vector store.
   - Each document contains information such as the city name, population, and area.

5. **Executing Similarity Searches:**
   - Demonstrates various types of similarity searches.

## Running the Example

1. Configure your environment by setting the MongoDB URI and OpenAI API key:
   ```bash
   export MONGODB_URI=<your_mongodb_uri>
   export OPENAI_API_KEY=<your_openai_api_key>
  
2. If you want to run this using docker-compose.yml, `MONGODB_URI` should be `mongodb://localhost:27017/?directConnection=true`: `docker-compose up -d`

3. Run the program: `go run mongovector_vectorstore_example.go`
