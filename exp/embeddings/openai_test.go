package embeddings_test

import (
	"testing"

	"github.com/tmc/langchaingo/embeddings"
)

func TestOpenaiEmbeddings(t *testing.T) {
	e, err := embeddings.NewOpenAI()
	if err != nil {
		t.Errorf("Unexpected error creating openai embeddings struct: %e", err)
	}

	_, err = e.EmbedQuery("Hello world!")
	if err != nil {
		t.Errorf("Unexpected error embed query: %e", err)
	}

	embeddings, err := e.EmbedDocuments([]string{"Hello world", "The world is ending", "bye bye"})
	if err != nil {
		t.Errorf("Unexpected error embed document: %e", err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Unexpected number of embeddings. Got: %v. Expect 3", len(embeddings))
	}
}
