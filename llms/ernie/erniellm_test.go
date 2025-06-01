package ernie

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

func TestNew(t *testing.T) {
	// Save and restore environment variables
	oldAPIKey := os.Getenv("ERNIE_API_KEY")
	oldSecretKey := os.Getenv("ERNIE_SECRET_KEY")
	defer func() {
		if oldAPIKey != "" {
			os.Setenv("ERNIE_API_KEY", oldAPIKey)
		} else {
			os.Unsetenv("ERNIE_API_KEY")
		}
		if oldSecretKey != "" {
			os.Setenv("ERNIE_SECRET_KEY", oldSecretKey)
		} else {
			os.Unsetenv("ERNIE_SECRET_KEY")
		}
	}()

	tests := []struct {
		name    string
		opts    []Option
		envVars map[string]string
		wantErr bool
		check   func(t *testing.T, llm *LLM)
	}{
		{
			name: "with environment variables",
			envVars: map[string]string{
				"ERNIE_API_KEY":    "test-api-key",
				"ERNIE_SECRET_KEY": "test-secret-key",
			},
			opts:    []Option{},
			wantErr: true, // Will fail with test credentials
		},
		{
			name: "with options",
			opts: []Option{
				WithAPIKey("custom-api-key"),
				WithSecretKey("custom-secret-key"),
			},
			wantErr: true, // Will fail with test credentials
		},
		{
			name: "with access token",
			opts: []Option{
				WithAccessToken("test-access-token"),
			},
			check: func(t *testing.T, llm *LLM) {
				assert.NotNil(t, llm)
			},
		},
		{
			name:    "without credentials",
			opts:    []Option{},
			wantErr: true,
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

			llm, err := New(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, llm)
				}
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithModel", func(t *testing.T) {
		opts := &options{}
		WithModel("ernie-bot-4")(opts)
		assert.Equal(t, ModelName("ernie-bot-4"), opts.modelName)
	})

	t.Run("WithAPIKey", func(t *testing.T) {
		opts := &options{}
		WithAPIKey("test-key")(opts)
		assert.Equal(t, "test-key", opts.apiKey)
	})

	t.Run("WithSecretKey", func(t *testing.T) {
		opts := &options{}
		WithSecretKey("test-secret")(opts)
		assert.Equal(t, "test-secret", opts.secretKey)
	})

	t.Run("WithAccessToken", func(t *testing.T) {
		opts := &options{}
		WithAccessToken("test-token")(opts)
		assert.Equal(t, "test-token", opts.accessToken)
	})

	t.Run("WithCacheType", func(t *testing.T) {
		opts := &options{}
		WithCacheType("memory")(opts)
		assert.Equal(t, "memory", opts.cacheType)
	})

	t.Run("WithModelPath", func(t *testing.T) {
		opts := &options{}
		WithModelPath("/custom/path")(opts)
		assert.Equal(t, "/custom/path", opts.modelPath)
	})

	t.Run("WithBaseURL", func(t *testing.T) {
		opts := &options{}
		WithBaseURL("https://custom.ernie.com")(opts)
		assert.Equal(t, "https://custom.ernie.com", opts.baseURL)
	})
}

func TestLLM_Call(t *testing.T) {
	// Skip if no credentials
	if os.Getenv("ERNIE_API_KEY") == "" || os.Getenv("ERNIE_SECRET_KEY") == "" {
		t.Skip("Skipping test, requires ERNIE API credentials")
	}

	llm, err := New()
	require.NoError(t, err)

	ctx := context.Background()
	result, err := llm.Call(ctx, "Hello, how are you?")
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestLLM_GenerateContent(t *testing.T) {
	// Skip if no credentials
	if os.Getenv("ERNIE_API_KEY") == "" || os.Getenv("ERNIE_SECRET_KEY") == "" {
		t.Skip("Skipping test, requires ERNIE API credentials")
	}

	llm, err := New()
	require.NoError(t, err)

	ctx := context.Background()
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	}

	response, err := llm.GenerateContent(ctx, messages)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Choices)
}

func TestLLM_CreateEmbedding(t *testing.T) {
	// Skip if no credentials
	if os.Getenv("ERNIE_API_KEY") == "" || os.Getenv("ERNIE_SECRET_KEY") == "" {
		t.Skip("Skipping test, requires ERNIE API credentials")
	}

	llm, err := New()
	require.NoError(t, err)

	ctx := context.Background()
	embeddings, err := llm.CreateEmbedding(ctx, []string{"hello world", "goodbye world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
}
