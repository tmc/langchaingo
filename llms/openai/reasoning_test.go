package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
)

// TestReasoningStrengthMapping tests that reasoning strength is correctly mapped to reasoning_effort
func TestReasoningStrengthMapping(t *testing.T) {
	tests := []struct {
		name     string
		strength float64
		expected string
	}{
		{"Low effort (0.0)", 0.0, "low"},
		{"Low effort (0.33)", 0.33, "low"},
		{"Medium effort (0.34)", 0.34, "medium"},
		{"Medium effort (0.5)", 0.5, "medium"},
		{"Medium effort (0.66)", 0.66, "medium"},
		{"High effort (0.67)", 0.67, "high"},
		{"High effort (1.0)", 1.0, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server to capture the request
			var capturedRequest *openaiclient.ChatRequest
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Capture the request
				capturedRequest = &openaiclient.ChatRequest{}
				if err := json.NewDecoder(r.Body).Decode(capturedRequest); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}

				// Return a mock response
				response := openaiclient.ChatCompletionResponse{
					Choices: []*openaiclient.ChatCompletionChoice{
						{
							Message: openaiclient.ChatMessage{
								Role:    "assistant",
								Content: "Test response",
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			// Create client pointing to mock server
			llm, err := New(
				WithToken("test-token"),
				WithModel("o1"),
				WithBaseURL(server.URL),
			)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Make request with reasoning strength
			messages := []llms.MessageContent{
				llms.TextParts(llms.ChatMessageTypeHuman, "Solve this problem"),
			}

			_, err = llm.GenerateContent(context.Background(), messages,
				llms.WithReasoningStrength(tt.strength),
			)
			if err != nil {
				t.Fatalf("GenerateContent failed: %v", err)
			}

			// Verify reasoning_effort was set correctly
			if capturedRequest.ReasoningEffort != tt.expected {
				t.Errorf("expected reasoning_effort=%s, got %s", tt.expected, capturedRequest.ReasoningEffort)
			}
		})
	}
}

// TestReasoningContentExtraction tests that reasoning content and tokens are extracted
func TestReasoningContentExtraction(t *testing.T) {
	// Create a mock server with reasoning response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := openaiclient.ChatCompletionResponse{
			Choices: []*openaiclient.ChatCompletionChoice{
				{
					Message: openaiclient.ChatMessage{
						Role:             "assistant",
						Content:          "The answer is 42",
						ReasoningContent: "Let me think through this step by step...",
					},
				},
			},
			Usage: openaiclient.ChatUsage{
				PromptTokens:     100,
				CompletionTokens: 50,
				TotalTokens:      150,
			},
		}
		// Set reasoning tokens
		response.Usage.CompletionTokensDetails.ReasoningTokens = 30

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	llm, err := New(
		WithToken("test-token"),
		WithModel("o1"),
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Make request
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "What is the answer?"),
	}

	resp, err := llm.GenerateContent(context.Background(), messages,
		llms.WithReasoningStrength(0.8),
	)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	// Verify response content
	if len(resp.Choices) == 0 {
		t.Fatal("expected at least one choice")
	}

	choice := resp.Choices[0]
	if choice.Content != "The answer is 42" {
		t.Errorf("expected content='The answer is 42', got '%s'", choice.Content)
	}

	if choice.ReasoningContent != "Let me think through this step by step..." {
		t.Errorf("expected reasoning content to be set, got '%s'", choice.ReasoningContent)
	}

	// Verify GenerationInfo contains reasoning data
	if choice.GenerationInfo == nil {
		t.Fatal("expected GenerationInfo to be set")
	}

	// Check ReasoningTokens
	reasoningTokens, ok := choice.GenerationInfo["ReasoningTokens"].(int)
	if !ok {
		t.Error("expected ReasoningTokens in GenerationInfo")
	}
	if reasoningTokens != 30 {
		t.Errorf("expected ReasoningTokens=30, got %d", reasoningTokens)
	}

	// Check ReasoningContent
	reasoningContent, ok := choice.GenerationInfo["ReasoningContent"].(string)
	if !ok {
		t.Error("expected ReasoningContent in GenerationInfo")
	}
	if reasoningContent != "Let me think through this step by step..." {
		t.Errorf("expected ReasoningContent to be set, got '%s'", reasoningContent)
	}

	// Check standardized fields
	thinkingContent, ok := choice.GenerationInfo["ThinkingContent"].(string)
	if !ok {
		t.Error("expected ThinkingContent in GenerationInfo")
	}
	if thinkingContent != "Let me think through this step by step..." {
		t.Errorf("expected ThinkingContent to be set, got '%s'", thinkingContent)
	}

	thinkingTokens, ok := choice.GenerationInfo["ThinkingTokens"].(int)
	if !ok {
		t.Error("expected ThinkingTokens in GenerationInfo")
	}
	if thinkingTokens != 30 {
		t.Errorf("expected ThinkingTokens=30, got %d", thinkingTokens)
	}
}

// TestExtractReasoningUsage tests the ExtractReasoningUsage helper function
func TestExtractReasoningUsage(t *testing.T) {
	// Create mock GenerationInfo
	genInfo := map[string]any{
		"ReasoningTokens":  50,
		"ReasoningContent": "Thinking step by step",
		"ThinkingContent":  "Thinking step by step",
		"ThinkingTokens":   50,
	}

	// Extract reasoning usage
	usage := llms.ExtractReasoningUsage(genInfo)
	if usage == nil {
		t.Fatal("expected non-nil ReasoningUsage")
	}

	if usage.ReasoningTokens != 50 {
		t.Errorf("expected ReasoningTokens=50, got %d", usage.ReasoningTokens)
	}

	if usage.ReasoningContent != "Thinking step by step" {
		t.Errorf("expected ReasoningContent to be set, got '%s'", usage.ReasoningContent)
	}

	if usage.ThinkingContent != "Thinking step by step" {
		t.Errorf("expected ThinkingContent to be set, got '%s'", usage.ThinkingContent)
	}
}

// TestReasoningWithoutStrength tests that requests work without reasoning strength
func TestReasoningWithoutStrength(t *testing.T) {
	var capturedRequest *openaiclient.ChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = &openaiclient.ChatRequest{}
		if err := json.NewDecoder(r.Body).Decode(capturedRequest); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		response := openaiclient.ChatCompletionResponse{
			Choices: []*openaiclient.ChatCompletionChoice{
				{
					Message: openaiclient.ChatMessage{
						Role:    "assistant",
						Content: "Test response",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	llm, err := New(
		WithToken("test-token"),
		WithModel("o1"),
		WithBaseURL(server.URL),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Test"),
	}

	_, err = llm.GenerateContent(context.Background(), messages)
	if err != nil {
		t.Fatalf("GenerateContent failed: %v", err)
	}

	// Verify reasoning_effort is empty when not specified
	if capturedRequest.ReasoningEffort != "" {
		t.Errorf("expected empty reasoning_effort, got %s", capturedRequest.ReasoningEffort)
	}
}

// TestSupportsReasoning tests the SupportsReasoning method
func TestSupportsReasoning(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{"o1-preview", "o1-preview", true},
		{"o1-mini", "o1-mini", true},
		{"o3", "o3", true},
		{"o3-mini", "o3-mini", true},
		{"gpt-4", "gpt-4", false},
		{"gpt-4-turbo", "gpt-4-turbo", false},
		{"gpt-3.5-turbo", "gpt-3.5-turbo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := &LLM{model: tt.model}
			result := llm.SupportsReasoning()
			if result != tt.expected {
				t.Errorf("expected SupportsReasoning()=%v for model %s, got %v",
					tt.expected, tt.model, result)
			}
		})
	}
}
