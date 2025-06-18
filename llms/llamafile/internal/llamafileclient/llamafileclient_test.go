package llamafileclient

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/internal/httprr"
)

func TestClient_Generate(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "LLAMAFILE_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &GenerateRequest{
		Prompt: "Hello, how are you?",
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    100,
		},
	}

	var response *GenerateResponse
	err = client.Generate(ctx, req, func(resp GenerateResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Response)
	assert.True(t, response.Done)
}

func TestClient_GenerateStream(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "LLAMAFILE_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := true
	req := &GenerateRequest{
		Prompt: "Count from 1 to 5",
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    50,
		},
	}

	var responses []GenerateResponse
	err = client.Generate(ctx, req, func(resp GenerateResponse) error {
		responses = append(responses, resp)
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, responses)
	assert.True(t, responses[len(responses)-1].Done)
}

func TestClient_GenerateChat(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "LLAMAFILE_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	stream := false
	req := &ChatRequest{
		Messages: []*Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
		},
		Stream: &stream,
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			NPredict:    50,
		},
	}

	var response *ChatResponse
	err = client.GenerateChat(ctx, req, func(resp ChatResponse) error {
		response = &resp
		return nil
	})
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Content)
}

func TestClient_CreateEmbedding(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "LLAMAFILE_HOST")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	baseURL := "http://localhost:8080"
	if envURL := os.Getenv("LLAMAFILE_HOST"); envURL != "" && rr.Recording() {
		baseURL = envURL
	}

	parsedURL, err := url.Parse(baseURL)
	require.NoError(t, err)

	client, err := NewClient(parsedURL, rr.Client())
	require.NoError(t, err)

	texts := []string{"Hello world", "How are you?"}
	resp, err := client.CreateEmbedding(ctx, texts)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Results, 2)
	assert.NotEmpty(t, resp.Results[0].Embedding)
	assert.NotEmpty(t, resp.Results[1].Embedding)
}

// Unit tests that don't require external dependencies

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		ourl      *url.URL
		ohttp     *http.Client
		wantError bool
	}{
		{
			name:  "with nil URL and nil client",
			ourl:  nil,
			ohttp: nil,
		},
		{
			name:  "with valid URL and nil client",
			ourl:  &url.URL{Scheme: "http", Host: "localhost:8080"},
			ohttp: nil,
		},
		{
			name:  "with nil URL and custom client",
			ourl:  nil,
			ohttp: &http.Client{},
		},
		{
			name:  "with valid URL and custom client",
			ourl:  &url.URL{Scheme: "http", Host: "localhost:8080"},
			ohttp: &http.Client{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variable for consistent test behavior
			oldEnv := os.Getenv("LLAMAFILE_HOST")
			os.Unsetenv("LLAMAFILE_HOST")
			defer func() {
				if oldEnv != "" {
					os.Setenv("LLAMAFILE_HOST", oldEnv)
				}
			}()

			client, err := NewClient(tt.ourl, tt.ohttp)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.NotNil(t, client.base)
			assert.NotNil(t, client.httpClient)
		})
	}
}

func TestNewClientWithEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name    string
		envHost string
	}{
		{
			name:    "with scheme in environment",
			envHost: "https://example.com:8080",
		},
		{
			name:    "without scheme in environment",
			envHost: "example.com:8080",
		},
		{
			name:    "with IPv6 address",
			envHost: "[::1]:8080",
		},
		{
			name:    "with IPv4 address",
			envHost: "192.168.1.1:8080",
		},
		{
			name:    "invalid host port format",
			envHost: "invalid_host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LLAMAFILE_HOST", tt.envHost)
			defer os.Unsetenv("LLAMAFILE_HOST")

			client, err := NewClient(nil, nil)
			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.NotNil(t, client.base)
		})
	}
}

func TestExtractJSONFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		want    string
	}{
		{
			name:    "empty input",
			input:   []byte(""),
			wantErr: true,
		},
		{
			name:    "simple JSON",
			input:   []byte(`data: {"test": "value"}`),
			wantErr: false,
			want:    `{"test": "value"}`,
		},
		{
			name:    "JSON without data prefix",
			input:   []byte(`{"test": "value"}`),
			wantErr: false,
			want:    `{"test": "value"}`,
		},
		{
			name:    "invalid JSON",
			input:   []byte(`data: {invalid json}`),
			wantErr: true,
		},
		{
			name:    "escaped JSON",
			input:   []byte(`data: "{\"test\": \"value\"}"`),
			wantErr: false,
			want:    `"{\"test\": \"value\"}"`,
		},
		{
			name:    "array JSON",
			input:   []byte(`data: [1, 2, 3]`),
			wantErr: false,
			want:    `[1, 2, 3]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractJSONFromBytes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(result))
		})
	}
}

func TestStatusError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      StatusError
		expected string
	}{
		{
			name: "with status and error message",
			err: StatusError{
				Status:       "400 Bad Request",
				ErrorMessage: "Invalid input",
			},
			expected: "400 Bad Request: Invalid input",
		},
		{
			name: "with status only",
			err: StatusError{
				Status: "500 Internal Server Error",
			},
			expected: "500 Internal Server Error",
		},
		{
			name: "with error message only",
			err: StatusError{
				ErrorMessage: "Connection timeout",
			},
			expected: "Connection timeout",
		},
		{
			name:     "empty error",
			err:      StatusError{},
			expected: "something went wrong, please see the ollama server logs for details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantErr    bool
	}{
		{
			name:       "success status code",
			statusCode: 200,
			body:       []byte("success"),
			wantErr:    false,
		},
		{
			name:       "success status code 201",
			statusCode: 201,
			body:       []byte("created"),
			wantErr:    false,
		},
		{
			name:       "client error with JSON",
			statusCode: 400,
			body:       []byte(`{"error": "Bad Request"}`),
			wantErr:    true,
		},
		{
			name:       "client error with plain text",
			statusCode: 404,
			body:       []byte("Not Found"),
			wantErr:    true,
		},
		{
			name:       "server error",
			statusCode: 500,
			body:       []byte("Internal Server Error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}
			err := checkError(resp, tt.body)
			if tt.wantErr {
				assert.Error(t, err)
				var statusErr StatusError
				assert.ErrorAs(t, err, &statusErr)
				assert.Equal(t, tt.statusCode, statusErr.StatusCode)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPrepareBuffer(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
		},
		{
			name: "valid struct",
			data: struct {
				Test string `json:"test"`
			}{Test: "value"},
			wantErr: false,
		},
		{
			name:    "string data",
			data:    "test string",
			wantErr: false,
		},
		{
			name:    "number data",
			data:    42,
			wantErr: false,
		},
		{
			name:    "map data",
			data:    map[string]string{"key": "value"},
			wantErr: false,
		},
		{
			name:    "unmarshalable data",
			data:    func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf, err := prepareBuffer(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, buf)
		})
	}
}

func TestSetRequestHeaders(t *testing.T) {
	req, err := http.NewRequest("POST", "http://example.com", nil)
	require.NoError(t, err)

	setRequestHeaders(req)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Equal(t, "application/x-ndjson", req.Header.Get("Accept"))
}

func TestProcessScan(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		statusCode int
		wantErr    bool
	}{
		{
			name:       "valid JSON",
			input:      []byte(`{"test": "value"}`),
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "JSON with error field",
			input:      []byte(`{"error": "something went wrong"}`),
			statusCode: 200,
			wantErr:    true,
		},
		{
			name:       "empty input",
			input:      []byte(""),
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "bad status code with error",
			input:      []byte(`{"error": "bad request"}`),
			statusCode: 400,
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			input:      []byte(`{invalid json}`),
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}

			called := false
			fn := func(data []byte) error {
				called = true
				return nil
			}

			err := processScan(tt.input, resp, fn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if len(tt.input) > 0 && tt.input[0] != '{' {
					// If input doesn't start with JSON, fn shouldn't be called
					assert.False(t, called)
				}
			}
		})
	}
}

func TestGenerateRequestStructure(t *testing.T) {
	req := &GenerateRequest{
		Prompt:   "test prompt",
		System:   "test system",
		Template: "test template",
		Context:  []int{1, 2, 3},
		Stream:   boolPtr(true),
		GenerationSettings: GenerationSettings{
			Temperature: 0.7,
			TopP:        0.9,
		},
	}

	assert.Equal(t, "test prompt", req.Prompt)
	assert.Equal(t, "test system", req.System)
	assert.Equal(t, "test template", req.Template)
	assert.Equal(t, []int{1, 2, 3}, req.Context)
	assert.NotNil(t, req.Stream)
	assert.True(t, *req.Stream)
	assert.Equal(t, 0.7, req.Temperature)
	assert.Equal(t, 0.9, req.TopP)
}

func TestChatRequestStructure(t *testing.T) {
	messages := []*Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	req := &ChatRequest{
		Messages: messages,
		Prompt:   stringPtr("test prompt"),
		Stream:   boolPtr(false),
		GenerationSettings: GenerationSettings{
			Model:       "test-model",
			Temperature: 0.5,
		},
	}

	assert.Len(t, req.Messages, 2)
	assert.Equal(t, "user", req.Messages[0].Role)
	assert.Equal(t, "Hello", req.Messages[0].Content)
	assert.NotNil(t, req.Prompt)
	assert.Equal(t, "test prompt", *req.Prompt)
	assert.NotNil(t, req.Stream)
	assert.False(t, *req.Stream)
	assert.Equal(t, "test-model", req.Model)
	assert.Equal(t, 0.5, req.Temperature)
}

func TestEmbeddingStructures(t *testing.T) {
	req := &EmbeddingRequest{
		Content: []string{"text1", "text2", "text3"},
	}
	assert.Len(t, req.Content, 3)
	assert.Equal(t, "text1", req.Content[0])

	embData := EmbeddingData{
		Embedding: []float32{0.1, 0.2, 0.3},
	}
	assert.Len(t, embData.Embedding, 3)
	assert.Equal(t, float32(0.1), embData.Embedding[0])

	resp := &EmbeddingResponse{
		Results: []EmbeddingData{embData},
	}
	assert.Len(t, resp.Results, 1)
	assert.Len(t, resp.Results[0].Embedding, 3)
}

func TestImageData(t *testing.T) {
	data := []byte("fake image data")
	imgData := ImageData(data)
	assert.Equal(t, data, []byte(imgData))
}

func TestGenerationSettings(t *testing.T) {
	settings := GenerationSettings{
		FrequencyPenalty:       0.5,
		Grammar:                "test-grammar",
		IgnoreEOS:              true,
		LogitBias:              []interface{}{1, 2, 3},
		MinP:                   0.05,
		Mirostat:               2,
		MirostatEta:            0.1,
		MirostatTau:            5.0,
		Model:                  "test-model",
		NCtx:                   2048,
		NKeep:                  10,
		NPredict:               100,
		NProbs:                 5,
		PenalizeNL:             true,
		PenaltyPromptTokens:    []interface{}{4, 5, 6},
		PresencePenalty:        0.6,
		RepeatLastN:            64,
		RepeatPenalty:          1.1,
		Seed:                   42,
		Stop:                   []string{"</s>", "stop"},
		Stream:                 true,
		Temperature:            0.8,
		TfsZ:                   1.0,
		TopK:                   40,
		TopP:                   0.95,
		TypicalP:               1.0,
		UsePenaltyPromptTokens: true,
		EmbeddingSize:          384,
	}

	assert.Equal(t, 0.5, settings.FrequencyPenalty)
	assert.Equal(t, "test-grammar", settings.Grammar)
	assert.True(t, settings.IgnoreEOS)
	assert.Len(t, settings.LogitBias, 3)
	assert.Equal(t, 0.05, settings.MinP)
	assert.Equal(t, 2, settings.Mirostat)
	assert.Equal(t, 0.1, settings.MirostatEta)
	assert.Equal(t, 5.0, settings.MirostatTau)
	assert.Equal(t, "test-model", settings.Model)
	assert.Equal(t, 2048, settings.NCtx)
	assert.Equal(t, 10, settings.NKeep)
	assert.Equal(t, 100, settings.NPredict)
	assert.Equal(t, 5, settings.NProbs)
	assert.True(t, settings.PenalizeNL)
	assert.Len(t, settings.PenaltyPromptTokens, 3)
	assert.Equal(t, 0.6, settings.PresencePenalty)
	assert.Equal(t, 64, settings.RepeatLastN)
	assert.Equal(t, 1.1, settings.RepeatPenalty)
	assert.Equal(t, uint32(42), settings.Seed)
	assert.Len(t, settings.Stop, 2)
	assert.True(t, settings.Stream)
	assert.Equal(t, 0.8, settings.Temperature)
	assert.Equal(t, 1.0, settings.TfsZ)
	assert.Equal(t, 40, settings.TopK)
	assert.Equal(t, 0.95, settings.TopP)
	assert.Equal(t, 1.0, settings.TypicalP)
	assert.True(t, settings.UsePenaltyPromptTokens)
	assert.Equal(t, 384, settings.EmbeddingSize)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}
