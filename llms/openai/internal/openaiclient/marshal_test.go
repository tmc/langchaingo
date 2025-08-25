package openaiclient

import (
	"encoding/json"
	"testing"
)

func TestChatRequest_MarshalJSON(t *testing.T) {
	tests := []struct {
		name                    string
		request                 ChatRequest
		wantMaxTokens           bool
		wantMaxCompletionTokens bool
	}{
		{
			name: "only MaxCompletionTokens set",
			request: ChatRequest{
				Model:               "gpt-4",
				MaxCompletionTokens: 100,
			},
			wantMaxTokens:           false,
			wantMaxCompletionTokens: true,
		},
		{
			name: "only MaxTokens set",
			request: ChatRequest{
				Model:     "gpt-4",
				MaxTokens: 200,
			},
			wantMaxTokens:           true,
			wantMaxCompletionTokens: false,
		},
		{
			name: "both set - only MaxCompletionTokens sent",
			request: ChatRequest{
				Model:               "gpt-4",
				MaxTokens:           300,
				MaxCompletionTokens: 400,
			},
			wantMaxTokens:           false,
			wantMaxCompletionTokens: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			hasMaxTokens := result["max_tokens"] != nil
			hasMaxCompletionTokens := result["max_completion_tokens"] != nil

			if hasMaxTokens != tt.wantMaxTokens {
				t.Errorf("max_tokens presence: got %v, want %v", hasMaxTokens, tt.wantMaxTokens)
			}
			if hasMaxCompletionTokens != tt.wantMaxCompletionTokens {
				t.Errorf("max_completion_tokens presence: got %v, want %v", hasMaxCompletionTokens, tt.wantMaxCompletionTokens)
			}

			// Never both
			if hasMaxTokens && hasMaxCompletionTokens {
				t.Error("Both max_tokens and max_completion_tokens are present - OpenAI API will reject!")
			}
		})
	}
}

func TestChatRequest_TemperatureMarshalJSON(t *testing.T) {
	tests := []struct {
		name            string
		request         ChatRequest
		wantTemperature bool
	}{
		{
			name: "regular model with temperature",
			request: ChatRequest{
				Model:       "gpt-4",
				Temperature: 0.7,
			},
			wantTemperature: true,
		},
		{
			name: "regular model with zero temperature",
			request: ChatRequest{
				Model:       "gpt-3.5-turbo",
				Temperature: 0.0,
			},
			wantTemperature: true,
		},
		{
			name: "gpt-5 model omits temperature",
			request: ChatRequest{
				Model:       "gpt-5-preview",
				Temperature: 0.7,
			},
			wantTemperature: false,
		},
		{
			name: "gpt-5 model omits zero temperature",
			request: ChatRequest{
				Model:       "gpt-5-mini",
				Temperature: 0.0,
			},
			wantTemperature: false,
		},
		{
			name: "o1 model omits temperature",
			request: ChatRequest{
				Model:       "o1-preview",
				Temperature: 0.5,
			},
			wantTemperature: false,
		},
		{
			name: "o1-mini model omits temperature",
			request: ChatRequest{
				Model:       "o1-mini",
				Temperature: 1.0,
			},
			wantTemperature: false,
		},
		{
			name: "o3 model omits temperature",
			request: ChatRequest{
				Model:       "o3-mini",
				Temperature: 0.8,
			},
			wantTemperature: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			hasTemperature := result["temperature"] != nil

			if hasTemperature != tt.wantTemperature {
				t.Errorf("temperature presence: got %v, want %v, JSON: %s", hasTemperature, tt.wantTemperature, string(data))
			}

			// If temperature should be present, verify the value
			if hasTemperature && tt.wantTemperature {
				temp, ok := result["temperature"].(float64)
				if !ok {
					t.Errorf("temperature is not a float64: %T", result["temperature"])
				} else if temp != tt.request.Temperature {
					t.Errorf("temperature value: got %v, want %v", temp, tt.request.Temperature)
				}
			}
		})
	}
}

func TestIsReasoningModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		// Regular models - should not be reasoning models
		{"gpt-4", false},
		{"gpt-3.5-turbo", false},
		{"gpt-4-turbo", false},
		{"gpt-4o", false},
		{"text-davinci-003", false},
		
		// GPT-5 models - should be reasoning models
		{"gpt-5", true},
		{"gpt-5-preview", true},
		{"gpt-5-mini", true},
		{"gpt-5-turbo", true},
		
		// o1 models - should be reasoning models
		{"o1-preview", true},
		{"o1-mini", true},
		{"o1-large", true},
		
		// o3 models - should be reasoning models
		{"o3", true}, // Base o3 model
		{"o3-mini", true},
		{"o3-preview", true},
		{"o3-large", true},
		
		// Edge cases
		{"", false},
		{"o10-preview", false}, // Doesn't start with "o1-"
		{"o30-mini", false}, // Doesn't start with "o3-"
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := isReasoningModel(tt.model)
			if result != tt.expected {
				t.Errorf("isReasoningModel(%q) = %v, want %v", tt.model, result, tt.expected)
			}
		})
	}
}
