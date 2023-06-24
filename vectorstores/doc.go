/*
Package vectorstores contains the implementation of VectorStore, an interface for saving
and querying documents as vector embeddings.

The main components of this package are:

- VectorStore interface: a common interface for saving and querying vector embeddings of documents.
- Options: a set of options for similarity search and document addition.
- Retriever: a retriever for vector stores that implements the schema.Retriever interface.

The package provides a flexible way to handle different types of vector stores
by using the VectorStore interface as an abstraction.
It supports customization of the search and storage operation via the Options mechanism.
*/
package vectorstores
