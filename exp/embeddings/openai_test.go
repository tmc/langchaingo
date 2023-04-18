package embeddings

import (
	"os"
	"testing"
)

func TestOpenaiEmbeddings(t *testing.T) {
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	e, err := NewOpenAI()
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.EmbedQuery("Hello world!")
	if err != nil {
		t.Fatal(err)
	}

	embeddings, err := e.EmbedDocuments([]string{"Hello world", "The world is ending", "bye bye"})
	if err != nil {
		t.Fatal(err)
	}

	if len(embeddings) != 3 {
		t.Errorf("Unexpected number of embeddings. Got: %v. Expect 3", len(embeddings))
	}
}
