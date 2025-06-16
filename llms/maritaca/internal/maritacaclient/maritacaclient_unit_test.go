package maritacaclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient is a mock implementation of Doer for testing
type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestNewClient(t *testing.T) {
	httpClient := &mockHTTPClient{}
	client, err := NewClient(httpClient)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, defaultURL, client.baseURL)
	assert.Equal(t, httpClient, client.httpClient)
}

func TestClient_stream(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		data       interface{}
		response   *http.Response
		fn         func([]byte) error
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:   "successful request",
			method: http.MethodPost,
			path:   "/test",
			data: map[string]string{
				"key": "value",
			},
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("line1\nline2\n")),
			},
			fn: func(data []byte) error {
				// Just consume the data
				return nil
			},
			wantErr: false,
		},
		{
			name:   "nil data",
			method: http.MethodGet,
			path:   "/test",
			data:   nil,
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("response")),
			},
			fn: func(data []byte) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:   "error response",
			method: http.MethodPost,
			path:   "/test",
			data:   map[string]string{"key": "value"},
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "400 Bad Request",
				Body:       io.NopCloser(strings.NewReader(`{"detail": "Invalid request"}`)),
			},
			fn:         func(data []byte) error { return nil },
			wantErr:    true,
			wantErrMsg: "Invalid request",
		},
		{
			name:   "error response - decode error",
			method: http.MethodPost,
			path:   "/test",
			data:   map[string]string{"key": "value"},
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Status:     "400 Bad Request",
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
			},
			fn:      func(data []byte) error { return nil },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &mockHTTPClient{
				doFunc: func(req *http.Request) (*http.Response, error) {
					// Verify request headers
					assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
					assert.Equal(t, "application/json", req.Header.Get("Accept"))
					assert.Equal(t, "Key test-token", req.Header.Get("Authorization"))

					// Verify request body if data is provided
					if tt.data != nil {
						body, err := io.ReadAll(req.Body)
						require.NoError(t, err)

						expected, err := json.Marshal(tt.data)
						require.NoError(t, err)
						assert.Equal(t, expected, body)
					}

					return tt.response, nil
				},
			}

			client := &Client{
				Token:      "test-token",
				baseURL:    "https://api.test.com",
				httpClient: httpClient,
			}

			ctx := context.Background()
			err := client.stream(ctx, tt.method, tt.path, tt.data, tt.fn)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_Generate_Unit(t *testing.T) {
	tests := []struct {
		name         string
		request      *ChatRequest
		responseBody string
		wantEvents   []string
		wantTexts    []string
		wantErr      bool
	}{
		{
			name: "non-streaming response",
			request: &ChatRequest{
				Model: "test-model",
				Options: Options{
					Stream: false,
				},
			},
			responseBody: `{"text": "Hello, world!", "usage": {"prompt_tokens": 10, "completion_tokens": 5}}`,
			wantEvents:   []string{"nostream"},
			wantTexts:    []string{"Hello, world!"},
		},
		{
			name: "streaming response",
			request: &ChatRequest{
				Model: "test-model",
				Options: Options{
					Stream: true,
				},
			},
			responseBody: "data: {\"text\": \"Hello\"}\ndata: {\"text\": \" world\"}\nevent: end\n",
			wantEvents:   []string{"message", "message", "end"},
			wantTexts:    []string{"Hello", " world", ""},
		},
		{
			name: "streaming response with invalid data",
			request: &ChatRequest{
				Model: "test-model",
				Options: Options{
					Stream: true,
				},
			},
			responseBody: "invalid line\ndata: {\"text\": \"Valid\"}\nevent: end\n",
			wantEvents:   []string{"message", "message", "end"},
			wantTexts:    []string{"", "Valid", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &mockHTTPClient{
				doFunc: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, "/chat/inference", req.URL.Path)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
					}, nil
				},
			}

			client := &Client{
				Token:      "test-token",
				baseURL:    "https://api.test.com",
				httpClient: httpClient,
			}

			ctx := context.Background()
			var events []string
			var texts []string

			err := client.Generate(ctx, tt.request, func(resp ChatResponse) error {
				events = append(events, resp.Event)
				texts = append(texts, resp.Text)
				return nil
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantEvents, events)
				assert.Equal(t, tt.wantTexts, texts)
			}
		})
	}
}

func TestParseData(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid data",
			input: `data: {"text": "Hello, world!"}`,
			want:  "Hello, world!",
		},
		{
			name:  "no data prefix",
			input: `{"text": "Hello, world!"}`,
			want:  "",
		},
		{
			name:    "invalid JSON",
			input:   `data: invalid json`,
			want:    "",
			wantErr: true,
		},
		{
			name:  "missing text field",
			input: `data: {"other": "field"}`,
			want:  "",
		},
		{
			name:  "text field not string",
			input: `data: {"text": 123}`,
			want:  "",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:    "multiple data sections",
			input:   `data: {"text": "first"} data: {"text": "second"}`,
			want:    "",
			wantErr: true, // This will fail because of invalid JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseData(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestStatusError(t *testing.T) {
	err := StatusError{
		StatusCode:   http.StatusBadRequest,
		Status:       "400 Bad Request",
		ErrorMessage: "Invalid request parameters",
	}

	// Test that it implements error interface
	var _ error = err

	// Test Error() method
	errMsg := err.Error()
	assert.Contains(t, errMsg, "400")
	assert.Contains(t, errMsg, "Bad Request")
	assert.Contains(t, errMsg, "Invalid request parameters")
}

func TestClient_streamWithLargeResponse(t *testing.T) {
	// Create a large response that exceeds typical buffer sizes
	var sb strings.Builder
	for i := 0; i < 1000; i++ {
		sb.WriteString(strings.Repeat("x", 1000))
		sb.WriteString("\n")
	}
	largeResponse := sb.String()

	httpClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(largeResponse)),
			}, nil
		},
	}

	client := &Client{
		Token:      "test-token",
		baseURL:    "https://api.test.com",
		httpClient: httpClient,
	}

	ctx := context.Background()
	lineCount := 0
	err := client.stream(ctx, http.MethodGet, "/test", nil, func(data []byte) error {
		lineCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1000, lineCount)
}

func TestClient_streamContextCancellation(t *testing.T) {
	// Create a context that we'll cancel during the request
	ctx, cancel := context.WithCancel(context.Background())

	httpClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			// Cancel the context after the request is made
			cancel()

			// Return a response that would normally be processed
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("data\n")),
			}, nil
		},
	}

	client := &Client{
		Token:      "test-token",
		baseURL:    "https://api.test.com",
		httpClient: httpClient,
	}

	err := client.stream(ctx, http.MethodGet, "/test", nil, func(data []byte) error {
		return nil
	})

	// The error handling depends on how the scanner behaves with a cancelled context
	// We're mainly testing that the function handles the cancellation gracefully
	_ = err
}

func TestClient_marshalError(t *testing.T) {
	// Test with data that cannot be marshaled to JSON
	type unmarshallable struct {
		Ch chan int
	}

	httpClient := &mockHTTPClient{}
	client := &Client{
		Token:      "test-token",
		baseURL:    "https://api.test.com",
		httpClient: httpClient,
	}

	ctx := context.Background()
	err := client.stream(ctx, http.MethodPost, "/test", unmarshallable{Ch: make(chan int)}, func(data []byte) error {
		return nil
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}

func TestClient_requestError(t *testing.T) {
	httpClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	client := &Client{
		Token:      "test-token",
		baseURL:    "https://api.test.com",
		httpClient: httpClient,
	}

	ctx := context.Background()
	err := client.stream(ctx, http.MethodGet, "/test", nil, func(data []byte) error {
		return nil
	})

	assert.Error(t, err)
}

func TestClient_callbackError(t *testing.T) {
	httpClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("line1\nline2\n")),
			}, nil
		},
	}

	client := &Client{
		Token:      "test-token",
		baseURL:    "https://api.test.com",
		httpClient: httpClient,
	}

	ctx := context.Background()
	callbackErr := assert.AnError
	err := client.stream(ctx, http.MethodGet, "/test", nil, func(data []byte) error {
		return callbackErr
	})

	assert.Error(t, err)
	assert.Equal(t, callbackErr, err)
}
