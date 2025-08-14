package ernie

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
	"github.com/0xDezzy/langchaingo/llms"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			name: "missing required options",
			opts: []Option{
				WithAPIKey("test-key"), // Missing secret key
			},
			wantErr: true,
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

	t.Run("WithHTTPClient", func(t *testing.T) {
		opts := &options{}
		client := &http.Client{}
		WithHTTPClient(client)(opts)
		assert.Equal(t, client, opts.httpClient)
	})
}

func newErnieTestLLM(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	// Always check for recordings first - prefer recordings over environment variables
	if !hasExistingRecording(t) {
		t.Skip("No httprr recording available. Hint: Re-run tests with -httprecord=. to record new HTTP interactions")
	}

	// Use httputil.DefaultTransport - httprr handles wrapping
	rr := httprr.OpenForTest(t, httputil.DefaultTransport)

	// Scrub access token from recordings
	rr.ScrubReq(func(req *http.Request) error {
		q := req.URL.Query()
		if q.Get("access_token") != "" {
			q.Set("access_token", "test-access-token")
			req.URL.RawQuery = q.Encode()
		}
		return nil
	})

	// Create LLM with test credentials
	defaultOpts := []Option{
		WithAKSK("test-api-key", "test-secret-key"),
		WithHTTPClient(rr.Client()),
		WithModelName(ModelNameERNIEBot),
	}
	allOpts := append(defaultOpts, opts...)

	llm, err := New(allOpts...)
	require.NoError(t, err)
	return llm
}

// hasExistingRecording checks if a httprr recording exists for this test
func hasExistingRecording(t *testing.T) bool {
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	recordingPath := filepath.Join("testdata", testName+".httprr")
	_, err := os.Stat(recordingPath)
	return err == nil
}

func TestLLM_Call(t *testing.T) {
	llm := newErnieTestLLM(t)

	ctx := context.Background()
	result, err := llm.Call(ctx, "Hello, how are you?")
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestLLM_GenerateContent(t *testing.T) {
	llm := newErnieTestLLM(t)

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
	llm := newErnieTestLLM(t)

	ctx := context.Background()
	embeddings, err := llm.CreateEmbedding(ctx, []string{"hello world", "goodbye world"})
	require.NoError(t, err)
	assert.Len(t, embeddings, 2)
	assert.NotEmpty(t, embeddings[0])
	assert.NotEmpty(t, embeddings[1])
}
