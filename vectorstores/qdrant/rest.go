package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
)

// upsertPoints updates or inserts points into the Qdrant collection.
func (s Store) upsertPoints(
	ctx context.Context,
	baseURL *url.URL,
	vectors [][]float32,
	payloads []map[string]interface{},
) ([]string, error) {
	ids := make([]string, len(vectors))
	for i := range ids {
		ids[i] = uuid.NewString()
	}

	payload := upsertBody{
		Batch: upsertBatch{
			IDs:      ids,
			Vectors:  vectors,
			Payloads: payloads,
		},
	}

	url := baseURL.JoinPath("collections", s.collectionName, "points")
	body,
		status,
		err := DoRequest(
		ctx, *url,
		s.apiKey,
		http.MethodPut,
		payload,
	)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if status == http.StatusOK {
		return ids, nil
	}

	return nil,
		newAPIError("upserting vectors", body)
}

// searchPoints queries the Qdrant collection for points based on the provided parameters.
func (s Store) searchPoints(
	ctx context.Context,
	baseURL *url.URL,
	vector []float32,
	numVectors int,
	scoreThreshold float32,
	filter any,
) ([]schema.Document, error) {
	payload := searchBody{
		WithPayload: true,
		WithVector:  false,
		Vector:      vector,
		Limit:       numVectors,
		Filter:      filter,
	}

	if scoreThreshold != 0 {
		payload.ScoreThreshold = scoreThreshold
	}

	url := baseURL.JoinPath("collections", s.collectionName, "points", "search")
	body,
		statusCode,
		err := DoRequest(
		ctx, *url,
		s.apiKey,
		http.MethodPost,
		payload,
	)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode != http.StatusOK {
		return nil, newAPIError("querying collection", body)
	}

	var response searchResponse

	decoder := json.NewDecoder(body)
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	docs := make([]schema.Document, len(response.Result))
	for i, match := range response.Result {
		pageContent, ok := match.Payload[s.contentKey].(string)
		if !ok {
			return nil, fmt.Errorf("payload does not contain content key '%s'", s.contentKey)
		}
		delete(match.Payload, s.contentKey)

		doc := schema.Document{
			PageContent: pageContent,
			Metadata:    match.Payload,
			Score:       match.Score,
		}

		docs[i] = doc
	}

	return docs, nil
}

// doRequest performs an HTTP request to the Qdrant API.
func DoRequest(ctx context.Context,
	url url.URL,
	apiKey,
	method string,
	payload interface{},
) (io.ReadCloser, int, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, method, url.String()+"?wait=true", body)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-Key", apiKey)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	return r.Body, r.StatusCode, err
}

// newAPIError creates an error based on the Qdrant API response.
func newAPIError(task string, body io.ReadCloser) error {
	buf := new(bytes.Buffer)
	_,
		err := io.Copy(buf, body)
	if err != nil {
		return fmt.Errorf("failed to read body of error message: %w", err)
	}

	return fmt.Errorf("%s: %s", task, buf.String())
}
