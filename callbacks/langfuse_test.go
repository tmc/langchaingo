package callbacks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

func TestNewLangfuseHandler(t *testing.T) {
	tests := []struct {
		name    string
		opts    LangfuseOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: LangfuseOptions{
				PublicKey: "pk_test",
				SecretKey: "sk_test",
			},
			wantErr: false,
		},
		{
			name: "missing public key",
			opts: LangfuseOptions{
				SecretKey: "sk_test",
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			opts: LangfuseOptions{
				PublicKey: "pk_test",
			},
			wantErr: true,
		},
		{
			name: "custom base URL",
			opts: LangfuseOptions{
				BaseURL:   "https://custom.langfuse.com",
				PublicKey: "pk_test",
				SecretKey: "sk_test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := NewLangfuseHandler(tt.opts)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
				assert.NotEmpty(t, handler.traceID)
				
				if tt.opts.BaseURL != "" {
					assert.Equal(t, tt.opts.BaseURL, handler.baseURL)
				} else {
					assert.Equal(t, "https://cloud.langfuse.com", handler.baseURL)
				}
			}
		})
	}
}

func TestLangfuseHandler_LLMHandling(t *testing.T) {
	// Create a test server to capture API calls
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authentication
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "pk_test", username)
		assert.Equal(t, "sk_test", password)

		// Capture the payload
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		
		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		require.NoError(t, err)
		
		receivedPayloads = append(receivedPayloads, payload)
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
		UserID:    "test-user",
		SessionID: "test-session",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test LLM start
	prompts := []string{"What is the capital of France?"}
	handler.HandleLLMStart(ctx, prompts)

	// Test LLM generate content start
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	}
	handler.HandleLLMGenerateContentStart(ctx, messages)

	// Test LLM generate content end
	response := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:    "Paris is the capital of France.",
				StopReason: "stop",
				GenerationInfo: map[string]interface{}{
					"model": "gpt-3.5-turbo",
					"prompt_tokens": 10,
					"completion_tokens": 8,
					"total_tokens": 18,
				},
			},
		},
	}
	handler.HandleLLMGenerateContentEnd(ctx, response)

	// Wait a bit for async requests to complete
	time.Sleep(100 * time.Millisecond)

	// Verify we received the span
	assert.GreaterOrEqual(t, len(receivedPayloads), 1)
	
	// Check the latest payload (should be the completed span)
	lastPayload := receivedPayloads[len(receivedPayloads)-1]
	assert.Equal(t, "SPAN", lastPayload["type"])
	
	body, ok := lastPayload["body"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "llm-generation", body["name"])
	assert.Equal(t, handler.traceID, body["traceId"])
	assert.NotNil(t, body["input"])
	assert.NotNil(t, body["output"])
	assert.NotNil(t, body["usage"])
}

func TestLangfuseHandler_ChainHandling(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test chain handling
	inputs := map[string]any{"question": "What is AI?"}
	handler.HandleChainStart(ctx, inputs)

	outputs := map[string]any{"answer": "AI is artificial intelligence"}
	handler.HandleChainEnd(ctx, outputs)

	time.Sleep(100 * time.Millisecond)

	assert.GreaterOrEqual(t, len(receivedPayloads), 1)
	
	lastPayload := receivedPayloads[len(receivedPayloads)-1]
	body := lastPayload["body"].(map[string]interface{})
	assert.Equal(t, "chain", body["name"])
	assert.Equal(t, inputs, body["input"])
	assert.Equal(t, outputs, body["output"])
}

func TestLangfuseHandler_ToolHandling(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test tool handling
	input := "calculate 2+2"
	handler.HandleToolStart(ctx, input)

	output := "4"
	handler.HandleToolEnd(ctx, output)

	time.Sleep(100 * time.Millisecond)

	assert.GreaterOrEqual(t, len(receivedPayloads), 1)
	
	lastPayload := receivedPayloads[len(receivedPayloads)-1]
	body := lastPayload["body"].(map[string]interface{})
	assert.Equal(t, "tool", body["name"])
	assert.Equal(t, input, body["input"])
	assert.Equal(t, output, body["output"])
}

func TestLangfuseHandler_AgentHandling(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test agent action
	action := schema.AgentAction{
		Tool:      "calculator",
		ToolInput: "2+2",
		Log:       "I need to calculate 2+2",
	}
	handler.HandleAgentAction(ctx, action)

	// Test agent finish
	finish := schema.AgentFinish{
		ReturnValues: map[string]any{"result": "4"},
		Log:          "The calculation is complete",
	}
	handler.HandleAgentFinish(ctx, finish)

	time.Sleep(100 * time.Millisecond)

	assert.GreaterOrEqual(t, len(receivedPayloads), 2)
	
	// Find the agent action and agent finish payloads
	var actionPayload, finishPayload map[string]interface{}
	for _, payload := range receivedPayloads {
		body := payload["body"].(map[string]interface{})
		if name := body["name"].(string); name == "agent-action" {
			actionPayload = payload
		} else if name == "agent-finish" {
			finishPayload = payload
		}
	}
	
	// Check agent action payload
	require.NotNil(t, actionPayload, "agent-action payload not found")
	actionBody := actionPayload["body"].(map[string]interface{})
	assert.Equal(t, "agent-action", actionBody["name"])
	
	input := actionBody["input"].(map[string]interface{})
	assert.Equal(t, action.Tool, input["tool"])
	assert.Equal(t, action.ToolInput, input["toolInput"])
	assert.Equal(t, action.Log, input["log"])

	// Check agent finish payload
	require.NotNil(t, finishPayload, "agent-finish payload not found")
	finishBody := finishPayload["body"].(map[string]interface{})
	assert.Equal(t, "agent-finish", finishBody["name"])
	assert.Equal(t, finish.ReturnValues, finishBody["input"])
	assert.Equal(t, finish.Log, finishBody["output"])
}

func TestLangfuseHandler_RetrieverHandling(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test retriever handling
	query := "What is machine learning?"
	handler.HandleRetrieverStart(ctx, query)

	documents := []schema.Document{
		{PageContent: "Machine learning is a subset of AI"},
		{PageContent: "ML algorithms learn from data"},
	}
	handler.HandleRetrieverEnd(ctx, query, documents)

	time.Sleep(100 * time.Millisecond)

	assert.GreaterOrEqual(t, len(receivedPayloads), 1)
	
	lastPayload := receivedPayloads[len(receivedPayloads)-1]
	body := lastPayload["body"].(map[string]interface{})
	assert.Equal(t, "retriever", body["name"])
	assert.Equal(t, query, body["input"])
	
	output := body["output"].(map[string]interface{})
	assert.Equal(t, float64(len(documents)), output["count"])
}

func TestLangfuseHandler_ErrorHandling(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test LLM error
	handler.HandleLLMStart(ctx, []string{"test prompt"})
	testErr := fmt.Errorf("API rate limit exceeded")
	handler.HandleLLMError(ctx, testErr)

	// Test chain error
	handler.HandleChainStart(ctx, map[string]any{"input": "test"})
	handler.HandleChainError(ctx, testErr)

	// Test tool error
	handler.HandleToolStart(ctx, "test input")
	handler.HandleToolError(ctx, testErr)

	time.Sleep(100 * time.Millisecond)

	assert.GreaterOrEqual(t, len(receivedPayloads), 3)
	
	// Check that error spans have the correct level and status message
	for _, payload := range receivedPayloads {
		body := payload["body"].(map[string]interface{})
		if level, ok := body["level"]; ok && level == "ERROR" {
			assert.Equal(t, testErr.Error(), body["statusMessage"])
		}
	}
}

func TestLangfuseHandler_Flush(t *testing.T) {
	var receivedPayloads []map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		receivedPayloads = append(receivedPayloads, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "pk_test",
		SecretKey: "sk_test",
		UserID:    "test-user",
		SessionID: "test-session",
		Metadata:  map[string]interface{}{"version": "1.0"},
	})
	require.NoError(t, err)

	// Flush should send the trace
	err = handler.Flush()
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(receivedPayloads), 1)
	
	// Check that we received a trace
	tracePayload := receivedPayloads[0]
	assert.Equal(t, "TRACE_CREATE", tracePayload["type"])
	
	body := tracePayload["body"].(map[string]interface{})
	assert.Equal(t, handler.traceID, body["id"])
	assert.Equal(t, "test-user", body["userId"])
	assert.Equal(t, "test-session", body["sessionId"])
	
	metadata := body["metadata"].(map[string]interface{})
	assert.Equal(t, "1.0", metadata["version"])
}

func TestLangfuseHandler_SetTraceMetadata(t *testing.T) {
	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   "https://test.example.com",
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	// Test setting metadata
	metadata := map[string]interface{}{
		"environment": "test",
		"version":     "2.0",
	}
	handler.SetTraceMetadata(metadata)

	// Verify metadata was set
	handler.mu.RLock()
	trace := handler.traces[handler.traceID]
	handler.mu.RUnlock()

	require.NotNil(t, trace)
	assert.Equal(t, "test", trace.Metadata["environment"])
	assert.Equal(t, "2.0", trace.Metadata["version"])
}

func TestLangfuseHandler_GetTraceID(t *testing.T) {
	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   "https://test.example.com",
		PublicKey: "pk_test",
		SecretKey: "sk_test",
		TraceID:   "custom-trace-id",
	})
	require.NoError(t, err)

	assert.Equal(t, "custom-trace-id", handler.GetTraceID())
}

func TestLangfuseHandler_APIErrorHandling(t *testing.T) {
	// Create a server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   server.URL,
		PublicKey: "invalid_key",
		SecretKey: "invalid_secret",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// This should not panic even if the API call fails
	handler.HandleLLMStart(ctx, []string{"test"})
	handler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "test"}},
	})

	// Wait for async request to complete
	time.Sleep(100 * time.Millisecond)
	
	// Test should pass without panicking
}

func TestLangfuseHandler_ConcurrentAccess(t *testing.T) {
	handler, err := NewLangfuseHandler(LangfuseOptions{
		BaseURL:   "https://test.example.com",
		PublicKey: "pk_test",
		SecretKey: "sk_test",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			handler.HandleChainStart(ctx, map[string]any{"id": id})
			handler.HandleChainEnd(ctx, map[string]any{"result": id * 2})
			handler.SetTraceMetadata(map[string]interface{}{
				fmt.Sprintf("worker_%d", id): true,
			})
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test should complete without race condition errors
}