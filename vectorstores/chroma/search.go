package chroma

import (
	"context"
	"fmt"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/tmc/langchaingo/schema"
)

// SearchOption configures a Search call.
type SearchOption func(*searchConfig)

type searchConfig struct {
	scoreThreshold float32
	filter         map[string]any
	offset         int
	readLevel      chroma.ReadLevel
}

// WithSearchScoreThreshold filters results below the given relevance score (0-1, higher is better).
func WithSearchScoreThreshold(threshold float32) SearchOption {
	return func(c *searchConfig) {
		c.scoreThreshold = threshold
	}
}

// WithSearchFilter adds a metadata filter to the search using Chroma's where clause syntax.
func WithSearchFilter(filter map[string]any) SearchOption {
	return func(c *searchConfig) {
		c.filter = filter
	}
}

// WithSearchOffset sets the pagination offset for search results.
func WithSearchOffset(offset int) SearchOption {
	return func(c *searchConfig) {
		c.offset = offset
	}
}

// WithSearchReadLevel sets the read consistency level for the search.
func WithSearchReadLevel(level chroma.ReadLevel) SearchOption {
	return func(c *searchConfig) {
		c.readLevel = level
	}
}

// Search performs a KNN text search using Chroma's Search API, which returns relevance scores
// (higher is better) rather than distances. This provides richer functionality than SimilaritySearch
// including pagination, read levels, and composable ranking via the Collection() escape hatch.
// See https://docs.trychroma.com/cloud/search-api/overview for details.
func (s Store) Search(ctx context.Context, query string, numDocuments int, opts ...SearchOption) ([]schema.Document, error) {
	cfg := &searchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.scoreThreshold < 0 || cfg.scoreThreshold > 1 {
		return nil, ErrInvalidScoreThreshold
	}

	filter := s.getNamespacedSearchFilter(cfg.filter)

	searchReqOpts := []chroma.SearchRequestOption{
		chroma.WithKnnRank(chroma.KnnQueryText(query), chroma.WithKnnLimit(numDocuments)),
		chroma.WithLimit(numDocuments),
		chroma.WithSelect(chroma.KDocument, chroma.KMetadata, chroma.KScore, chroma.KID),
	}

	if cfg.offset > 0 {
		searchReqOpts = append(searchReqOpts, chroma.WithOffset(cfg.offset))
	}

	if filter != nil {
		searchReqOpts = append(searchReqOpts,
			chroma.WithSearchFilter(&chroma.SearchFilter{Where: &rawWhereClause{data: filter}}),
		)
	}

	collOpts := []chroma.SearchCollectionOption{
		chroma.NewSearchRequest(searchReqOpts...),
	}
	if cfg.readLevel != "" {
		collOpts = append(collOpts, chroma.WithReadLevel(cfg.readLevel))
	}

	result, err := s.collection.Search(ctx, collOpts...)
	if err != nil {
		return nil, err
	}

	searchResult, ok := result.(*chroma.SearchResultImpl)
	if !ok {
		return nil, fmt.Errorf("unexpected search result type: %T", result)
	}

	var docs []schema.Document
	for _, row := range searchResult.Rows() {
		score := float32(row.Score)
		if score >= cfg.scoreThreshold {
			docs = append(docs, schema.Document{
				PageContent: row.Document,
				Metadata:    documentMetadataToMap(row.Metadata),
				Score:       score,
			})
		}
	}

	return docs, nil
}
