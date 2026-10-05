---
title: "Getting Started: Document Loaders"
draft: true
---

# Getting Started: Document Loaders

<Info>
[Conceptual Guide](https://python.langchain.com/docs/modules/data_connection/document_loaders)
</Info>

Document loaders make it easy to create documents from a variety of sources. These documents can then be loaded onto [Vector Stores](../vector_stores/index.md) to load documents from a source.

## All Document Loaders

## Advanced

If you want to implement your own Document Loader, you must create a struct that implement the document loader interface.

```go
// Loader is the interface for loading and splitting documents from a source.
type Loader interface {
	// Loads loads from a source and returns documents.
	Load(context.Context) ([]schema.Document, error)
	// LoadAndSplit loads from a source and splits the documents using a text splitter.
	LoadAndSplit(context.Context, textsplitter.TextSplitter) ([]schema.Document, error)
}
```
