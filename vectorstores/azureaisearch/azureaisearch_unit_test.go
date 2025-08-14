package azureaisearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/0xDezzy/langchaingo/embeddings"
	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/vectorstores"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEmbedder is a test embedder for unit testing.
type testEmbedder struct {
	embedDocumentsFunc func(ctx context.Context, texts []string) ([][]float32, error)
	embedQueryFunc     func(ctx context.Context, text string) ([]float32, error)
}

func (m *testEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedDocumentsFunc != nil {
		return m.embedDocumentsFunc(ctx, texts)
	}
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = []float32{float32(len(texts[i])), 0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

func (m *testEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if m.embedQueryFunc != nil {
		return m.embedQueryFunc(ctx, text)
	}
	return []float32{float32(len(text)), 0.1, 0.2, 0.3}, nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		envVars     map[string]string
		opts        []Option
		wantErr     bool
		errContains string
		validate    func(t *testing.T, s Store)
	}{
		{
			name: "success with all options",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net",
				"AZURE_AI_SEARCH_API_KEY":  "test-key",
			},
			opts: []Option{
				WithEmbedder(&testEmbedder{}),
				WithHTTPClient(&http.Client{}),
				WithAPIKey("override-key"),
			},
			validate: func(t *testing.T, s Store) {
				assert.Equal(t, "https://test.search.windows.net", s.azureAISearchEndpoint)
				// Environment variable takes precedence over option
				assert.Equal(t, "test-key", s.azureAISearchAPIKey)
				assert.NotNil(t, s.embedder)
				assert.NotNil(t, s.client)
			},
		},
		{
			name:        "missing endpoint",
			opts:        []Option{WithEmbedder(&testEmbedder{})},
			wantErr:     true,
			errContains: "missing azureAISearchEndpoint",
		},
		{
			name: "missing embedder",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net",
			},
			wantErr:     true,
			errContains: "missing embedder",
		},
		{
			name: "endpoint trailing slash removed",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net/",
			},
			opts: []Option{WithEmbedder(&testEmbedder{})},
			validate: func(t *testing.T, s Store) {
				assert.Equal(t, "https://test.search.windows.net", s.azureAISearchEndpoint)
			},
		},
		{
			name: "api key from environment",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net",
				"AZURE_AI_SEARCH_API_KEY":  "env-api-key",
			},
			opts: []Option{WithEmbedder(&testEmbedder{})},
			validate: func(t *testing.T, s Store) {
				assert.Equal(t, "env-api-key", s.azureAISearchAPIKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			store, err := New(tt.opts...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, store)
			}
		})
	}
}

func TestStore_AddDocuments(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	tests := []struct {
		name           string
		docs           []schema.Document
		embedder       *testEmbedder
		mockServer     func() *httptest.Server
		options        []vectorstores.Option
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, ids []string)
	}{
		{
			name: "successful add documents",
			docs: []schema.Document{
				{
					PageContent: "Test document 1",
					Metadata:    map[string]any{"key": "value1"},
				},
				{
					PageContent: "Test document 2",
					Metadata:    map[string]any{"key": "value2"},
				},
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "POST", r.Method)
					assert.Contains(t, r.URL.Path, "/indexes/")
					assert.Contains(t, r.URL.Path, "/docs/index")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{"value":[]}`)
				}))
			},
			validateResult: func(t *testing.T, ids []string) {
				assert.Len(t, ids, 2)
				for _, id := range ids {
					assert.NotEmpty(t, id)
					_, err := uuid.Parse(id)
					assert.NoError(t, err)
				}
			},
		},
		{
			name: "embedder returns wrong number of vectors",
			docs: []schema.Document{
				{PageContent: "Test document"},
			},
			embedder: &testEmbedder{
				embedDocumentsFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
					return [][]float32{}, nil // Return empty vectors
				},
			},
			wantErr:     true,
			errContains: "number of vectors from embedder does not match number of documents",
		},
		{
			name: "embedder error",
			docs: []schema.Document{
				{PageContent: "Test document"},
			},
			embedder: &testEmbedder{
				embedDocumentsFunc: func(ctx context.Context, texts []string) ([][]float32, error) {
					return nil, errors.New("embedding failed")
				},
			},
			wantErr:     true,
			errContains: "embedding failed",
		},
		{
			name: "upload document error",
			docs: []schema.Document{
				{PageContent: "Test document"},
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, `{"error":"upload failed"}`)
				}))
			},
			wantErr:     true,
			errContains: "error returned from",
		},
		{
			name: "with namespace option",
			docs: []schema.Document{
				{PageContent: "Test document"},
			},
			options: []vectorstores.Option{
				vectorstores.WithNameSpace("test-namespace"),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.URL.Path, "/indexes/test-namespace/")
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{"value":[]}`)
				}))
			},
			validateResult: func(t *testing.T, ids []string) {
				assert.Len(t, ids, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			if tt.mockServer != nil {
				server := tt.mockServer()
				defer server.Close()
				serverURL = server.URL
			} else {
				serverURL = "https://test.search.windows.net"
			}

			embedder := tt.embedder
			if embedder == nil {
				embedder = &testEmbedder{}
			}

			store := Store{
				azureAISearchEndpoint: serverURL,
				azureAISearchAPIKey:   "test-key",
				embedder:              embedder,
				client:                httputil.DefaultClient,
			}

			ctx := context.Background()
			ids, err := store.AddDocuments(ctx, tt.docs, tt.options...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, ids)
			}
		})
	}
}

func TestStore_SimilaritySearch(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	tests := []struct {
		name           string
		query          string
		numDocuments   int
		options        []vectorstores.Option
		embedder       *testEmbedder
		mockServer     func() *httptest.Server
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, docs []schema.Document)
	}{
		{
			name:         "successful search",
			query:        "test query",
			numDocuments: 2,
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "POST", r.Method)
					assert.Contains(t, r.URL.Path, "/docs/search")

					var req SearchDocumentsRequestInput
					err := json.NewDecoder(r.Body).Decode(&req)
					assert.NoError(t, err)
					assert.Len(t, req.Vectors, 1)
					assert.Equal(t, "contentVector", req.Vectors[0].Fields)
					assert.Equal(t, 2, req.Vectors[0].K)

					response := SearchDocumentsRequestOuput{
						Value: []map[string]interface{}{
							{
								"@search.score": 0.95,
								"content":       "Result 1",
								"metadata":      `{"key":"value1"}`,
							},
							{
								"@search.score": 0.85,
								"content":       "Result 2",
								"metadata":      `{"key":"value2"}`,
							},
						},
					}
					err = json.NewEncoder(w).Encode(response)
					assert.NoError(t, err)
				}))
			},
			validateResult: func(t *testing.T, docs []schema.Document) {
				assert.Len(t, docs, 2)
				assert.Equal(t, "Result 1", docs[0].PageContent)
				assert.Equal(t, float32(0.95), docs[0].Score)
				assert.Equal(t, "value1", docs[0].Metadata["key"])
			},
		},
		{
			name:         "with score threshold",
			query:        "test query",
			numDocuments: 3,
			options: []vectorstores.Option{
				vectorstores.WithScoreThreshold(0.9),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := SearchDocumentsRequestOuput{
						Value: []map[string]interface{}{
							{
								"@search.score": 0.95,
								"content":       "High score",
								"metadata":      `{}`,
							},
							{
								"@search.score": 0.85,
								"content":       "Low score",
								"metadata":      `{}`,
							},
						},
					}
					err := json.NewEncoder(w).Encode(response)
					assert.NoError(t, err)
				}))
			},
			validateResult: func(t *testing.T, docs []schema.Document) {
				assert.Len(t, docs, 1)
				assert.Equal(t, "High score", docs[0].PageContent)
			},
		},
		{
			name:         "with filter",
			query:        "test query",
			numDocuments: 1,
			options: []vectorstores.Option{
				WithFilters("category eq 'technology'"),
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var req SearchDocumentsRequestInput
					err := json.NewDecoder(r.Body).Decode(&req)
					assert.NoError(t, err)
					assert.Equal(t, "category eq 'technology'", req.Filter)

					response := SearchDocumentsRequestOuput{
						Value: []map[string]interface{}{
							{
								"@search.score": 0.9,
								"content":       "Filtered result",
								"metadata":      `{}`,
							},
						},
					}
					err = json.NewEncoder(w).Encode(response)
					assert.NoError(t, err)
				}))
			},
			validateResult: func(t *testing.T, docs []schema.Document) {
				assert.Len(t, docs, 1)
			},
		},
		{
			name:  "embedder error",
			query: "test query",
			embedder: &testEmbedder{
				embedQueryFunc: func(ctx context.Context, text string) ([]float32, error) {
					return nil, errors.New("embed query failed")
				},
			},
			wantErr:     true,
			errContains: "embed query failed",
		},
		{
			name:         "search error",
			query:        "test query",
			numDocuments: 1,
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintln(w, `{"error":"search failed"}`)
				}))
			},
			wantErr:     true,
			errContains: "error returned from",
		},
		{
			name:         "invalid search results",
			query:        "test query",
			numDocuments: 1,
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					response := SearchDocumentsRequestOuput{
						Value: []map[string]interface{}{
							{
								// Missing @search.score
								"content":  "Result",
								"metadata": `{}`,
							},
						},
					}
					err := json.NewEncoder(w).Encode(response)
					assert.NoError(t, err)
				}))
			},
			wantErr:     true,
			errContains: "couldn't assert @search.score to float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			if tt.mockServer != nil {
				server := tt.mockServer()
				defer server.Close()
				serverURL = server.URL
			} else {
				serverURL = "https://test.search.windows.net"
			}

			embedder := tt.embedder
			if embedder == nil {
				embedder = &testEmbedder{}
			}

			store := Store{
				azureAISearchEndpoint: serverURL,
				azureAISearchAPIKey:   "test-key",
				embedder:              embedder,
				client:                httputil.DefaultClient,
			}

			ctx := context.Background()
			docs, err := store.SimilaritySearch(ctx, tt.query, tt.numDocuments, tt.options...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateResult != nil {
				tt.validateResult(t, docs)
			}
		})
	}
}

func TestStore_CreateIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		indexName   string
		options     []IndexOption
		mockServer  func() *httptest.Server
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful create index",
			indexName: "test-index",
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "PUT", r.Method)
					assert.Contains(t, r.URL.Path, "/indexes/test-index")
					assert.Equal(t, "test-key", r.Header.Get("api-key"))
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

					var body map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&body)
					assert.NoError(t, err)
					assert.Equal(t, "test-index", body["name"])

					w.WriteHeader(http.StatusCreated)
					fmt.Fprintln(w, `{"name":"test-index"}`)
				}))
			},
		},
		{
			name:      "server error",
			indexName: "test-index",
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusConflict)
					fmt.Fprintln(w, `{"error":"index already exists"}`)
				}))
			},
			wantErr:     true,
			errContains: "error returned from",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.mockServer()
			defer server.Close()

			store := Store{
				azureAISearchEndpoint: server.URL,
				azureAISearchAPIKey:   "test-key",
				client:                httputil.DefaultClient,
			}

			ctx := context.Background()
			err := store.CreateIndex(ctx, tt.indexName, tt.options...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestStore_DeleteIndex(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/indexes/test-index")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	store := Store{
		azureAISearchEndpoint: server.URL,
		azureAISearchAPIKey:   "test-key",
		client:                httputil.DefaultClient,
	}

	ctx := context.Background()
	err := store.DeleteIndex(ctx, "test-index")
	require.NoError(t, err)
}

func TestStore_ListIndexes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/indexes")

		response := map[string]interface{}{
			"value": []map[string]interface{}{
				{"name": "index1"},
				{"name": "index2"},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		assert.NoError(t, err)
	}))
	defer server.Close()

	store := Store{
		azureAISearchEndpoint: server.URL,
		azureAISearchAPIKey:   "test-key",
		client:                httputil.DefaultClient,
	}

	ctx := context.Background()
	var result map[string]interface{}
	err := store.ListIndexes(ctx, &result)
	require.NoError(t, err)

	value, ok := result["value"].([]interface{})
	require.True(t, ok)
	assert.Len(t, value, 2)
}

func TestStore_RetrieveIndex(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/indexes/test-index")

		response := map[string]interface{}{
			"name": "test-index",
			"fields": []map[string]interface{}{
				{"name": "id", "type": "Edm.String"},
			},
		}
		err := json.NewEncoder(w).Encode(response)
		assert.NoError(t, err)
	}))
	defer server.Close()

	store := Store{
		azureAISearchEndpoint: server.URL,
		azureAISearchAPIKey:   "test-key",
		client:                httputil.DefaultClient,
	}

	ctx := context.Background()
	var result map[string]interface{}
	err := store.RetrieveIndex(ctx, "test-index", &result)
	require.NoError(t, err)
	assert.Equal(t, "test-index", result["name"])
}

func TestAssertResultValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       map[string]interface{}
		wantErr     bool
		errContains string
		validate    func(t *testing.T, doc *schema.Document)
	}{
		{
			name: "valid result",
			input: map[string]interface{}{
				"@search.score": 0.95,
				"content":       "Test content",
				"metadata":      `{"key":"value","number":123}`,
			},
			validate: func(t *testing.T, doc *schema.Document) {
				assert.Equal(t, "Test content", doc.PageContent)
				assert.Equal(t, float32(0.95), doc.Score)
				assert.Equal(t, "value", doc.Metadata["key"])
				assert.Equal(t, float64(123), doc.Metadata["number"])
			},
		},
		{
			name: "missing score",
			input: map[string]interface{}{
				"content":  "Test content",
				"metadata": `{}`,
			},
			wantErr:     true,
			errContains: "couldn't assert @search.score to float64",
		},
		{
			name: "invalid score type",
			input: map[string]interface{}{
				"@search.score": "not a number",
				"content":       "Test content",
				"metadata":      `{}`,
			},
			wantErr:     true,
			errContains: "couldn't assert @search.score to float64",
		},
		{
			name: "missing metadata",
			input: map[string]interface{}{
				"@search.score": 0.95,
				"content":       "Test content",
			},
			wantErr:     true,
			errContains: "couldn't assert metadata to string",
		},
		{
			name: "invalid metadata JSON",
			input: map[string]interface{}{
				"@search.score": 0.95,
				"content":       "Test content",
				"metadata":      `{invalid json}`,
			},
			wantErr:     true,
			errContains: "couldn't unmarshall metadata",
		},
		{
			name: "missing content",
			input: map[string]interface{}{
				"@search.score": 0.95,
				"metadata":      `{}`,
			},
			wantErr:     true,
			errContains: "couldn't assert content to string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := assertResultValues(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, doc)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 42,
			},
			expected: map[string]interface{}{
				"name":  "test",
				"value": float64(42),
			},
		},
		{
			name: "nested struct",
			input: struct {
				ID     string `json:"id"`
				Nested struct {
					Key string `json:"key"`
				} `json:"nested"`
			}{
				ID: "123",
				Nested: struct {
					Key string `json:"key"`
				}{
					Key: "value",
				},
			},
			expected: map[string]interface{}{
				"id": "123",
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
		{
			name:    "unmarshalable input",
			input:   make(chan int),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output map[string]interface{}
			err := structToMap(tt.input, &output)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestGetOptions(t *testing.T) {
	t.Parallel()

	store := Store{}

	tests := []struct {
		name     string
		options  []vectorstores.Option
		validate func(t *testing.T, opts vectorstores.Options)
	}{
		{
			name:    "no options",
			options: nil,
			validate: func(t *testing.T, opts vectorstores.Options) {
				assert.Empty(t, opts.NameSpace)
				assert.Nil(t, opts.Filters)
				assert.Zero(t, opts.ScoreThreshold)
			},
		},
		{
			name: "with namespace",
			options: []vectorstores.Option{
				vectorstores.WithNameSpace("test-namespace"),
			},
			validate: func(t *testing.T, opts vectorstores.Options) {
				assert.Equal(t, "test-namespace", opts.NameSpace)
			},
		},
		{
			name: "with score threshold",
			options: []vectorstores.Option{
				vectorstores.WithScoreThreshold(0.8),
			},
			validate: func(t *testing.T, opts vectorstores.Options) {
				assert.Equal(t, float32(0.8), opts.ScoreThreshold)
			},
		},
		{
			name: "with filters",
			options: []vectorstores.Option{
				WithFilters("category eq 'test'"),
			},
			validate: func(t *testing.T, opts vectorstores.Options) {
				assert.Equal(t, "category eq 'test'", opts.Filters)
			},
		},
		{
			name: "multiple options",
			options: []vectorstores.Option{
				vectorstores.WithNameSpace("namespace"),
				vectorstores.WithScoreThreshold(0.7),
				WithFilters("filter"),
			},
			validate: func(t *testing.T, opts vectorstores.Options) {
				assert.Equal(t, "namespace", opts.NameSpace)
				assert.Equal(t, float32(0.7), opts.ScoreThreshold)
				assert.Equal(t, "filter", opts.Filters)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := store.getOptions(tt.options...)
			tt.validate(t, opts)
		})
	}
}

func TestHTTPReadBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		response     *http.Response
		serviceName  string
		output       interface{}
		wantErr      bool
		errContains  string
		validateJSON func(t *testing.T, output interface{})
	}{
		{
			name: "successful JSON response",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"key":"value","number":42}`)),
			},
			serviceName: "test-service",
			output:      &map[string]interface{}{},
			validateJSON: func(t *testing.T, output interface{}) {
				m := output.(*map[string]interface{})
				assert.Equal(t, "value", (*m)["key"])
				assert.Equal(t, float64(42), (*m)["number"])
			},
		},
		{
			name: "successful no content",
			response: &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
			},
			serviceName: "test-service",
			output:      nil,
		},
		{
			name: "error response",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "Bad Request",
				Body:       io.NopCloser(strings.NewReader(`{"error":"bad request"}`)),
			},
			serviceName: "test-service",
			wantErr:     true,
			errContains: "error returned from test-service",
		},
		{
			name: "invalid JSON",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{invalid json}`)),
			},
			serviceName: "test-service",
			output:      &map[string]interface{}{},
			wantErr:     true,
			errContains: "err unmarshal body for test-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := httpReadBody(tt.response, tt.serviceName, tt.output)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validateJSON != nil {
				tt.validateJSON(t, tt.output)
			}
		})
	}
}

func TestWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		option   Option
		validate func(t *testing.T, s *Store)
	}{
		{
			name:   "WithEmbedder",
			option: WithEmbedder(&testEmbedder{}),
			validate: func(t *testing.T, s *Store) {
				assert.NotNil(t, s.embedder)
			},
		},
		{
			name:   "WithHTTPClient",
			option: WithHTTPClient(&http.Client{Timeout: 30}),
			validate: func(t *testing.T, s *Store) {
				assert.NotNil(t, s.client)
			},
		},
		{
			name:   "WithAPIKey",
			option: WithAPIKey("test-api-key"),
			validate: func(t *testing.T, s *Store) {
				assert.Equal(t, "test-api-key", s.azureAISearchAPIKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{}
			tt.option(s)
			tt.validate(t, s)
		})
	}
}

func TestEnvironmentVariableHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		envVars     map[string]string
		options     []Option
		wantErr     bool
		errContains string
		validate    func(t *testing.T, s Store)
	}{
		{
			name: "endpoint from env",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://env.search.windows.net",
			},
			options: []Option{WithEmbedder(&testEmbedder{})},
			validate: func(t *testing.T, s Store) {
				assert.Equal(t, "https://env.search.windows.net", s.azureAISearchEndpoint)
			},
		},
		{
			name: "api key from env",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net",
				"AZURE_AI_SEARCH_API_KEY":  "env-key",
			},
			options: []Option{WithEmbedder(&testEmbedder{})},
			validate: func(t *testing.T, s Store) {
				assert.Equal(t, "env-key", s.azureAISearchAPIKey)
			},
		},
		{
			name: "env overrides option for api key",
			envVars: map[string]string{
				"AZURE_AI_SEARCH_ENDPOINT": "https://test.search.windows.net",
				"AZURE_AI_SEARCH_API_KEY":  "env-key",
			},
			options: []Option{
				WithEmbedder(&testEmbedder{}),
				WithAPIKey("option-key"),
			},
			validate: func(t *testing.T, s Store) {
				// Environment variable takes precedence over option
				assert.Equal(t, "env-key", s.azureAISearchAPIKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			store, err := New(tt.options...)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, store)
			}
		})
	}
}

func TestDocumentUploadEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		docs        []schema.Document
		mockServer  func() *httptest.Server
		wantErr     bool
		errContains string
	}{
		{
			name: "empty documents",
			docs: []schema.Document{},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fatal("should not make any requests")
				}))
			},
			wantErr: false,
		},
		{
			name: "document with empty content",
			docs: []schema.Document{
				{PageContent: "", Metadata: map[string]any{"key": "value"}},
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{"value":[]}`)
				}))
			},
		},
		{
			name: "document with nil metadata",
			docs: []schema.Document{
				{PageContent: "content", Metadata: nil},
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{"value":[]}`)
				}))
			},
		},
		{
			name: "large metadata",
			docs: []schema.Document{
				{
					PageContent: "content",
					Metadata: map[string]any{
						"large_array": make([]int, 1000),
					},
				},
			},
			mockServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, `{"value":[]}`)
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.mockServer()
			defer server.Close()

			store := Store{
				azureAISearchEndpoint: server.URL,
				azureAISearchAPIKey:   "test-key",
				embedder:              &testEmbedder{},
				client:                httputil.DefaultClient,
			}

			ctx := context.Background()
			ids, err := store.AddDocuments(ctx, tt.docs)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Len(t, ids, len(tt.docs))
		})
	}
}

func TestSearchWithComplexFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filter   string
		validate func(t *testing.T, sentFilter string)
	}{
		{
			name:   "simple equality filter",
			filter: "category eq 'technology'",
			validate: func(t *testing.T, sentFilter string) {
				assert.Equal(t, "category eq 'technology'", sentFilter)
			},
		},
		{
			name:   "complex AND filter",
			filter: "category eq 'technology' and rating gt 4",
			validate: func(t *testing.T, sentFilter string) {
				assert.Equal(t, "category eq 'technology' and rating gt 4", sentFilter)
			},
		},
		{
			name:   "complex OR filter",
			filter: "(category eq 'technology' or category eq 'science') and published ge 2023-01-01",
			validate: func(t *testing.T, sentFilter string) {
				assert.Equal(t, "(category eq 'technology' or category eq 'science') and published ge 2023-01-01", sentFilter)
			},
		},
		{
			name:   "search.in filter",
			filter: "search.in(tags, 'ai,ml,deeplearning')",
			validate: func(t *testing.T, sentFilter string) {
				assert.Equal(t, "search.in(tags, 'ai,ml,deeplearning')", sentFilter)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedFilter string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req SearchDocumentsRequestInput
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				capturedFilter = req.Filter

				response := SearchDocumentsRequestOuput{
					Value: []map[string]interface{}{
						{
							"@search.score": 0.9,
							"content":       "Result",
							"metadata":      `{}`,
						},
					},
				}
				if err := json.NewEncoder(w).Encode(response); err != nil {
					t.Errorf("Failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			store := Store{
				azureAISearchEndpoint: server.URL,
				azureAISearchAPIKey:   "test-key",
				embedder:              &testEmbedder{},
				client:                httputil.DefaultClient,
			}

			ctx := context.Background()
			_, err := store.SimilaritySearch(ctx, "query", 1, WithFilters(tt.filter))
			require.NoError(t, err)

			tt.validate(t, capturedFilter)
		})
	}
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	// Create a server that handles concurrent requests
	var requestCount int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()

		// Simulate some processing time
		if strings.Contains(r.URL.Path, "/docs/search") {
			response := SearchDocumentsRequestOuput{
				Value: []map[string]interface{}{
					{
						"@search.score": 0.9,
						"content":       fmt.Sprintf("Result %d", requestCount),
						"metadata":      `{}`,
					},
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		} else if strings.Contains(r.URL.Path, "/docs/index") {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"value":[]}`)
		}
	}))
	defer server.Close()

	store := Store{
		azureAISearchEndpoint: server.URL,
		azureAISearchAPIKey:   "test-key",
		embedder:              &testEmbedder{},
		client:                httputil.DefaultClient,
	}

	ctx := context.Background()
	const numGoroutines = 10

	// Test concurrent similarity searches
	t.Run("concurrent searches", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_, err := store.SimilaritySearch(ctx, fmt.Sprintf("query %d", idx), 1)
				errors[idx] = err
			}(i)
		}

		wg.Wait()

		for i, err := range errors {
			assert.NoError(t, err, "goroutine %d failed", i)
		}
	})

	// Test concurrent document additions
	t.Run("concurrent add documents", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				docs := []schema.Document{
					{
						PageContent: fmt.Sprintf("Document %d", idx),
						Metadata:    map[string]any{"id": idx},
					},
				}
				_, err := store.AddDocuments(ctx, docs)
				errors[idx] = err
			}(i)
		}

		wg.Wait()

		for i, err := range errors {
			assert.NoError(t, err, "goroutine %d failed", i)
		}
	})
}

// TestInterfaceCompliance verifies that Store implements the VectorStore interface
func TestInterfaceCompliance(t *testing.T) {
	var _ vectorstores.VectorStore = &Store{}
	var _ embeddings.Embedder = &testEmbedder{}
}
