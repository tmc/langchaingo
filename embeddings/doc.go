/*
Package embeddings contains the implementation for creating vector
embeddings from text using different APIs, like OpenAI and Google PaLM (VertexAI).

The main components of this package are:

- Embedder interface: a common interface for creating vector embeddings from texts.
- OpenAI: an Embedder implementation using the OpenAI API.
- VertexAIPaLM: an Embedder implementation using Google PaLM (VertexAI) API.
- Helper functions: utility functions for embedding, such as `batchTexts` and `maybeRemoveNewLines`.

The package provides a flexible way to handle different APIs for generating
embeddings by using the Embedder interface as an abstraction.
*/
package embeddings
