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

func TestChatRequest_WebSearchOptionsMarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		request ChatRequest
		want    map[string]interface{}
	}{
		{
			name: "no web search options",
			request: ChatRequest{
				Model: "gpt-4o-search-preview",
			},
			want: nil,
		},
		{
			name: "empty web search options",
			request: ChatRequest{
				Model:            "gpt-4o-search-preview",
				WebSearchOptions: &WebSearchOptions{},
			},
			want: map[string]interface{}{},
		},
		{
			name: "web search with search context size",
			request: ChatRequest{
				Model: "gpt-4o-search-preview",
				WebSearchOptions: &WebSearchOptions{
					SearchContextSize: "high",
				},
			},
			want: map[string]interface{}{
				"search_context_size": "high",
			},
		},
		{
			name: "web search with user location",
			request: ChatRequest{
				Model: "gpt-4o-search-preview",
				WebSearchOptions: &WebSearchOptions{
					SearchContextSize: "medium",
					UserLocation: &UserLocation{
						Type: "approximate",
						Approximate: &ApproximateLocation{
							Country: "US",
							City:    "San Francisco",
							Region:  "California",
						},
					},
				},
			},
			want: map[string]interface{}{
				"search_context_size": "medium",
				"user_location": map[string]interface{}{
					"type": "approximate",
					"approximate": map[string]interface{}{
						"country": "US",
						"city":    "San Francisco",
						"region":  "California",
					},
				},
			},
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

			webSearchOpts, hasWebSearch := result["web_search_options"]
			if tt.want == nil {
				if hasWebSearch {
					t.Errorf("expected no web_search_options, got %v", webSearchOpts)
				}
			} else {
				if !hasWebSearch {
					t.Fatal("expected web_search_options to be present")
				}
				// Check that it's properly serialized
				webSearchMap, ok := webSearchOpts.(map[string]interface{})
				if !ok {
					t.Fatalf("web_search_options is not a map: %T", webSearchOpts)
				}
				if tt.want["search_context_size"] != nil {
					if webSearchMap["search_context_size"] != tt.want["search_context_size"] {
						t.Errorf("search_context_size: got %v, want %v",
							webSearchMap["search_context_size"], tt.want["search_context_size"])
					}
				}
				if tt.want["user_location"] != nil {
					userLoc, ok := webSearchMap["user_location"].(map[string]interface{})
					if !ok {
						t.Fatalf("user_location is not a map: %T", webSearchMap["user_location"])
					}
					wantUserLoc := tt.want["user_location"].(map[string]interface{})
					if userLoc["type"] != wantUserLoc["type"] {
						t.Errorf("user_location.type: got %v, want %v", userLoc["type"], wantUserLoc["type"])
					}
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
		{"o30-mini", false},    // Doesn't start with "o3-"
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

func TestChatRequest_ExtraBodyMarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		request   ChatRequest
		wantExtra map[string]interface{}
	}{
		{
			name: "with extra_body",
			request: ChatRequest{
				Model: "gpt-4",
				ExtraBody: map[string]any{
					"parallel_tool_calls": false,
					"custom_param":        "test_value",
				},
			},
			wantExtra: map[string]interface{}{
				"parallel_tool_calls": false,
				"custom_param":        "test_value",
			},
		},
		{
			name: "without extra_body",
			request: ChatRequest{
				Model: "gpt-4",
			},
			wantExtra: nil,
		},
		{
			name: "empty extra_body",
			request: ChatRequest{
				Model:     "gpt-4",
				ExtraBody: map[string]any{},
			},
			wantExtra: nil,
		},
		{
			name: "extra_body with nested objects",
			request: ChatRequest{
				Model: "gpt-4",
				ExtraBody: map[string]any{
					"nested": map[string]interface{}{
						"key1": "value1",
						"key2": 123,
					},
				},
			},
			wantExtra: map[string]interface{}{
				"nested": map[string]interface{}{
					"key1": "value1",
					"key2": float64(123),
				},
			},
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

			if tt.wantExtra == nil {
				// Check that extra_body fields are not present
				for key := range tt.request.ExtraBody {
					if _, exists := result[key]; exists {
						t.Errorf("unexpected extra_body field %q in result", key)
					}
				}
			} else {
				// Check that all extra_body fields are present in result
				for key, wantValue := range tt.wantExtra {
					gotValue, exists := result[key]
					if !exists {
						t.Errorf("missing extra_body field %q in result", key)
						continue
					}

					// For nested maps, need deep comparison
					wantJSON, _ := json.Marshal(wantValue)
					gotJSON, _ := json.Marshal(gotValue)
					if string(wantJSON) != string(gotJSON) {
						t.Errorf("extra_body field %q: got %v, want %v", key, gotValue, wantValue)
					}
				}
			}

			// Verify model field is still present
			if result["model"] != tt.request.Model {
				t.Errorf("model field: got %v, want %v", result["model"], tt.request.Model)
			}
		})
	}
}

func TestChatRequest_ExtraBodyOverridesFields(t *testing.T) {
	// Test that ExtraBody can override standard fields
	request := ChatRequest{
		Model:       "gpt-4",
		Temperature: 0.7,
		ExtraBody: map[string]any{
			"temperature": 0.9,
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// ExtraBody should take precedence
	if temp := result["temperature"].(float64); temp != 0.9 {
		t.Errorf("temperature: got %v, want 0.9 (from ExtraBody)", temp)
	}
}
