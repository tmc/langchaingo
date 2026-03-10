package openaiclient

import (
	"encoding/json"
	"testing"
)

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
				Temperature: getFloatPointer(0.7),
			},
			wantTemperature: true,
		},
		{
			name: "gpt-4.1-mini model omits zero temperature",
			request: ChatRequest{
				Model: "gpt-4.1-mini",
			},
			wantTemperature: false,
		},
		{
			name: "gpt-5 model omits zero temperature",
			request: ChatRequest{
				Model:       "gpt-5-mini",
				Temperature: getFloatPointer(0.0),
			},
			wantTemperature: true,
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
				var temperature float64
				if tt.request.Temperature != nil {
					temperature = *tt.request.Temperature
				}

				temp, ok := result["temperature"].(float64)
				if !ok {
					t.Errorf("temperature is not a float64: %T", result["temperature"])
				} else if temp != temperature {
					t.Errorf("temperature value: got %v, want %v", temp, temperature)
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

func getFloatPointer(f float64) *float64 {
	return &f
}

func getIntPointer(i int) *int {
	return &i
}
