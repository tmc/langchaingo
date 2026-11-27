package vectorstores

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/schema"
)

type stubVectorStore struct {
	docs []schema.Document
	err  error
}

func (s stubVectorStore) AddDocuments(context.Context, []schema.Document, ...Option) ([]string, error) {
	return nil, nil
}

func (s stubVectorStore) SimilaritySearch(context.Context, string, int, ...Option) ([]schema.Document, error) {
	return s.docs, s.err
}

// recordingHandler implements callbacks.Handler and the optional
// callbacks.RetrieverErrorHandler.
type recordingHandler struct {
	callbacks.SimpleHandler
	started int
	ended   int
	errs    []error
}

func (h *recordingHandler) HandleRetrieverStart(context.Context, string) {
	h.started++
}

func (h *recordingHandler) HandleRetrieverEnd(context.Context, string, []schema.Document) {
	h.ended++
}

func (h *recordingHandler) HandleRetrieverError(_ context.Context, err error) {
	h.errs = append(h.errs, err)
}

func TestRetrieverReportsFailureThroughErrorCallback(t *testing.T) {
	t.Parallel()

	searchErr := errors.New("vector store unavailable")
	handler := &recordingHandler{}
	retriever := ToRetriever(stubVectorStore{err: searchErr}, 3)
	retriever.CallbacksHandler = handler

	docs, err := retriever.GetRelevantDocuments(context.Background(), "query")

	require.ErrorIs(t, err, searchErr)
	assert.Nil(t, docs)
	assert.Equal(t, 1, handler.started)
	assert.Equal(t, 0, handler.ended)
	require.Len(t, handler.errs, 1)
	assert.ErrorIs(t, handler.errs[0], searchErr)
}

func TestRetrieverDoesNotReportErrorOnSuccess(t *testing.T) {
	t.Parallel()

	handler := &recordingHandler{}
	retriever := ToRetriever(stubVectorStore{docs: []schema.Document{{PageContent: "doc"}}}, 3)
	retriever.CallbacksHandler = handler

	docs, err := retriever.GetRelevantDocuments(context.Background(), "query")

	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, 1, handler.started)
	assert.Equal(t, 1, handler.ended)
	assert.Empty(t, handler.errs)
}

func TestRetrieverErrorCallbackIsOptional(t *testing.T) {
	t.Parallel()

	searchErr := errors.New("vector store unavailable")
	retriever := ToRetriever(stubVectorStore{err: searchErr}, 3)
	retriever.CallbacksHandler = callbacks.SimpleHandler{}

	docs, err := retriever.GetRelevantDocuments(context.Background(), "query")

	require.ErrorIs(t, err, searchErr)
	assert.Nil(t, docs)
}
