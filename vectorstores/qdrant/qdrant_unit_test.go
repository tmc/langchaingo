package qdrant

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/vectorstores"
)

// testEmbedder is a mock embedder for unit testing
type testEmbedder struct {
	embedFn func(context.Context, []string) ([][]float32, error)
}

func (m *testEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if m.embedFn != nil {
		return m.embedFn(ctx, texts)
	}
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = []float32{float32(len(texts[i])), 0.1, 0.2, 0.3}
	}
	return embeddings, nil
}

func (m *testEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if m.embedFn != nil {
		vecs, err := m.embedFn(ctx, []string{text})
		if err != nil || len(vecs) == 0 {
			return nil, err
		}
		return vecs[0], nil
	}
	return []float32{float32(len(text)), 0.1, 0.2, 0.3}, nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        []Option
		wantErr     bool
		errContains string
	}{
		{
			name: "success with all required options",
			opts: []Option{
				WithURL(url.URL{Scheme: "http", Host: "localhost:6333"}),
				WithCollectionName("test-collection"),
				WithEmbedder(&testEmbedder{}),
			},
			wantErr: false,
		},
		{
			name: "missing collection name",
			opts: []Option{
				WithURL(url.URL{Scheme: "http", Host: "localhost:6333"}),
				WithEmbedder(&testEmbedder{}),
			},
			wantErr:     true,
			errContains: "missing collection name",
		},
		{
			name: "missing URL",
			opts: []Option{
				WithCollectionName("test-collection"),
				WithEmbedder(&testEmbedder{}),
			},
			wantErr:     true,
			errContains: "missing Qdrant URL",
		},
		{
			name: "missing embedder",
			opts: []Option{
				WithURL(url.URL{Scheme: "http", Host: "localhost:6333"}),
				WithCollectionName("test-collection"),
			},
			wantErr:     true,
			errContains: "missing embedder",
		},
		{
			name: "with all options",
			opts: []Option{
				WithURL(url.URL{Scheme: "http", Host: "localhost:6333"}),
				WithCollectionName("test-collection"),
				WithEmbedder(&testEmbedder{}),
				WithAPIKey("test-key"),
				WithContentKey("custom-content"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStore_AddDocuments_Unit(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	tests := []struct {
		name         string
		docs         []schema.Document
		opts         []vectorstores.Option
		mockResponse func(w http.ResponseWriter, r *http.Request)
		embedFn      func(context.Context, []string) ([][]float32, error)
		wantErr      bool
		errContains  string
		wantIDs      int
	}{
		{
			name: "success",
			docs: []schema.Document{
				{PageContent: "doc1", Metadata: map[string]any{"key": "value1"}},
				{PageContent: "doc2", Metadata: map[string]any{"key": "value2"}},
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				var req upsertBody
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				// URL path should be /collections/test-collection/points/search
				pathParts := strings.Split(r.URL.Path, "/")
				assert.Equal(t, "collections", pathParts[1])
				assert.Equal(t, "test-collection", pathParts[2])
				if len(pathParts) > 3 {
					assert.Equal(t, "points", pathParts[3])
				}
				assert.Equal(t, 2, len(req.Batch.IDs))

				w.WriteHeader(http.StatusOK)
				err = json.NewEncoder(w).Encode(map[string]any{
					"status": "ok",
					"result": map[string]any{
						"operation_id": 123,
						"status":       "acknowledged",
					},
				})
				assert.NoError(t, err)
			},
			wantIDs: 2,
			wantErr: false,
		},
		{
			name: "empty documents",
			docs: []schema.Document{},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Should not make API call for empty documents")
			},
			wantIDs: 0,
			wantErr: false,
		},
		{
			name: "embedding error",
			docs: []schema.Document{
				{PageContent: "doc1"},
			},
			embedFn: func(ctx context.Context, docs []string) ([][]float32, error) {
				return nil, errors.New("embedding failed")
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Should not make API call on embedding error")
			},
			wantErr:     true,
			errContains: "embedding failed",
		},
		{
			name: "wrong number of embeddings",
			docs: []schema.Document{
				{PageContent: "doc1"},
				{PageContent: "doc2"},
			},
			embedFn: func(ctx context.Context, docs []string) ([][]float32, error) {
				// Return only one embedding for two documents
				return [][]float32{{0.1, 0.2, 0.3}}, nil
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Should not make API call with wrong number of embeddings")
			},
			wantErr:     true,
			errContains: "number of vectors",
		},
		{
			name: "API error",
			docs: []schema.Document{
				{PageContent: "doc1"},
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				err := json.NewEncoder(w).Encode(map[string]any{
					"status": "error",
					"result": map[string]any{
						"error": "Invalid request",
					},
				})
				assert.NoError(t, err)
			},
			wantErr:     true,
			errContains: "upserting vectors:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResponse != nil {
					tt.mockResponse(w, r)
				} else {
					w.WriteHeader(http.StatusOK)
					err := json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
					assert.NoError(t, err)
				}
			}))
			defer server.Close()

			serverURL, _ := url.Parse(server.URL)
			embedder := &testEmbedder{embedFn: tt.embedFn}

			store, err := New(
				WithURL(*serverURL),
				WithCollectionName("test-collection"),
				WithEmbedder(embedder),
			)
			require.NoError(t, err)

			ids, err := store.AddDocuments(context.Background(), tt.docs, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, ids, tt.wantIDs)
			}
		})
	}
}

func TestStore_SimilaritySearch_Unit(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	tests := []struct {
		name         string
		query        string
		numDocuments int
		opts         []vectorstores.Option
		mockResponse func(w http.ResponseWriter, r *http.Request)
		embedFn      func(context.Context, []string) ([][]float32, error)
		wantDocs     int
		wantErr      bool
		errContains  string
	}{
		{
			name:         "success",
			query:        "test query",
			numDocuments: 2,
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				var req searchBody
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				// URL path should be /collections/test-collection/points/search
				pathParts := strings.Split(r.URL.Path, "/")
				assert.Equal(t, "collections", pathParts[1])
				assert.Equal(t, "test-collection", pathParts[2])
				if len(pathParts) > 3 {
					assert.Equal(t, "points", pathParts[3])
				}
				assert.Equal(t, 2, req.Limit)
				assert.NotNil(t, req.Vector)
				assert.True(t, req.WithPayload)

				w.WriteHeader(http.StatusOK)
				err = json.NewEncoder(w).Encode(searchResponse{
					Result: []result{
						{
							Score: 0.95,
							Payload: map[string]any{
								"content": "Result 1",
								"meta":    "data1",
							},
						},
						{
							Score: 0.85,
							Payload: map[string]any{
								"content": "Result 2",
								"meta":    "data2",
							},
						},
					},
				})
				assert.NoError(t, err)
			},
			wantDocs: 2,
			wantErr:  false,
		},
		{
			name:         "with score threshold",
			query:        "test query",
			numDocuments: 3,
			opts: []vectorstores.Option{
				vectorstores.WithScoreThreshold(0.9),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				var req searchBody
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, float32(0.9), req.ScoreThreshold)

				w.WriteHeader(http.StatusOK)
				err = json.NewEncoder(w).Encode(searchResponse{
					Result: []result{
						{
							Score: 0.95,
							Payload: map[string]any{
								"content": "High score result",
							},
						},
					},
				})
				assert.NoError(t, err)
			},
			wantDocs: 1,
			wantErr:  false,
		},
		{
			name:         "invalid score threshold",
			query:        "test query",
			numDocuments: 2,
			opts: []vectorstores.Option{
				vectorstores.WithScoreThreshold(1.5),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Should not make API call with invalid score threshold")
			},
			wantErr:     true,
			errContains: "score threshold",
		},
		{
			name:         "with filters",
			query:        "test query",
			numDocuments: 2,
			opts: []vectorstores.Option{
				vectorstores.WithFilters(map[string]any{"category": "tech"}),
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				var req searchBody
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.NotNil(t, req.Filter)

				w.WriteHeader(http.StatusOK)
				err = json.NewEncoder(w).Encode(searchResponse{
					Result: []result{
						{
							Score: 0.9,
							Payload: map[string]any{
								"content":  "Filtered result",
								"category": "tech",
							},
						},
					},
				})
				assert.NoError(t, err)
			},
			wantDocs: 1,
			wantErr:  false,
		},
		{
			name:         "embedding error",
			query:        "test query",
			numDocuments: 2,
			embedFn: func(ctx context.Context, docs []string) ([][]float32, error) {
				return nil, errors.New("embedding query failed")
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				t.Error("Should not make API call on embedding error")
			},
			wantErr:     true,
			errContains: "embedding query failed",
		},
		{
			name:         "missing content key in response",
			query:        "test query",
			numDocuments: 1,
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(searchResponse{
					Result: []result{
						{
							Score: 0.9,
							Payload: map[string]any{
								"other": "data",
							},
						},
					},
				}); err != nil {
					panic(err) // Test helper, panic is acceptable
				}
			},
			wantErr:     true,
			errContains: "does not contain content key",
		},
		{
			name:         "API error",
			query:        "test query",
			numDocuments: 2,
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(map[string]any{
					"status": "error",
					"result": map[string]any{
						"error": "Server error",
					},
				}); err != nil {
					panic(err) // Test helper, panic is acceptable
				}
			},
			wantErr:     true,
			errContains: "querying collection:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResponse != nil {
					tt.mockResponse(w, r)
				} else {
					w.WriteHeader(http.StatusOK)
					if err := json.NewEncoder(w).Encode(searchResponse{}); err != nil {
						panic(err) // Test helper, panic is acceptable
					}
				}
			}))
			defer server.Close()

			serverURL, _ := url.Parse(server.URL)
			embedder := &testEmbedder{embedFn: tt.embedFn}

			store, err := New(
				WithURL(*serverURL),
				WithCollectionName("test-collection"),
				WithEmbedder(embedder),
			)
			require.NoError(t, err)

			docs, err := store.SimilaritySearch(context.Background(), tt.query, tt.numDocuments, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, docs, tt.wantDocs)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithURL", func(t *testing.T) {
		url := url.URL{Scheme: "http", Host: "localhost:6333"}
		opt := WithURL(url)
		store := &Store{}
		opt(store)
		assert.Equal(t, url, store.qdrantURL)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		key := "test-api-key"
		opt := WithAPIKey(key)
		store := &Store{}
		opt(store)
		assert.Equal(t, key, store.apiKey)
	})

	t.Run("WithCollectionName", func(t *testing.T) {
		name := "test-collection"
		opt := WithCollectionName(name)
		store := &Store{}
		opt(store)
		assert.Equal(t, name, store.collectionName)
	})

	t.Run("WithEmbedder", func(t *testing.T) {
		embedder := &testEmbedder{}
		opt := WithEmbedder(embedder)
		store := &Store{}
		opt(store)
		assert.Equal(t, embedder, store.embedder)
	})

	t.Run("WithContentKey", func(t *testing.T) {
		key := "custom-content"
		opt := WithContentKey(key)
		store := &Store{}
		opt(store)
		assert.Equal(t, key, store.contentKey)
	})
}

func TestGetScoreThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		opts          vectorstores.Options
		expectedScore float32
		expectError   bool
	}{
		{
			name:          "valid score 0.5",
			opts:          vectorstores.Options{ScoreThreshold: 0.5},
			expectedScore: 0.5,
			expectError:   false,
		},
		{
			name:          "score threshold zero",
			opts:          vectorstores.Options{ScoreThreshold: 0},
			expectedScore: 0,
			expectError:   false,
		},
		{
			name:          "score threshold one",
			opts:          vectorstores.Options{ScoreThreshold: 1},
			expectedScore: 1,
			expectError:   false,
		},
		{
			name:        "negative score",
			opts:        vectorstores.Options{ScoreThreshold: -0.1},
			expectError: true,
		},
		{
			name:        "score greater than 1",
			opts:        vectorstores.Options{ScoreThreshold: 1.1},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{}
			score, err := store.getScoreThreshold(tt.opts)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "score threshold must be between 0 and 1")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedScore, score)
			}
		})
	}
}

func TestGetFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		opts           vectorstores.Options
		expectedFilter any
	}{
		{
			name:           "no filters",
			opts:           vectorstores.Options{},
			expectedFilter: nil,
		},
		{
			name: "with map filter",
			opts: vectorstores.Options{
				Filters: map[string]any{
					"category": "tech",
					"year":     2024,
				},
			},
			expectedFilter: map[string]any{
				"category": "tech",
				"year":     2024,
			},
		},
		{
			name: "with struct filter",
			opts: vectorstores.Options{
				Filters: struct {
					Category string
					Year     int
				}{
					Category: "tech",
					Year:     2024,
				},
			},
			expectedFilter: struct {
				Category string
				Year     int
			}{
				Category: "tech",
				Year:     2024,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{}
			filter := store.getFilters(tt.opts)
			assert.Equal(t, tt.expectedFilter, filter)
		})
	}
}

func TestNewAPIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		task        string
		body        string
		expectedMsg string
	}{
		{
			name:        "simple error",
			task:        "upserting vectors",
			body:        `{"error": "Invalid request"}`,
			expectedMsg: "upserting vectors: {\"error\": \"Invalid request\"}",
		},
		{
			name:        "empty body",
			task:        "querying collection",
			body:        "",
			expectedMsg: "querying collection: ",
		},
		{
			name:        "complex error",
			task:        "searching points",
			body:        `{"status": "error", "result": {"error": "Collection not found"}}`,
			expectedMsg: `searching points: {"status": "error", "result": {"error": "Collection not found"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(tt.body))
			err := newAPIError(tt.task, body)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedMsg)
		})
	}
}

func TestStoreImplementsVectorStore(t *testing.T) {
	t.Parallel()

	// This test ensures Store implements the vectorstores.VectorStore interface
	var _ vectorstores.VectorStore = &Store{}
}

func TestEdgeCases(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	t.Run("AddDocuments with nil metadata", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req upsertBody
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			// Verify nil metadata is handled properly
			assert.Len(t, req.Batch.IDs, 1)
			assert.NotNil(t, req.Batch.Payloads)

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
			assert.NoError(t, err)
		}))
		defer server.Close()

		serverURL, _ := url.Parse(server.URL)
		store, err := New(
			WithURL(*serverURL),
			WithCollectionName("test"),
			WithEmbedder(&testEmbedder{}),
		)
		require.NoError(t, err)

		docs := []schema.Document{
			{PageContent: "test", Metadata: nil},
		}
		_, err = store.AddDocuments(context.Background(), docs)
		assert.NoError(t, err)
	})

	t.Run("AddDocuments with custom content key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req upsertBody
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			// Verify custom content key is used
			assert.Contains(t, req.Batch.Payloads[0], "custom-content")

			w.WriteHeader(http.StatusOK)
			err = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
			assert.NoError(t, err)
		}))
		defer server.Close()

		serverURL, _ := url.Parse(server.URL)
		store, err := New(
			WithURL(*serverURL),
			WithCollectionName("test"),
			WithEmbedder(&testEmbedder{}),
			WithContentKey("custom-content"),
		)
		require.NoError(t, err)

		docs := []schema.Document{
			{PageContent: "test content"},
		}
		_, err = store.AddDocuments(context.Background(), docs)
		assert.NoError(t, err)
	})

	t.Run("SimilaritySearch with custom content key", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(searchResponse{
				Result: []result{
					{
						Score: 0.9,
						Payload: map[string]any{
							"custom-content": "Custom result",
							"other":          "data",
						},
					},
				},
			})
			assert.NoError(t, err)
		}))
		defer server.Close()

		serverURL, _ := url.Parse(server.URL)
		store, err := New(
			WithURL(*serverURL),
			WithCollectionName("test"),
			WithEmbedder(&testEmbedder{}),
			WithContentKey("custom-content"),
		)
		require.NoError(t, err)

		docs, err := store.SimilaritySearch(context.Background(), "query", 1)
		assert.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, "Custom result", docs[0].PageContent)
		assert.Equal(t, "data", docs[0].Metadata["other"])
		// custom-content should be removed from metadata
		assert.NotContains(t, docs[0].Metadata, "custom-content")
	})
}

func TestDoRequest_Unit(t *testing.T) { //nolint:funlen // comprehensive test
	t.Parallel()

	tests := []struct {
		name         string
		method       string
		payload      interface{}
		mockResponse func(w http.ResponseWriter, r *http.Request)
		wantStatus   int
		wantErr      bool
		errContains  string
	}{
		{
			name:   "GET request success",
			method: http.MethodGet,
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "test-key", r.Header.Get("api-Key"))
				assert.Contains(t, r.URL.RawQuery, "wait=true")

				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(map[string]any{"status": "ok"}); err != nil {
					panic(err) // Test helper, panic is acceptable
				}
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "POST request with payload",
			method: http.MethodPost,
			payload: map[string]any{
				"test": "data",
			},
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				var body map[string]any
				err := json.NewDecoder(r.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "data", body["test"])

				w.WriteHeader(http.StatusCreated)
				err = json.NewEncoder(w).Encode(map[string]any{"id": "123"})
				assert.NoError(t, err)
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:   "Error response",
			method: http.MethodGet,
			mockResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				if err := json.NewEncoder(w).Encode(map[string]any{
					"error": "Not found",
				}); err != nil {
					panic(err) // Test helper, panic is acceptable
				}
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.mockResponse != nil {
					tt.mockResponse(w, r)
				}
			}))
			defer server.Close()

			testURL, _ := url.Parse(server.URL + "/test")
			body, status, err := DoRequest(
				context.Background(),
				*testURL,
				"test-key",
				tt.method,
				tt.payload,
			)

			assert.Equal(t, tt.wantStatus, status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, body)
				body.Close()
			}
		})
	}
}
