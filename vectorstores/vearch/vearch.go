package vearch

import (
	"context"
	"net/url"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type Store struct {
	DbName     string
	SpaceName  string
	ClusterUrl url.URL
	embedder   embeddings.Embedder
}

_ vectorstores.VectorStore = Store{}

func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	return s, nil
}

func (s *Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
	options ...vectorstores.Option,
) ([]string, error) {
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}

	vectors,
		err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

	if len(vectors) != len(docs) {
		return nil, errors.New("number of vectors from embedder does not match number of documents")
	}

	metadatas := make([]map[string]interface{}, 0, len(docs))
	for i := 0; i < len(docs); i++ {
		metadata := make(map[string]interface{}, len(docs[i].Metadata))
		for key, value := range docs[i].Metadata {
			metadata[key] = value
		}
		metadata[s.contentKey] = texts[i]

		metadatas = append(metadatas, metadata)
	}
	return s.upsertDocuments(ctx, &s.ClusterUrl, vectors, metadatas)
}

func (s *Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)

	filters := s.getFilters(opts)

	scoreThreshold,
		err := s.getScoreThreshold(opts)
	if err != nil {
		return nil, err
	}

	vector,
		err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.searchDocuments(ctx, &s.ClusterUrl, vector, numDocuments, scoreThreshold, filters)
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getScoreThreshold(opts vectorstores.Options) (float32, error) {
	if opts.ScoreThreshold < 0 || opts.ScoreThreshold > 1 {
		return 0, ErrInvalidScoreThreshold
	}
	return opts.ScoreThreshold, nil
}

func (s Store) getFilters(opts vectorstores.Options) any {
	if opts.Filters != nil {
		return opts.Filters
	}
	return nil
}