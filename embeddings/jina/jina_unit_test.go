package jina

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJina(t *testing.T) { //nolint:funlen // comprehensive test
	// Save and restore environment variable
	oldAPIKey := os.Getenv("JINA_API_KEY")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("JINA_API_KEY", oldAPIKey)
		} else {
			os.Unsetenv("JINA_API_KEY")
		}
	}()

	tests := []struct {
		name    string
		opts    []Option
		envVars map[string]string
		wantErr bool
		check   func(t *testing.T, j *Jina)
	}{
		{
			name: "default options with env var",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, "jina-embeddings-v2-small-en", j.Model)
				assert.Equal(t, true, j.StripNewLines)
				assert.Equal(t, 512, j.BatchSize)
				assert.Equal(t, "https://api.jina.ai/v1/embeddings", j.APIBaseURL)
				assert.Equal(t, "test-key", j.APIKey)
			},
		},
		{
			name: "with model option - small",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithModel(SmallModel),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, SmallModel, j.Model)
				assert.Equal(t, 512, j.BatchSize)
			},
		},
		{
			name: "with model option - base",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithModel(BaseModel),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, BaseModel, j.Model)
				assert.Equal(t, 768, j.BatchSize)
			},
		},
		{
			name: "with model option - large",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithModel(LargeModel),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, LargeModel, j.Model)
				assert.Equal(t, 1024, j.BatchSize)
			},
		},
		{
			name: "with unknown model",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithModel("unknown-model"),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, "unknown-model", j.Model)
				assert.Equal(t, 512, j.BatchSize) // Should use default batch size
			},
		},
		{
			name: "with strip new lines option",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithStripNewLines(false),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, false, j.StripNewLines)
			},
		},
		{
			name: "with batch size option",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithBatchSize(256),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, 512, j.BatchSize) // Batch size is overridden by model
			},
		},
		{
			name: "with API base URL option",
			envVars: map[string]string{
				"JINA_API_KEY": "test-key",
			},
			opts: []Option{
				WithAPIBaseURL("https://custom.jina.ai/v1/embeddings"),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, "https://custom.jina.ai/v1/embeddings", j.APIBaseURL)
			},
		},
		{
			name: "with API key option",
			opts: []Option{
				WithAPIKey("custom-api-key"),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, "custom-api-key", j.APIKey)
			},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithModel(BaseModel),
				WithStripNewLines(false),
				WithAPIBaseURL("https://custom.jina.ai/v1/embeddings"),
				WithAPIKey("custom-key"),
			},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, BaseModel, j.Model)
				assert.Equal(t, false, j.StripNewLines)
				assert.Equal(t, 768, j.BatchSize)
				assert.Equal(t, "https://custom.jina.ai/v1/embeddings", j.APIBaseURL)
				assert.Equal(t, "custom-key", j.APIKey)
			},
		},
		{
			name: "without API key",
			opts: []Option{},
			check: func(t *testing.T, j *Jina) {
				assert.Equal(t, "", j.APIKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			j, err := NewJina(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, j)
				if tt.check != nil {
					tt.check(t, j)
				}
			}
		})
	}
}

func TestJinaOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		j := &Jina{}
		WithModel("test-model")(j)
		assert.Equal(t, "test-model", j.Model)
	})

	t.Run("WithStripNewLines", func(t *testing.T) {
		j := &Jina{StripNewLines: true}
		WithStripNewLines(false)(j)
		assert.Equal(t, false, j.StripNewLines)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		j := &Jina{}
		WithBatchSize(256)(j)
		assert.Equal(t, 256, j.BatchSize)
	})

	t.Run("WithAPIBaseURL", func(t *testing.T) {
		j := &Jina{}
		WithAPIBaseURL("https://custom.api.com")(j)
		assert.Equal(t, "https://custom.api.com", j.APIBaseURL)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		j := &Jina{}
		WithAPIKey("test-api-key")(j)
		assert.Equal(t, "test-api-key", j.APIKey)
	})

	t.Run("WithClient", func(t *testing.T) {
		j := &Jina{}
		customClient := &http.Client{}
		WithClient(customClient)(j)
		assert.Equal(t, customClient, j.client)
	})
}

func TestJina_EmbedQuery_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	j := &Jina{
		Model:         SmallModel,
		APIBaseURL:    server.URL,
		APIKey:        "test-key",
		StripNewLines: true,
		client:        &http.Client{},
	}

	ctx := context.Background()

	// Test with CreateEmbedding returning an error from the mock server
	_, err := j.EmbedQuery(ctx, "test query")
	require.Error(t, err)
}

func TestJina_EmbedDocuments_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	j := &Jina{
		Model:         SmallModel,
		APIBaseURL:    server.URL,
		APIKey:        "test-key",
		StripNewLines: true,
		BatchSize:     512,
		client:        &http.Client{},
	}

	ctx := context.Background()

	// Test with CreateEmbedding returning an error from the mock server
	_, err := j.EmbedDocuments(ctx, []string{"doc1", "doc2"})
	require.Error(t, err)
}

func TestApplyOptions(t *testing.T) {
	// Save and restore environment variable
	oldAPIKey := os.Getenv("JINA_API_KEY")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("JINA_API_KEY", oldAPIKey)
		} else {
			os.Unsetenv("JINA_API_KEY")
		}
	}()

	t.Run("with environment variable", func(t *testing.T) {
		os.Setenv("JINA_API_KEY", "env-api-key")
		j := applyOptions()
		assert.Equal(t, "env-api-key", j.APIKey)
		assert.Equal(t, _defaultModel, j.Model)
		assert.Equal(t, _defaultStripNewLines, j.StripNewLines)
		assert.Equal(t, APIBaseURL, j.APIBaseURL)
		assert.Equal(t, 512, j.BatchSize)
	})

	t.Run("model batch size mapping", func(t *testing.T) {
		tests := []struct {
			model        string
			expectedSize int
		}{
			{SmallModel, 512},
			{BaseModel, 768},
			{LargeModel, 1024},
			{"unknown-model", 512}, // Default size
		}

		for _, tt := range tests {
			t.Run(tt.model, func(t *testing.T) {
				j := applyOptions(WithModel(tt.model), WithAPIKey("test"))
				assert.Equal(t, tt.model, j.Model)
				assert.Equal(t, tt.expectedSize, j.BatchSize)
			})
		}
	})
}
