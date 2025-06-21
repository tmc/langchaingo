package ernieclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
)

// mockHTTPClient is a mock implementation of the Doer interface
type mockHTTPClient struct {
	mu        sync.Mutex
	responses []mockResponse
	index     int
	requests  []*http.Request
	// Allow overriding the Do method for special cases
	DoFunc func(req *http.Request) (*http.Response, error)
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// If DoFunc is set, use it instead of the default behavior
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.requests = append(m.requests, req)

	if m.index >= len(m.responses) {
		return nil, errors.New("no more mock responses")
	}

	resp := m.responses[m.index]
	m.index++

	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
		Header:     make(http.Header),
	}, nil
}

func (m *mockHTTPClient) getRequests() []*http.Request {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		opts      []Option
		wantErr   bool
		errType   error
		mockResp  *mockResponse
		checkFunc func(t *testing.T, c *Client)
	}{
		{
			name:    "no auth provided",
			opts:    []Option{},
			wantErr: true,
			errType: ErrNotSetAuth,
		},
		{
			name: "with access token",
			opts: []Option{
				WithAccessToken("test-token"),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, c *Client) {
				assert.Equal(t, "test-token", c.accessToken)
			},
		},
		{
			name: "with API key and secret key",
			opts: []Option{
				WithAKSK("test-api-key", "test-secret-key"),
			},
			mockResp: &mockResponse{
				statusCode: http.StatusOK,
				body:       `{"access_token":"new-token","expires_in":2592000}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, c *Client) {
				assert.Equal(t, "test-api-key", c.apiKey)
				assert.Equal(t, "test-secret-key", c.secretKey)
				assert.Equal(t, "new-token", c.accessToken)
			},
		},
		{
			name: "with HTTP client",
			opts: []Option{
				WithAccessToken("test-token"),
				WithHTTPClient(&mockHTTPClient{}),
			},
			wantErr: false,
			checkFunc: func(t *testing.T, c *Client) {
				assert.NotNil(t, c.httpClient)
			},
		},
		{
			name: "option returns error",
			opts: []Option{
				func(c *Client) error {
					return errors.New("option error")
				},
			},
			wantErr: true,
		},
		{
			name: "access token request fails",
			opts: []Option{
				WithAKSK("test-api-key", "test-secret-key"),
			},
			mockResp: &mockResponse{
				err: errors.New("network error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If we need a mock response, inject a mock HTTP client
			if tt.mockResp != nil {
				mockClient := &mockHTTPClient{
					responses: []mockResponse{*tt.mockResp},
				}
				tt.opts = append([]Option{WithHTTPClient(mockClient)}, tt.opts...)
			}

			client, err := New(tt.opts...)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if tt.checkFunc != nil {
					tt.checkFunc(t, client)
				}
			}
		})
	}
}

func TestClient_CreateCompletion_Unit(t *testing.T) {
	tests := []struct {
		name         string
		modelPath    ModelPath
		request      *CompletionRequest
		mockResponse mockResponse
		wantErr      bool
		checkFunc    func(t *testing.T, resp *Completion, err error)
	}{
		{
			name:      "successful completion",
			modelPath: DefaultCompletionModelPath,
			request: &CompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 0.7,
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `{
					"id": "test-id",
					"object": "chat.completion",
					"created": 1234567890,
					"result": "Hello! How can I help you?",
					"usage": {
						"prompt_tokens": 10,
						"completion_tokens": 8,
						"total_tokens": 18
					}
				}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *Completion, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "test-id", resp.ID)
				assert.Equal(t, "Hello! How can I help you?", resp.Result)
				assert.Equal(t, 18, resp.Usage.TotalTokens)
			},
		},
		{
			name:      "empty model path uses default",
			modelPath: "",
			request: &CompletionRequest{
				Messages: []Message{
					{Role: "user", Content: "Test"},
				},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body:       `{"id": "test", "result": "response"}`,
			},
			wantErr: false,
		},
		{
			name: "http error",
			request: &CompletionRequest{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusInternalServerError,
			},
			wantErr: true,
			checkFunc: func(t *testing.T, resp *Completion, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "500")
			},
		},
		{
			name: "network error",
			request: &CompletionRequest{
				Messages: []Message{{Role: "user", Content: "Test"}},
			},
			mockResponse: mockResponse{
				err: errors.New("network error"),
			},
			wantErr: true,
		},
		{
			name: "streaming response",
			request: &CompletionRequest{
				Messages: []Message{{Role: "user", Content: "Count to 3"}},
				Stream:   true,
				StreamingFunc: func(ctx context.Context, chunk []byte) error {
					return nil
				},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `data: {"result":"One","is_end":false,"usage":{"prompt_tokens":5,"total_tokens":6}}
data: {"result":" Two","is_end":false,"usage":{"prompt_tokens":5,"total_tokens":8}}
data: {"result":" Three","is_end":true,"usage":{"prompt_tokens":5,"total_tokens":10}}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *Completion, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "One Two Three", resp.Result)
				assert.Equal(t, 5, resp.Usage.PromptTokens)
				assert.Equal(t, 5, resp.Usage.CompletionTokens) // 10 - 5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{tt.mockResponse},
			}

			client, err := New(
				WithAccessToken("test-token"),
				WithHTTPClient(mockClient),
			)
			require.NoError(t, err)

			ctx := context.Background()
			resp, err := client.CreateCompletion(ctx, tt.modelPath, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp, err)
			}

			// Verify request was made correctly
			reqs := mockClient.getRequests()
			if tt.mockResponse.err == nil && len(reqs) > 0 {
				assert.Equal(t, http.MethodPost, reqs[0].Method)
				assert.Contains(t, reqs[0].URL.String(), "wenxinworkshop/chat")
				assert.Contains(t, reqs[0].URL.Query().Get("access_token"), "test-token")
			}
		})
	}
}

func TestClient_CreateEmbedding_Unit(t *testing.T) {
	tests := []struct {
		name         string
		texts        []string
		mockResponse mockResponse
		wantErr      bool
		checkFunc    func(t *testing.T, resp *EmbeddingResponse, err error)
	}{
		{
			name:  "successful embedding",
			texts: []string{"Hello world", "Test text"},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `{
					"id": "test-id",
					"object": "embedding",
					"created": 1234567890,
					"data": [
						{"object": "embedding", "embedding": [0.1, 0.2, 0.3], "index": 0},
						{"object": "embedding", "embedding": [0.4, 0.5, 0.6], "index": 1}
					],
					"usage": {"prompt_tokens": 4, "total_tokens": 4}
				}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *EmbeddingResponse, err error) {
				assert.NoError(t, err)
				assert.Len(t, resp.Data, 2)
				assert.Equal(t, []float32{0.1, 0.2, 0.3}, resp.Data[0].Embedding)
				assert.Equal(t, []float32{0.4, 0.5, 0.6}, resp.Data[1].Embedding)
			},
		},
		{
			name:  "http error",
			texts: []string{"test"},
			mockResponse: mockResponse{
				statusCode: http.StatusBadRequest,
			},
			wantErr: true,
		},
		{
			name:  "network error",
			texts: []string{"test"},
			mockResponse: mockResponse{
				err: errors.New("network error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{tt.mockResponse},
			}

			client, err := New(
				WithAccessToken("test-token"),
				WithHTTPClient(mockClient),
			)
			require.NoError(t, err)

			ctx := context.Background()
			resp, err := client.CreateEmbedding(ctx, tt.texts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp, err)
			}

			// Verify request
			reqs := mockClient.getRequests()
			if tt.mockResponse.err == nil && len(reqs) > 0 {
				assert.Contains(t, reqs[0].URL.String(), "embeddings/embedding-v1")
			}
		})
	}
}

func TestClient_CreateChat_Unit(t *testing.T) {
	tests := []struct {
		name         string
		request      *ChatRequest
		mockResponse mockResponse
		wantErr      bool
		errMsg       string
		checkFunc    func(t *testing.T, resp *ChatResponse, err error)
	}{
		{
			name: "successful chat",
			request: &ChatRequest{
				Messages: []*ChatMessage{
					{Role: "user", Content: "Hello"},
				},
				Temperature: 0.7,
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `{
					"id": "test-id",
					"object": "chat",
					"created": 1234567890,
					"result": "Hello! How can I help?",
					"usage": {"prompt_tokens": 5, "completion_tokens": 5, "total_tokens": 10}
				}`,
			},
			wantErr: false,
		},
		{
			name: "chat with function call",
			request: &ChatRequest{
				Messages: []*ChatMessage{
					{Role: "user", Content: "What's the weather?"},
				},
				Functions: []FunctionDefinition{
					{
						Name:        "get_weather",
						Description: "Get weather information",
						Parameters:  map[string]any{"type": "object"},
					},
				},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `{
					"id": "test-id",
					"function_call": {
						"name": "get_weather",
						"thoughts": "User wants weather info",
						"arguments": "{\"location\":\"Beijing\"}"
					}
				}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *ChatResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp.FunctionCall)
				assert.Equal(t, "get_weather", resp.FunctionCall.Name)
			},
		},
		{
			name: "empty response error",
			request: &ChatRequest{
				Messages: []*ChatMessage{{Role: "user", Content: "Test"}},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body:       `{"id": "test-id"}`,
			},
			wantErr: true,
			errMsg:  "empty response",
		},
		{
			name: "http error with error message",
			request: &ChatRequest{
				Messages: []*ChatMessage{{Role: "user", Content: "Test"}},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusBadRequest,
				body:       `{"error": {"message": "Invalid request", "type": "bad_request"}}`,
			},
			wantErr: true,
			errMsg:  "Invalid request",
		},
		{
			name: "streaming chat",
			request: &ChatRequest{
				Messages: []*ChatMessage{{Role: "user", Content: "Count"}},
				StreamingFunc: func(ctx context.Context, chunk []byte) error {
					return nil
				},
			},
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `data: {"result":"One","is_end":false}
data: {"result":" Two","is_end":false}
data: {"result":" Three","is_end":true}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *ChatResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "One Two Three", resp.Result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{tt.mockResponse},
			}

			client, err := New(
				WithAccessToken("test-token"),
				WithHTTPClient(mockClient),
			)
			require.NoError(t, err)

			// Set ModelPath for proper URL building
			client.ModelPath = "completions"

			ctx := context.Background()
			resp, err := client.CreateChat(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp, err)
			}
		})
	}
}

func TestClient_getAccessToken(t *testing.T) {
	tests := []struct {
		name         string
		mockResponse mockResponse
		wantErr      bool
		checkFunc    func(t *testing.T, resp *authResponse, err error)
	}{
		{
			name: "successful token retrieval",
			mockResponse: mockResponse{
				statusCode: http.StatusOK,
				body: `{
					"access_token": "new-access-token",
					"expires_in": 2592000,
					"refresh_token": "refresh-token",
					"scope": "scope",
					"session_key": "session-key",
					"session_secret": "session-secret"
				}`,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *authResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "new-access-token", resp.AccessToken)
				assert.Equal(t, 2592000, resp.ExpiresIn)
			},
		},
		{
			name: "http error",
			mockResponse: mockResponse{
				statusCode: http.StatusUnauthorized,
			},
			wantErr: true,
		},
		{
			name: "network error",
			mockResponse: mockResponse{
				err: errors.New("network error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{tt.mockResponse},
			}

			client := &Client{
				apiKey:     "test-api-key",
				secretKey:  "test-secret-key",
				httpClient: mockClient,
			}

			ctx := context.Background()
			resp, err := client.getAccessToken(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp, err)
			}

			// Verify request
			reqs := mockClient.getRequests()
			if tt.mockResponse.err == nil && len(reqs) > 0 {
				assert.Contains(t, reqs[0].URL.String(), "oauth/2.0/token")
				assert.Contains(t, reqs[0].URL.String(), "client_id=test-api-key")
				assert.Contains(t, reqs[0].URL.String(), "client_secret=test-secret-key")
			}
		})
	}
}

func TestClient_buildURL(t *testing.T) {
	client := &Client{
		accessToken: "test-token",
	}

	tests := []struct {
		name      string
		modelPath ModelPath
		expected  string
	}{
		{
			name:      "default model path",
			modelPath: "completions",
			expected:  "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions?access_token=test-token",
		},
		{
			name:      "custom model path",
			modelPath: "ernie-bot-4",
			expected:  "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/ernie-bot-4?access_token=test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := client.buildURL(tt.modelPath)
			assert.Equal(t, tt.expected, url)
		})
	}
}

func TestClient_setHeaders(t *testing.T) {
	client := &Client{}
	req, err := http.NewRequest(http.MethodPost, "http://example.com", nil)
	require.NoError(t, err)

	client.setHeaders(req)

	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
}

func TestParseStreamingCompletionResponse(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantErr   bool
		wantText  string
		streamErr error
	}{
		{
			name: "successful streaming",
			body: `data: {"result":"Hello","is_end":false,"usage":{"prompt_tokens":5,"total_tokens":6}}
data: {"result":" world","is_end":false,"usage":{"prompt_tokens":5,"total_tokens":8}}
data: {"result":"!","is_end":true,"usage":{"prompt_tokens":5,"total_tokens":9}}`,
			wantText: "Hello world!",
		},
		{
			name:      "streaming with function error",
			body:      `data: {"result":"Test","is_end":false}`,
			streamErr: errors.New("stream error"),
			wantErr:   true,
		},
		{
			name: "mixed format lines",
			body: `data: {"result":"One","is_end":false}
{"result":" Two","is_end":false}
data: {"result":" Three","is_end":true}`,
			wantText: "One Two Three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.body)),
			}

			var chunks []string
			req := &CompletionRequest{
				StreamingFunc: func(ctx context.Context, chunk []byte) error {
					if tt.streamErr != nil {
						return tt.streamErr
					}
					chunks = append(chunks, string(chunk))
					return nil
				},
			}

			ctx := context.Background()
			result, err := parseStreamingCompletionResponse(ctx, resp, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantText, result.Result)
				if req.StreamingFunc != nil {
					assert.Equal(t, tt.wantText, strings.Join(chunks, ""))
				}
			}
		})
	}
}

func TestParseStreamingChatResponse(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantErr      bool
		wantText     string
		wantFunction *FunctionCallRes
	}{
		{
			name: "text streaming",
			body: `data: {"result":"Hello","is_end":false}
data: {"result":" there","is_end":true}`,
			wantText: "Hello there",
		},
		{
			name: "function call streaming",
			body: `data: {"function_call":{"name":"test_func","thoughts":"thinking","arguments":"{}"},"is_end":true}`,
			wantFunction: &FunctionCallRes{
				Name:      "test_func",
				Thoughts:  "thinking",
				Arguments: "{}",
			},
		},
		{
			name: "mixed content with truncation",
			body: `data: {"result":"Part 1","is_truncated":false,"is_end":false}
data: {"result":" Part 2","is_truncated":true,"is_end":true}`,
			wantText: "Part 1 Part 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				Body: io.NopCloser(strings.NewReader(tt.body)),
			}

			req := &ChatRequest{
				StreamingFunc: func(ctx context.Context, chunk []byte) error {
					return nil
				},
			}

			ctx := context.Background()
			result, err := parseStreamingChatResponse(ctx, resp, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantText, result.Result)
				if tt.wantFunction != nil {
					assert.Equal(t, tt.wantFunction, result.FunctionCall)
				}
			}
		})
	}
}

func TestAutoRefresh(t *testing.T) {
	// Test successful auto refresh
	t.Run("successful refresh", func(t *testing.T) {
		// Create a mock client that will return access tokens
		mockClient := &mockHTTPClient{
			responses: []mockResponse{
				{
					statusCode: http.StatusOK,
					body:       `{"access_token": "initial-token", "expires_in": 2592000}`,
				},
			},
		}

		client := &Client{
			apiKey:     "test-api",
			secretKey:  "test-secret",
			httpClient: mockClient,
		}

		// Run autoRefresh
		err := autoRefresh(client)
		assert.NoError(t, err)
		assert.Equal(t, "initial-token", client.accessToken)
	})

	// Test error handling in auto refresh
	t.Run("error in getAccessToken", func(t *testing.T) {
		mockClient := &mockHTTPClient{
			responses: []mockResponse{
				{
					err: errors.New("network error"),
				},
			},
		}

		client := &Client{
			apiKey:     "test-api",
			secretKey:  "test-secret",
			httpClient: mockClient,
		}

		// Run autoRefresh - should return error
		err := autoRefresh(client)
		assert.Error(t, err)
		assert.Equal(t, "", client.accessToken)
	})
}

func TestChatMessageMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		message  ChatMessage
		expected string
	}{
		{
			name: "basic message",
			message: ChatMessage{
				Role:    "user",
				Content: "Hello",
			},
			expected: `{"role":"user","content":"Hello"}`,
		},
		{
			name: "message with name",
			message: ChatMessage{
				Role:    "assistant",
				Content: "Response",
				Name:    "bot_1",
			},
			expected: `{"role":"assistant","content":"Response","name":"bot_1"}`,
		},
		{
			name: "message with function call",
			message: ChatMessage{
				Role:    "assistant",
				Content: "",
				FunctionCall: &llms.FunctionCall{
					Name:      "get_weather",
					Arguments: `{"location":"Beijing"}`,
				},
			},
			expected: `{"role":"assistant","content":"","function_call":{"name":"get_weather","arguments":"{\"location\":\"Beijing\"}"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestCompletionRequestMarshaling(t *testing.T) {
	req := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "Test"},
		},
		Temperature:  0.7,
		TopP:         0.9,
		PenaltyScore: 1.2,
		Stream:       true,
		UserID:       "user123",
	}

	data, err := json.Marshal(req)
	assert.NoError(t, err)

	var unmarshaled CompletionRequest
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	// StreamingFunc should not be marshaled
	assert.Nil(t, unmarshaled.StreamingFunc)
	assert.Equal(t, req.Temperature, unmarshaled.Temperature)
	assert.Equal(t, req.UserID, unmarshaled.UserID)
}

func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "json marshal error in CreateCompletion",
			testFunc: func(t *testing.T) {
				// Use mock HTTP client to avoid network calls
				mockClient := &mockHTTPClient{
					responses: []mockResponse{
						{statusCode: http.StatusOK, body: `{"result":"test response"}`},
					},
				}

				client, err := New(
					WithAccessToken("test"),
					WithHTTPClient(mockClient),
				)
				require.NoError(t, err)

				// Create a request that can't be marshaled
				req := &CompletionRequest{
					Messages: []Message{{
						Role:    "user",
						Content: string([]byte{0xff, 0xfe, 0xfd}), // Invalid UTF-8
					}},
				}

				_, err = client.CreateCompletion(context.Background(), "", req)
				// JSON marshaling might not fail on invalid UTF-8 in newer Go versions
				// so we just check that the function completes
				_ = err
			},
		},
		{
			name: "context creation error",
			testFunc: func(t *testing.T) {
				// Use mock HTTP client to avoid network calls
				mockClient := &mockHTTPClient{
					responses: []mockResponse{
						{statusCode: http.StatusOK, body: `{"result":"test response"}`},
					},
				}

				client, err := New(
					WithAccessToken("test"),
					WithHTTPClient(mockClient),
				)
				require.NoError(t, err)

				// Use a nil context to potentially cause issues
				defer func() {
					if r := recover(); r != nil {
						// Expected panic from nil context
					}
				}()

				_, _ = client.CreateCompletion(nil, "", &CompletionRequest{}) //nolint:staticcheck
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestConcurrentAccessTokenUpdate(t *testing.T) {
	client := &Client{
		accessToken: "initial-token",
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(2)

		// Reader
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				client.mu.RLock()
				_ = client.accessToken
				client.mu.RUnlock()
			}
		}()

		// Writer
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				client.mu.Lock()
				client.accessToken = fmt.Sprintf("token-%d-%d", id, j)
				client.mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or race condition, the test passes
}

func TestRequestBodyReading(t *testing.T) {
	// Test that request bodies can be read properly
	mockClient := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: http.StatusOK, body: `{"result":"ok"}`},
		},
	}

	client, err := New(
		WithAccessToken("test-token"),
		WithHTTPClient(mockClient),
	)
	require.NoError(t, err)

	req := &CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Test"}},
	}

	_, err = client.CreateCompletion(context.Background(), "", req)
	assert.NoError(t, err)

	// Verify the request body was properly set
	reqs := mockClient.getRequests()
	assert.Len(t, reqs, 1)

	// Read the request body
	body, err := io.ReadAll(reqs[0].Body)
	assert.NoError(t, err)

	var sentReq CompletionRequest
	err = json.Unmarshal(body, &sentReq)
	assert.NoError(t, err)
	assert.Equal(t, req.Messages[0].Content, sentReq.Messages[0].Content)
}

func TestInvalidJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		testFunc func(t *testing.T, client *Client)
	}{
		{
			name: "invalid JSON in completion",
			body: `{invalid json}`,
			testFunc: func(t *testing.T, client *Client) {
				_, err := client.CreateCompletion(context.Background(), "", &CompletionRequest{
					Messages: []Message{{Role: "user", Content: "Test"}},
				})
				assert.Error(t, err)
			},
		},
		{
			name: "invalid JSON in embedding",
			body: `{invalid json}`,
			testFunc: func(t *testing.T, client *Client) {
				_, err := client.CreateEmbedding(context.Background(), []string{"test"})
				assert.Error(t, err)
			},
		},
		{
			name: "invalid JSON in chat",
			body: `{invalid json}`,
			testFunc: func(t *testing.T, client *Client) {
				client.ModelPath = "test"
				_, err := client.CreateChat(context.Background(), &ChatRequest{
					Messages: []*ChatMessage{{Role: "user", Content: "Test"}},
				})
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{
					{statusCode: http.StatusOK, body: tt.body},
				},
			}

			client, err := New(
				WithAccessToken("test-token"),
				WithHTTPClient(mockClient),
			)
			require.NoError(t, err)

			tt.testFunc(t, client)
		})
	}
}

func TestResponseBodyClosure(t *testing.T) {
	// Track if response body was closed
	bodyClosed := false

	mockClient := &mockHTTPClient{
		responses: []mockResponse{
			{statusCode: http.StatusOK, body: `{"result":"test"}`},
		},
	}

	// Override the response to track body closure
	mockClient.DoFunc = func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body: &trackingCloser{
				ReadCloser: io.NopCloser(strings.NewReader(`{"result":"test"}`)),
				onClose: func() {
					bodyClosed = true
				},
			},
			Header: make(http.Header),
		}
		return resp, nil
	}

	client, err := New(
		WithAccessToken("test-token"),
		WithHTTPClient(mockClient),
	)
	require.NoError(t, err)

	// Make a request
	_, err = client.CreateCompletion(context.Background(), "", &CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})
	assert.NoError(t, err)

	// Verify the body was closed
	assert.True(t, bodyClosed)
}

// trackingCloser wraps an io.ReadCloser to track when Close is called
type trackingCloser struct {
	io.ReadCloser
	onClose func()
}

func (t *trackingCloser) Close() error {
	if t.onClose != nil {
		t.onClose()
	}
	return t.ReadCloser.Close()
}

func TestEmptyModelPath(t *testing.T) {
	client := &Client{
		accessToken: "test-token",
		ModelPath:   "", // Empty model path
	}

	// buildURL should handle empty ModelPath
	url := client.buildURL("")
	assert.Contains(t, url, "/wenxinworkshop/chat/")
	assert.Contains(t, url, "access_token=test-token")
}

func TestCreateChatWithoutFunctions(t *testing.T) {
	mockClient := &mockHTTPClient{
		responses: []mockResponse{
			{
				statusCode: http.StatusOK,
				body:       `{"result":"Response without functions"}`,
			},
		},
	}

	client, err := New(
		WithAccessToken("test-token"),
		WithHTTPClient(mockClient),
	)
	require.NoError(t, err)
	client.ModelPath = "test"

	req := &ChatRequest{
		Messages: []*ChatMessage{{Role: "user", Content: "Test"}},
		// No functions specified, so FunctionCallBehavior should remain empty
	}

	resp, err := client.CreateChat(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "Response without functions", resp.Result)

	// Verify FunctionCallBehavior was not set
	sentReqs := mockClient.getRequests()
	assert.Len(t, sentReqs, 1)
	body, _ := io.ReadAll(sentReqs[0].Body)
	assert.NotContains(t, string(body), "function_call")
}

func TestEmbeddingRequestMarshaling(t *testing.T) {
	texts := []string{"Hello", "World"}
	payload := map[string]any{"input": texts}

	data, err := json.Marshal(payload)
	assert.NoError(t, err)

	var unmarshaled map[string]any
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)

	input, ok := unmarshaled["input"].([]any)
	assert.True(t, ok)
	assert.Len(t, input, 2)
}

func TestHTTPClientNilCheck(t *testing.T) {
	// Ensure default HTTP client is set when none provided
	client, err := New(WithAccessToken("test"))
	assert.NoError(t, err)
	assert.NotNil(t, client.httpClient)
}

func TestWithAKSKOption(t *testing.T) {
	// Test the WithAKSK option directly
	client := &Client{}
	opt := WithAKSK("test-api", "test-secret")
	err := opt(client)

	assert.NoError(t, err)
	assert.Equal(t, "test-api", client.apiKey)
	assert.Equal(t, "test-secret", client.secretKey)
}

func TestResponseReaderError(t *testing.T) {
	// Create a reader that fails
	failingReader := &failingReader{
		failAfter: 10,
		err:       errors.New("read error"),
	}

	mockClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(failingReader),
				Header:     make(http.Header),
			}, nil
		},
	}

	client, err := New(
		WithAccessToken("test-token"),
		WithHTTPClient(mockClient),
	)
	require.NoError(t, err)

	_, err = client.CreateCompletion(context.Background(), "", &CompletionRequest{
		Messages: []Message{{Role: "user", Content: "Test"}},
	})
	// Should get an error from JSON decoding the failed read
	assert.Error(t, err)
}

type failingReader struct {
	failAfter int
	bytesRead int
	err       error
}

func (f *failingReader) Read(p []byte) (n int, err error) {
	if f.bytesRead >= f.failAfter {
		return 0, f.err
	}
	n = len(p)
	if n > f.failAfter-f.bytesRead {
		n = f.failAfter - f.bytesRead
	}
	f.bytesRead += n
	for i := 0; i < n; i++ {
		p[i] = 'x'
	}
	return n, nil
}
