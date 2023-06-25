package pinecone

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

// APIError is an error type returned if the status code from the rest
// api is not 200.
type APIError struct {
	Task    string
	Message string
}

func newAPIError(task string, body io.ReadCloser) APIError {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, body)
	if err != nil {
		return APIError{Task: "reading body of error message", Message: err.Error()}
	}

	return APIError{Task: task, Message: buf.String()}
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Task, e.Message)
}

type vector struct {
	Values   []float64      `json:"values"`
	Metadata map[string]any `json:"metadata"`
	ID       string         `json:"id"`
}

type upsertPayload struct {
	Vectors   []vector `json:"vectors"`
	Namespace string   `json:"namespace"`
}

func (s Store) restUpsert(
	ctx context.Context,
	vectors [][]float64,
	metadatas []map[string]any,
	nameSpace string,
) error {
	v := make([]vector, 0, len(vectors))
	for i := 0; i < len(vectors); i++ {
		v = append(v, vector{
			Values:   vectors[i],
			Metadata: metadatas[i],
			ID:       uuid.New().String(),
		})
	}

	payload := upsertPayload{
		Vectors:   v,
		Namespace: nameSpace,
	}

	body, status, err := doRequest(
		ctx,
		payload,
		getEndpoint(s.indexName, s.projectName, s.environment)+"/vectors/upsert",
		s.apiKey,
		http.MethodPost,
	)
	if err != nil {
		return err
	}
	defer body.Close()

	if status == http.StatusOK {
		return nil
	}

	return newAPIError("upserting vectors", body)
}

type sparseValues struct {
	Indices []int     `json:"indices"`
	Values  []float64 `json:"values"`
}

type match struct {
	ID           string         `json:"id"`
	Score        float64        `json:"score"`
	Values       []float64      `json:"values"`
	SparseValues sparseValues   `json:"sparseValues"`
	Metadata     map[string]any `json:"metadata"`
}

type queriesResponse struct {
	Matches   []match `json:"matches"`
	Namespace string  `json:"namespace"`
}

type queryPayload struct {
	IncludeValues   bool      `json:"includeValues"`
	IncludeMetadata bool      `json:"includeMetadata"`
	Vector          []float64 `json:"vector"`
	TopK            int       `json:"topK"`
	Namespace       string    `json:"namespace"`
	Filter          any       `json:"filter"`
}

func (s Store) restQuery(
	ctx context.Context,
	vector []float64,
	numVectors int,
	nameSpace string,
	scoreThreshold float64,
	filter any,
) ([]schema.Document, error) {
	payload := queryPayload{
		IncludeValues:   true,
		IncludeMetadata: true,
		Vector:          vector,
		TopK:            numVectors,
		Namespace:       nameSpace,
		Filter:          filter,
	}

	body, statusCode, err := doRequest(
		ctx,
		payload,
		getEndpoint(s.indexName, s.projectName, s.environment)+"/query",
		s.apiKey,
		http.MethodPost,
	)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if statusCode != http.StatusOK {
		return nil, newAPIError("querying index", body)
	}

	var response queriesResponse

	decoder := json.NewDecoder(body)
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	if len(response.Matches) == 0 {
		return nil, ErrEmptyResponse
	}

	docs := make([]schema.Document, 0, len(response.Matches))
	for _, match := range response.Matches {
		pageContent, ok := match.Metadata[s.textKey].(string)
		if !ok {
			return nil, ErrMissingTextKey
		}
		delete(match.Metadata, s.textKey)

		doc := schema.Document{
			PageContent: pageContent,
			Metadata:    match.Metadata,
		}

		// If scoreThreshold is not 0, we only return matches with a score above the threshold.
		if scoreThreshold != 0 && match.Score >= scoreThreshold {
			docs = append(docs, doc)
		} else if scoreThreshold == 0 { // If scoreThreshold is 0, we return all matches.
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

func doRequest(ctx context.Context, payload any, url, apiKey, method string) (io.ReadCloser, int, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
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
	return r.Body, r.StatusCode, err
}

func getEndpoint(index, project, environment string) string {
	urlString := url.QueryEscape(
		fmt.Sprintf(
			"%s-%s.svc.%s.pinecone.io",
			index,
			project,
			environment,
		),
	)
	return "https://" + urlString
}
