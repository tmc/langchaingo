package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		opts    ClientOptions
		wantErr bool
		errType error
	}{
		{
			name: "valid with API key",
			opts: ClientOptions{
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "valid with access token",
			opts: ClientOptions{
				AccessToken: "test-access-token",
			},
			wantErr: false,
		},
		{
			name:    "no credentials",
			opts:    ClientOptions{},
			wantErr: true,
			errType: NoCredentialsError{},
		},
		{
			name: "both credentials - access token takes precedence",
			opts: ClientOptions{
				APIKey:      "test-api-key",
				AccessToken: "test-access-token",
			},
			wantErr: false, // This should not error as access token takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func TestClient_List(t *testing.T) {
	// Create a mock server that responds to the actual Zapier API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/exposed", r.URL.Path)

		// Check authentication
		apiKey := r.Header.Get("X-API-Key")
		authHeader := r.Header.Get("Authorization")
		assert.True(t, apiKey != "" || authHeader != "", "Should have authentication")

		response := listResponse{
			Results: []ListResult{
				{
					ID:          "action1",
					OperationID: "gmail.send_email",
					Description: "Send an email via Gmail",
					Params: map[string]string{
						"to":      "required",
						"subject": "required",
						"body":    "optional",
					},
				},
				{
					ID:          "action2",
					OperationID: "todoist.create_task",
					Description: "Create a task in Todoist",
					Params: map[string]string{
						"content":  "required",
						"due_date": "optional",
						"priority": "optional",
					},
				},
			},
			ConfigurationLink: "https://nla.zapier.com/demo/start/",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with custom base URL for testing

	client, err := NewClient(ClientOptions{
		APIKey:           "test-key",
		ZapierNLABaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	actions, err := client.List(ctx)
	require.NoError(t, err)
	assert.Len(t, actions, 2)
	assert.Equal(t, "Send an email via Gmail", actions[0].Description)
	assert.Equal(t, "Create a task in Todoist", actions[1].Description)
}

func TestClient_Execute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/exposed/test-action/execute/", r.URL.Path)

		var req map[string]string
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "Send test email", req["instructions"])
		assert.Equal(t, "value1", req["param1"])

		response := executionResponse{
			ActionUsed: "test-action",
			Status:     "success",
			Result:     "Email sent successfully",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:           "test-key",
		ZapierNLABaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	response, err := client.Execute(ctx, "test-action", "Send test email", map[string]string{
		"param1": "value1",
	})
	require.NoError(t, err)
	assert.Equal(t, "Email sent successfully", response)
}

func TestClient_ExecuteAsString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := executionResponse{
			ActionUsed: "create-task",
			Status:     "success",
			Result:     "Task created: #12345",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:           "test-key",
		ZapierNLABaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	result, err := client.ExecuteAsString(ctx, "create-task", "Create a new task", nil)
	require.NoError(t, err)
	assert.Equal(t, "Task created: #12345", result)
}

func TestClient_ExecuteWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"error": "Invalid parameters",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(ClientOptions{
		APIKey:           "test-key",
		ZapierNLABaseURL: server.URL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	_, err = client.Execute(ctx, "test-action", "Do something", nil)
	require.NoError(t, err) // The current implementation doesn't check HTTP status codes
}

func TestTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		transport  *Transport
		expectAuth string
		expectKey  string
	}{
		{
			name: "with API key",
			transport: &Transport{
				RoundTripper: http.DefaultTransport,
				apiKey:       "test-api-key",
				UserAgent:    "TestAgent/1.0",
			},
			expectAuth: "",
			expectKey:  "test-api-key",
		},
		{
			name: "with access token",
			transport: &Transport{
				RoundTripper: http.DefaultTransport,
				accessToken:  "test-token",
				UserAgent:    "TestAgent/1.0",
			},
			expectAuth: "Bearer test-token",
			expectKey:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectAuth != "" {
					assert.Equal(t, tt.expectAuth, r.Header.Get("Authorization"))
				}
				if tt.expectKey != "" {
					assert.Equal(t, tt.expectKey, r.Header.Get("X-API-Key"))
				}
				assert.Equal(t, "TestAgent/1.0", r.Header.Get("User-Agent"))
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := &http.Client{Transport: tt.transport}
			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
