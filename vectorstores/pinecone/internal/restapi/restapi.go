package restapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrMissingTextKey = errors.New("missing text key in vector metadata")
	ErrEmptyResponse  = errors.New("empty response")
)

type apiError struct {
	task    string
	message string
}

func newAPIError(task string, body io.ReadCloser) apiError {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, body)
	if err != nil {
		return apiError{task: "reading body of error message", message: err.Error()}
	}

	return apiError{task: task, message: buf.String()}
}

func (e apiError) Error() string {
	return fmt.Sprintf("%s: %s", e.task, e.message)
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

func Upsert(
	ctx context.Context,
	vectors [][]float64,
	metadatas []map[string]any,
	apiKey string,
	nameSpace string,
	indexName string,
	projectName string,
	environment string,
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
		getEndpoint(indexName, projectName, environment)+"/vectors/upsert",
		apiKey,
		"POST",
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
}

func Query(
	ctx context.Context,
	vector []float64,
	numVectors int,
	apiKey,
	textKey,
	nameSpace string,
	indexName string,
	projectName string,
	environment string,
) ([]schema.Document, error) {
	payload := queryPayload{
		IncludeValues:   true,
		IncludeMetadata: true,
		Vector:          vector,
		TopK:            numVectors,
		Namespace:       nameSpace,
	}

	body, statusCode, err := doRequest(
		ctx,
		payload,
		getEndpoint(indexName, projectName, environment)+"/query",
		apiKey,
		"POST",
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
		pageContent, ok := match.Metadata[textKey].(string)
		if !ok {
			return nil, ErrMissingTextKey
		}
		delete(match.Metadata, textKey)

		docs = append(docs, schema.Document{
			PageContent: pageContent,
			Metadata:    match.Metadata,
		})
	}

	return docs, nil
}

type whoamiResponse struct {
	ProjectName string `json:"project_name"`
	UserLabel   string `json:"user_label"`
	UserName    string `json:"user_name"`
}

// Whoami returns the project name associated with the api key.
func Whoami(ctx context.Context, environment, apiKey string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx, "GET",
		fmt.Sprintf("https://controller.%s.pinecone.io/actions/whoami", environment),
		nil,
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Api-Key", apiKey)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	var response whoamiResponse

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&response)
	return response.ProjectName, err
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
