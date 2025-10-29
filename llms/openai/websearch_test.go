package openai

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestNewWebSearchTool(t *testing.T) {
	tests := []struct {
		name        string
		contextSize WebSearchContextSize
		wantType    string
		wantSize    string
	}{
		{
			name:        "high context size",
			contextSize: WebSearchContextSizeHigh,
			wantType:    "web_search",
			wantSize:    "high",
		},
		{
			name:        "medium context size",
			contextSize: WebSearchContextSizeMedium,
			wantType:    "web_search",
			wantSize:    "medium",
		},
		{
			name:        "low context size",
			contextSize: WebSearchContextSizeLow,
			wantType:    "web_search",
			wantSize:    "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWebSearchTool(tt.contextSize)

			if tool.Type != tt.wantType {
				t.Errorf("NewWebSearchTool().Type = %v, want %v", tool.Type, tt.wantType)
			}

			if tool.WebSearch == nil {
				t.Fatal("NewWebSearchTool().WebSearch is nil")
			}

			if tool.WebSearch.SearchContextSize != tt.wantSize {
				t.Errorf("NewWebSearchTool().WebSearch.SearchContextSize = %v, want %v",
					tool.WebSearch.SearchContextSize, tt.wantSize)
			}

			// Function should be nil for web_search type
			if tool.Function != nil {
				t.Error("NewWebSearchTool().Function should be nil for web_search type")
			}
		})
	}
}

func TestWebSearchContextSizeConstants(t *testing.T) {
	// Verify the constant values match OpenAI's API expectations
	if WebSearchContextSizeHigh != "high" {
		t.Errorf("WebSearchContextSizeHigh = %v, want 'high'", WebSearchContextSizeHigh)
	}
	if WebSearchContextSizeMedium != "medium" {
		t.Errorf("WebSearchContextSizeMedium = %v, want 'medium'", WebSearchContextSizeMedium)
	}
	if WebSearchContextSizeLow != "low" {
		t.Errorf("WebSearchContextSizeLow = %v, want 'low'", WebSearchContextSizeLow)
	}
}

func TestWebSearchToolIntegration(t *testing.T) {
	// Test that the tool can be used in a tools array
	tool := NewWebSearchTool(WebSearchContextSizeMedium)
	tools := []llms.Tool{tool}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0].Type != "web_search" {
		t.Errorf("tool type = %v, want 'web_search'", tools[0].Type)
	}
}
