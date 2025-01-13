# MongoVector: MongoDB Vector Store for Embeddings

`mongovector` provide a way for users to read and write to a [MongoDB Atlas Database](https://www.mongodb.com/products/platform/atlas-database) as a vector store using the [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) and a supported embedding service.

## Project Goals
The goal of this project is to enable users to interact with an Atlas cluster as a vector database. The `mongovector` package is designed to meet the following requirements:
- **Embedding-Agnostic**: The package allows users to embed data using various services, including OpenAI, Ollama, Mistral, and others.
- **VectorStore Interface Implementation**: The package implements the `VectorStore` interface, providing methods to add documents and perform similarity searches.

## Features

- **Document Storage**: Easily add documents to the MongoDB vector store with their embeddings.
- **Similarity Search**: Perform similarity searches based on user-defined queries and retrieve relevant documents.
- **Customizable Options**: Configure various options for embedding and searching, including score thresholds and filters.

## Installation

To use the `mongovector` package, ensure you have Go installed on your machine. You can then install the package using the following command:

```bash
go get github.com/tmc/langchaingo/vectorstores/mongovector@v0.1.13-pre.0
```

## Docker 

You can also use the following Docker image to containerize a free (M0) tier: [MongoDB Atlas Local](https://hub.docker.com/r/mongodb/mongodb-atlas-local).
