package llms

import "errors"

var (
	// ErrEmptyResponse is thrown when an LLM returns an empty response
	ErrEmptyResponse = errors.New("no response")
	// ErrIncompleteEmbedding is thrown when the length of an embedding
	// request does not match the length of the returned embeddings array
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
)
