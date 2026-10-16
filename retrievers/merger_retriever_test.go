package retrievers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

var _ schema.Retriever = &Fakeretriever{}

type Fakeretriever struct {
	Docs []schema.Document
}

func (f *Fakeretriever) GetRelevantDocuments(_ context.Context, _ string) ([]schema.Document, error) {
	return f.Docs, nil
}

func TestMergerRetriever(t *testing.T) { //nolint:funlen
	t.Parallel()
	ctx := context.Background()
	content1 := schema.Document{PageContent: "fake doc 1"}
	content2 := schema.Document{PageContent: "fake doc 2"}
	content3 := schema.Document{PageContent: "fake doc 3"}
	content4 := schema.Document{PageContent: "fake doc 4"}
	retriever1 := Fakeretriever{Docs: []schema.Document{content1, content2}}
	retriever2 := Fakeretriever{Docs: []schema.Document{content3, content4}}
	merger := NewMergerRetriever([]schema.Retriever{&retriever1, &retriever2})
	documents, err := merger.GetRelevantDocuments(ctx, "fake query")
	require.NoError(t, err)
	require.Len(t, documents, 4)

	assert.Equal(t, documents[0], content1)
	assert.Equal(t, documents[1], content2)
	assert.Equal(t, documents[2], content3)
	assert.Equal(t, documents[3], content4)
}
