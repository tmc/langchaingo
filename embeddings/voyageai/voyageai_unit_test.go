package voyageai

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/httputil"
)

func TestNewVoyageAI(t *testing.T) { //nolint:funlen // comprehensive test
	// Save and restore environment variable
	oldAPIKey := os.Getenv("VOYAGEAI_API_KEY")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("VOYAGEAI_API_KEY", oldAPIKey)
		} else {
			os.Unsetenv("VOYAGEAI_API_KEY")
		}
	}()

	tests := []struct {
		name    string
		opts    []Option
		envVars map[string]string
		wantErr bool
		errMsg  string
		check   func(t *testing.T, v *VoyageAI)
	}{
		{
			name: "default options with env var",
			envVars: map[string]string{
				"VOYAGEAI_API_KEY": "test-key",
			},
			opts: []Option{},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, _defaultModel, v.Model)
				assert.Equal(t, _defaultStripNewLines, v.StripNewLines)
				assert.Equal(t, _defaultBatchSize, v.BatchSize)
				assert.Equal(t, _defaultBaseURL, v.baseURL)
				assert.Equal(t, "test-key", v.token)
				assert.NotNil(t, v.client)
			},
		},
		{
			name:    "missing API key",
			opts:    []Option{},
			wantErr: true,
			errMsg:  "missing the VoyageAI API key, set it as VOYAGEAI_API_KEY environment variable",
		},
		{
			name: "with model option",
			envVars: map[string]string{
				"VOYAGEAI_API_KEY": "test-key",
			},
			opts: []Option{
				WithModel("voyage-large-2"),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, "voyage-large-2", v.Model)
			},
		},
		{
			name: "with token option",
			opts: []Option{
				WithToken("custom-token"),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, "custom-token", v.token)
			},
		},
		{
			name: "with strip new lines option",
			opts: []Option{
				WithToken("test-token"),
				WithStripNewLines(false),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, false, v.StripNewLines)
			},
		},
		{
			name: "with batch size option",
			opts: []Option{
				WithToken("test-token"),
				WithBatchSize(256),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, 256, v.BatchSize)
			},
		},
		{
			name: "with custom client option",
			opts: []Option{
				WithToken("test-token"),
				WithClient(http.Client{Timeout: 0}),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.NotNil(t, v.client)
				// Note: We can't directly compare the client since it's wrapped
			},
		},
		{
			name: "with multiple options",
			opts: []Option{
				WithModel("voyage-3"),
				WithToken("custom-token"),
				WithStripNewLines(false),
				WithBatchSize(128),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, "voyage-3", v.Model)
				assert.Equal(t, "custom-token", v.token)
				assert.Equal(t, false, v.StripNewLines)
				assert.Equal(t, 128, v.BatchSize)
			},
		},
		{
			name: "token from env overridden by option",
			envVars: map[string]string{
				"VOYAGEAI_API_KEY": "env-key",
			},
			opts: []Option{
				WithToken("option-key"),
			},
			check: func(t *testing.T, v *VoyageAI) {
				assert.Equal(t, "option-key", v.token)
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

			v, err := NewVoyageAI(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.EqualError(t, err, tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, v)
				if tt.check != nil {
					tt.check(t, v)
				}
			}
		})
	}
}

func TestVoyageAIOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		v := &VoyageAI{}
		WithModel("test-model")(v)
		assert.Equal(t, "test-model", v.Model)
	})

	t.Run("WithClient", func(t *testing.T) {
		v := &VoyageAI{}
		client := http.Client{Timeout: 0}
		WithClient(client)(v)
		assert.NotNil(t, v.client)
	})

	t.Run("WithToken", func(t *testing.T) {
		v := &VoyageAI{}
		WithToken("test-token")(v)
		assert.Equal(t, "test-token", v.token)
	})

	t.Run("WithStripNewLines", func(t *testing.T) {
		v := &VoyageAI{StripNewLines: true}
		WithStripNewLines(false)(v)
		assert.Equal(t, false, v.StripNewLines)
	})

	t.Run("WithBatchSize", func(t *testing.T) {
		v := &VoyageAI{}
		WithBatchSize(256)(v)
		assert.Equal(t, 256, v.BatchSize)
	})
}

func TestVoyageAI_EmbedQuery_InvalidURL(t *testing.T) {
	v := &VoyageAI{
		baseURL:       "://invalid-url", // Invalid URL to trigger error
		token:         "test-token",
		Model:         _defaultModel,
		StripNewLines: true,
		BatchSize:     _defaultBatchSize,
		client:        httputil.DefaultClient,
	}

	ctx := context.Background()

	_, err := v.EmbedQuery(ctx, "test query")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed query request error")
}

func TestVoyageAI_EmbedDocuments_InvalidURL(t *testing.T) {
	v := &VoyageAI{
		baseURL:       "://invalid-url", // Invalid URL to trigger error
		token:         "test-token",
		Model:         _defaultModel,
		StripNewLines: true,
		BatchSize:     _defaultBatchSize,
		client:        httputil.DefaultClient,
	}

	ctx := context.Background()

	_, err := v.EmbedDocuments(ctx, []string{"doc1", "doc2"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed documents request error")
}

func TestApplyOptions(t *testing.T) {
	// Save and restore environment variable
	oldAPIKey := os.Getenv("VOYAGEAI_API_KEY")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("VOYAGEAI_API_KEY", oldAPIKey)
		} else {
			os.Unsetenv("VOYAGEAI_API_KEY")
		}
	}()

	t.Run("with environment variable", func(t *testing.T) {
		os.Setenv("VOYAGEAI_API_KEY", "env-api-key")
		v, err := applyOptions()
		assert.NoError(t, err)
		assert.Equal(t, "env-api-key", v.token)
		assert.Equal(t, _defaultModel, v.Model)
		assert.Equal(t, _defaultStripNewLines, v.StripNewLines)
		assert.Equal(t, _defaultBatchSize, v.BatchSize)
		assert.Equal(t, _defaultBaseURL, v.baseURL)
		assert.NotNil(t, v.client)
	})

	t.Run("without environment variable", func(t *testing.T) {
		os.Unsetenv("VOYAGEAI_API_KEY")
		v, err := applyOptions()
		assert.Error(t, err)
		assert.Nil(t, v)
		assert.EqualError(t, err, "missing the VoyageAI API key, set it as VOYAGEAI_API_KEY environment variable")
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := []Option{
			WithToken("custom-token"),
			WithModel("voyage-large"),
			WithBatchSize(256),
			WithStripNewLines(false),
		}
		v, err := applyOptions(opts...)
		assert.NoError(t, err)
		assert.Equal(t, "custom-token", v.token)
		assert.Equal(t, "voyage-large", v.Model)
		assert.Equal(t, 256, v.BatchSize)
		assert.Equal(t, false, v.StripNewLines)
	})

	t.Run("client defaults to httputil.DefaultClient", func(t *testing.T) {
		v, err := applyOptions(WithToken("test"))
		assert.NoError(t, err)
		assert.NotNil(t, v.client)
	})
}
