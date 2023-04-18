package pineconeClient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type SparseValues struct {
	Indices []int     `json:"indices"`
	Values  []float64 `json:"values"`
}

type Match struct {
	ID           string            `json:"id"`
	Score        float64           `json:"score"`
	Values       []float64         `json:"values"`
	SparseValues SparseValues      `json:"sparseValues"`
	Metadata     map[string]string `json:"metadata"`
}

type QueriesResponse struct {
	Matches   []Match `json:"matches"`
	Namespace string  `json:"namespace"`
}

type queryPayload struct {
	IncludeValues   bool      `json:"includeValues"`
	IncludeMetadata bool      `json:"includeMetadata"`
	Vector          []float64 `json:"vector"`
	TopK            int       `json:"topK"`
	Namespace       string    `json:"namespace"`
}

func (c Client) Query(ctx context.Context, vector []float64, numVectors int, nameSpace string) (QueriesResponse, error) {
	payload := queryPayload{
		IncludeValues:   true,
		IncludeMetadata: true,
		Vector:          vector,
		TopK:            numVectors,
		Namespace:       nameSpace,
	}

	body, statusCode, err := doRequest(ctx, payload, c.getEndpoint()+"/query", c.apiKey)
	if err != nil {
		return QueriesResponse{}, err
	}
	defer body.Close()

	if statusCode != 200 {
		return QueriesResponse{}, errorMessageFromErrorResponse("querying index", body)
	}

	var response QueriesResponse

	decoder := json.NewDecoder(body)
	err = decoder.Decode(&response)
	return response, err
}

func doRequest(ctx context.Context, payload any, url, apiKey string) (io.ReadCloser, int, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "text/plain")
	req.Header.Set("Api-Key", apiKey)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	return r.Body, r.StatusCode, nil
}
