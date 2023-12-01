/*
Package embeddings contains helpers for creating vector embeddings from text
using different providers.

The main components of this package are:

  - [Embedder] interface: a common interface for creating vector embeddings
    from texts, with optional batching.
  - [NewEmbedder] creates implementations of [Embedder] from provider LLM
    (or Chat) clients.

See the package example below.
*/
package embeddings
