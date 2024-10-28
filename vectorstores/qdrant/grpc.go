package qdrant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	qc "github.com/qdrant/go-client/qdrant"
	"github.com/tmc/langchaingo/schema"
)

// upsertPoints updates or inserts points into the Qdrant collection.
func (s Store) upsertPoints(
	ctx context.Context,
	vectors [][]float32,
	payloads []map[string]interface{},
) ([]string, error) {
	ids := make([]string, len(vectors))
	for i := range ids {
		ids[i] = uuid.NewString()
	}

	points := make([]*qc.PointStruct, len(vectors))
	for i, vector := range vectors {
		points[i] = &qc.PointStruct{
			Id:      qc.NewID(ids[i]),
			Payload: qc.NewValueMap(payloads[i]),
			Vectors: qc.NewVectors(vector...),
		}
	}

	_, err := s.client.Upsert(ctx, &qc.UpsertPoints{
		CollectionName: s.collectionName,
		Points:         points,
		Wait:           qc.PtrOf(true),
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// searchPoints queries the Qdrant collection for points based on the provided parameters.
func (s Store) searchPoints(
	ctx context.Context,
	vector []float32,
	numVectors int,
	scoreThreshold float32,
	filter any,
) ([]schema.Document, error) {
	var filterTyped *qc.Filter
	if filter != nil {
		f, ok := filter.(*qc.Filter)
		if !ok {
			return nil, fmt.Errorf("filter must be of type *qc.Filter, got %T", filter)
		}
		filterTyped = f
	}

	query := qc.QueryPoints{
		CollectionName: s.collectionName,
		WithPayload:    qc.NewWithPayload(true),
		Limit:          qc.PtrOf(uint64(numVectors)), //nolint:gosec
		Query:          qc.NewQuery(vector...),
		Filter:         filterTyped,
		ScoreThreshold: qc.PtrOf(scoreThreshold),
	}

	results, err := s.client.Query(ctx, &query)
	if err != nil {
		return nil, err
	}

	docs := make([]schema.Document, len(results))
	for i, result := range results {
		payload := result.GetPayload()
		pageContent := payload[s.contentKey].GetStringValue()

		delete(payload, s.contentKey)

		anyMap, err := NewAnyMap(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to convert payload to map[string]interface{}: %w", err)
		}
		doc := schema.Document{
			PageContent: pageContent,
			Metadata:    anyMap,
			Score:       result.GetScore(),
		}

		docs[i] = doc
	}

	return docs, nil
}
