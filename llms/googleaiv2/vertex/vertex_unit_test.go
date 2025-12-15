package vertex

import (
	"encoding/json"
	"strings"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/vendasta/langchaingo/llms"
)

func TestConvertToolSchemaType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected genai.Type
	}{
		{"object type", "object", genai.TypeObject},
		{"string type", "string", genai.TypeString},
		{"number type", "number", genai.TypeNumber},
		{"integer type", "integer", genai.TypeInteger},
		{"boolean type", "boolean", genai.TypeBoolean},
		{"array type", "array", genai.TypeArray},
		{"unknown type", "unknown", genai.TypeUnspecified},
		{"empty type", "", genai.TypeUnspecified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToolSchemaType(tt.input)
			if result != tt.expected {
				t.Errorf("convertToolSchemaType(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertParts(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	tests := []struct {
		name    string
		parts   []llms.ContentPart
		wantErr bool
		check   func(t *testing.T, result []genai.Part)
	}{
		{
			name: "text content",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello, world!"},
			},
			wantErr: false,
			check: func(t *testing.T, result []genai.Part) {
				if len(result) != 1 {
					t.Fatalf("expected 1 part, got %d", len(result))
				}
				text, ok := result[0].(genai.Text)
				if !ok {
					t.Fatalf("expected genai.Text, got %T", result[0])
				}
				if string(text) != "Hello, world!" {
					t.Errorf("expected text 'Hello, world!', got %q", text)
				}
			},
		},
		{
			name: "binary content",
			parts: []llms.ContentPart{
				llms.BinaryContent{
					MIMEType: "image/png",
					Data:     []byte{0x89, 0x50, 0x4E, 0x47},
				},
			},
			wantErr: false,
			check: func(t *testing.T, result []genai.Part) {
				if len(result) != 1 {
					t.Fatalf("expected 1 part, got %d", len(result))
				}
				blob, ok := result[0].(genai.Blob)
				if !ok {
					t.Fatalf("expected genai.Blob, got %T", result[0])
				}
				if blob.MIMEType != "image/png" {
					t.Errorf("expected MIME type 'image/png', got %q", blob.MIMEType)
				}
				if len(blob.Data) != 4 {
					t.Errorf("expected data length 4, got %d", len(blob.Data))
				}
			},
		},
		{
			name: "tool call",
			parts: []llms.ContentPart{
				llms.ToolCall{
					FunctionCall: &llms.FunctionCall{
						Name:      "test_function",
						Arguments: `{"arg1": "value1", "arg2": 42}`,
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, result []genai.Part) {
				if len(result) != 1 {
					t.Fatalf("expected 1 part, got %d", len(result))
				}
				fc, ok := result[0].(genai.FunctionCall)
				if !ok {
					t.Fatalf("expected genai.FunctionCall, got %T", result[0])
				}
				if fc.Name != "test_function" {
					t.Errorf("expected function name 'test_function', got %q", fc.Name)
				}
				if fc.Args["arg1"] != "value1" {
					t.Errorf("expected arg1='value1', got %v", fc.Args["arg1"])
				}
				if fc.Args["arg2"] != float64(42) { // JSON unmarshals numbers as float64
					t.Errorf("expected arg2=42, got %v", fc.Args["arg2"])
				}
			},
		},
		{
			name: "tool call response",
			parts: []llms.ContentPart{
				llms.ToolCallResponse{
					Name:    "test_function",
					Content: "Function executed successfully",
				},
			},
			wantErr: false,
			check: func(t *testing.T, result []genai.Part) {
				if len(result) != 1 {
					t.Fatalf("expected 1 part, got %d", len(result))
				}
				fr, ok := result[0].(genai.FunctionResponse)
				if !ok {
					t.Fatalf("expected genai.FunctionResponse, got %T", result[0])
				}
				if fr.Name != "test_function" {
					t.Errorf("expected function name 'test_function', got %q", fr.Name)
				}
				response, ok := fr.Response["response"].(string)
				if !ok {
					t.Fatalf("expected response string, got %T", fr.Response["response"])
				}
				if response != "Function executed successfully" {
					t.Errorf("expected response content, got %q", response)
				}
			},
		},
		{
			name: "invalid tool call JSON",
			parts: []llms.ContentPart{
				llms.ToolCall{
					FunctionCall: &llms.FunctionCall{
						Name:      "test_function",
						Arguments: `invalid json`,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple mixed parts",
			parts: []llms.ContentPart{
				llms.TextContent{Text: "First part"},
				llms.TextContent{Text: "Second part"},
				llms.BinaryContent{MIMEType: "image/jpeg", Data: []byte{0xFF, 0xD8}},
			},
			wantErr: false,
			check: func(t *testing.T, result []genai.Part) {
				if len(result) != 3 {
					t.Fatalf("expected 3 parts, got %d", len(result))
				}
				if text, ok := result[0].(genai.Text); !ok || string(text) != "First part" {
					t.Errorf("expected first part to be 'First part', got %v", result[0])
				}
				if text, ok := result[1].(genai.Text); !ok || string(text) != "Second part" {
					t.Errorf("expected second part to be 'Second part', got %v", result[1])
				}
				if blob, ok := result[2].(genai.Blob); !ok || blob.MIMEType != "image/jpeg" {
					t.Errorf("expected third part to be image/jpeg blob, got %v", result[2])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertParts(tt.parts)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertParts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestConvertContent(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	tests := []struct {
		name         string
		content      llms.MessageContent
		wantErr      bool
		expectedRole string
	}{
		{
			name: "system message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeSystem,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "You are a helpful assistant"},
				},
			},
			expectedRole: RoleSystem,
		},
		{
			name: "AI message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeAI,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "I'm here to help"},
				},
			},
			expectedRole: RoleModel,
		},
		{
			name: "human message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Hello"},
				},
			},
			expectedRole: RoleUser,
		},
		{
			name: "generic message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeGeneric,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Generic message"},
				},
			},
			expectedRole: RoleUser,
		},
		{
			name: "tool message",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Tool response"},
				},
			},
			expectedRole: RoleUser,
		},
		{
			name: "unsupported function role",
			content: llms.MessageContent{
				Role: llms.ChatMessageTypeFunction,
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Function message"},
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported custom role",
			content: llms.MessageContent{
				Role: llms.ChatMessageType("custom"),
				Parts: []llms.ContentPart{
					llms.TextContent{Text: "Custom message"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result.Role != tt.expectedRole {
					t.Errorf("expected role %q, got %q", tt.expectedRole, result.Role)
				}
				if len(result.Parts) != len(tt.content.Parts) {
					t.Errorf("expected %d parts, got %d", len(tt.content.Parts), len(result.Parts))
				}
			}
		})
	}
}

func TestConvertCandidates(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	tests := []struct {
		name       string
		candidates []*genai.Candidate
		usage      *genai.UsageMetadata
		wantErr    bool
		check      func(t *testing.T, result *llms.ContentResponse)
	}{
		{
			name: "single text candidate",
			candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("Response text"),
						},
					},
					FinishReason: genai.FinishReasonStop,
				},
			},
			check: func(t *testing.T, result *llms.ContentResponse) {
				if len(result.Choices) != 1 {
					t.Fatalf("expected 1 choice, got %d", len(result.Choices))
				}
				if result.Choices[0].Content != "Response text" {
					t.Errorf("expected content 'Response text', got %q", result.Choices[0].Content)
				}
				// The FinishReason.String() method returns the full enum name
				if result.Choices[0].StopReason != "FinishReasonStop" {
					t.Errorf("expected stop reason 'FinishReasonStop', got %q", result.Choices[0].StopReason)
				}
			},
		},
		{
			name: "multiple text parts",
			candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("Part 1"),
							genai.Text(" Part 2"),
						},
					},
				},
			},
			check: func(t *testing.T, result *llms.ContentResponse) {
				if result.Choices[0].Content != "Part 1 Part 2" {
					t.Errorf("expected concatenated content, got %q", result.Choices[0].Content)
				}
			},
		},
		{
			name: "function call candidate",
			candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.FunctionCall{
								Name: "test_function",
								Args: map[string]any{"x": 1, "y": 2},
							},
						},
					},
				},
			},
			check: func(t *testing.T, result *llms.ContentResponse) {
				if len(result.Choices[0].ToolCalls) != 1 {
					t.Fatalf("expected 1 tool call, got %d", len(result.Choices[0].ToolCalls))
				}
				tc := result.Choices[0].ToolCalls[0]
				if tc.FunctionCall.Name != "test_function" {
					t.Errorf("expected function name 'test_function', got %q", tc.FunctionCall.Name)
				}
				var args map[string]any
				if err := json.Unmarshal([]byte(tc.FunctionCall.Arguments), &args); err != nil {
					t.Fatalf("failed to unmarshal arguments: %v", err)
				}
				if args["x"] != float64(1) || args["y"] != float64(2) {
					t.Errorf("expected args {x:1, y:2}, got %v", args)
				}
			},
		},
		{
			name: "with usage metadata",
			candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{genai.Text("Response")},
					},
				},
			},
			usage: &genai.UsageMetadata{
				PromptTokenCount:     10,
				CandidatesTokenCount: 5,
				TotalTokenCount:      15,
			},
			check: func(t *testing.T, result *llms.ContentResponse) {
				metadata := result.Choices[0].GenerationInfo
				if metadata["input_tokens"] != int32(10) {
					t.Errorf("expected input_tokens=10, got %v", metadata["input_tokens"])
				}
				if metadata["output_tokens"] != int32(5) {
					t.Errorf("expected output_tokens=5, got %v", metadata["output_tokens"])
				}
				if metadata["total_tokens"] != int32(15) {
					t.Errorf("expected total_tokens=15, got %v", metadata["total_tokens"])
				}
			},
		},
		{
			name: "with safety ratings and citations",
			candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{genai.Text("Safe response")},
					},
					SafetyRatings: []*genai.SafetyRating{
						{Category: genai.HarmCategoryHateSpeech, Probability: genai.HarmProbabilityLow},
					},
					CitationMetadata: &genai.CitationMetadata{
						Citations: []*genai.Citation{
							{URI: "https://example.com"},
						},
					},
				},
			},
			check: func(t *testing.T, result *llms.ContentResponse) {
				metadata := result.Choices[0].GenerationInfo
				if metadata[SAFETY] == nil {
					t.Error("expected safety ratings in metadata")
				}
				if metadata[CITATIONS] == nil {
					t.Error("expected citations in metadata")
				}
			},
		},
		// Note: We can't test unknown part types easily because genai.Part
		// has an unexported method, so we can't create a mock implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertCandidates(tt.candidates, tt.usage)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertCandidates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

// Note: We cannot create a custom type that implements genai.Part
// because it has an unexported method toPart()

func TestConvertTools(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	tests := []struct {
		name    string
		tools   []llms.Tool
		wantErr bool
		check   func(t *testing.T, result []*genai.Tool)
	}{
		{
			name: "single function tool",
			tools: []llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:        "get_weather",
						Description: "Get weather information",
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"location": map[string]any{
									"type":        "string",
									"description": "The city name",
								},
								"unit": map[string]any{
									"type":        "string",
									"description": "Temperature unit",
								},
							},
							"required": []string{"location"},
						},
					},
				},
			},
			check: func(t *testing.T, result []*genai.Tool) {
				if len(result) != 1 {
					t.Fatalf("expected 1 tool, got %d", len(result))
				}
				if len(result[0].FunctionDeclarations) != 1 {
					t.Fatalf("expected 1 function declaration, got %d", len(result[0].FunctionDeclarations))
				}
				fd := result[0].FunctionDeclarations[0]
				if fd.Name != "get_weather" {
					t.Errorf("expected function name 'get_weather', got %q", fd.Name)
				}
				if fd.Description != "Get weather information" {
					t.Errorf("expected description, got %q", fd.Description)
				}
				if fd.Parameters.Type != genai.TypeObject {
					t.Errorf("expected object type, got %v", fd.Parameters.Type)
				}
				if len(fd.Parameters.Properties) != 2 {
					t.Errorf("expected 2 properties, got %d", len(fd.Parameters.Properties))
				}
				if fd.Parameters.Properties["location"].Type != genai.TypeString {
					t.Errorf("expected location to be string type")
				}
				if len(fd.Parameters.Required) != 1 || fd.Parameters.Required[0] != "location" {
					t.Errorf("expected required=['location'], got %v", fd.Parameters.Required)
				}
			},
		},
		{
			name: "unsupported tool type",
			tools: []llms.Tool{
				{
					Type: "unsupported",
					Function: &llms.FunctionDefinition{
						Name: "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid parameters type",
			tools: []llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name:       "test",
						Parameters: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid properties type",
			tools: []llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name: "test",
						Parameters: map[string]any{
							"type":       "object",
							"properties": "invalid",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "required as interface slice",
			tools: []llms.Tool{
				{
					Type: "function",
					Function: &llms.FunctionDefinition{
						Name: "test",
						Parameters: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"field": map[string]any{"type": "string"},
							},
							"required": []interface{}{"field"},
						},
					},
				},
			},
			check: func(t *testing.T, result []*genai.Tool) {
				fd := result[0].FunctionDeclarations[0]
				if len(fd.Parameters.Required) != 1 || fd.Parameters.Required[0] != "field" {
					t.Errorf("expected required=['field'], got %v", fd.Parameters.Required)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertTools(tt.tools)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertTools() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestShowContent(t *testing.T) {
	// This is mainly for coverage - just ensure it doesn't panic
	var buf strings.Builder
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.Text("Hello"),
				genai.Blob{MIMEType: "image/png", Data: []byte{1, 2, 3}},
				genai.FunctionCall{Name: "test", Args: map[string]any{"x": 1}},
				genai.FunctionResponse{Name: "test", Response: map[string]any{"result": "ok"}},
			},
		},
	}

	// Should not panic
	showContent(&buf, contents)

	output := buf.String()
	if !strings.Contains(output, "Role=user") {
		t.Error("expected output to contain role")
	}
	if !strings.Contains(output, "Text \"Hello\"") {
		t.Error("expected output to contain text")
	}
	if !strings.Contains(output, "Blob MIME=\"image/png\"") {
		t.Error("expected output to contain blob info")
	}
	if !strings.Contains(output, "FunctionCall Name=test") {
		t.Error("expected output to contain function call")
	}
	if !strings.Contains(output, "FunctionResponse Name=test") {
		t.Error("expected output to contain function response")
	}
}

func TestErrorValues(t *testing.T) {
	// Test that error variables are properly defined
	if ErrNoContentInResponse == nil {
		t.Error("ErrNoContentInResponse should not be nil")
	}
	if ErrUnknownPartInResponse == nil {
		t.Error("ErrUnknownPartInResponse should not be nil")
	}
	if ErrInvalidMimeType == nil {
		t.Error("ErrInvalidMimeType should not be nil")
	}

	// Test error messages
	if !strings.Contains(ErrNoContentInResponse.Error(), "no content") {
		t.Error("ErrNoContentInResponse should mention 'no content'")
	}
	if !strings.Contains(ErrUnknownPartInResponse.Error(), "unknown part") {
		t.Error("ErrUnknownPartInResponse should mention 'unknown part'")
	}
	if !strings.Contains(ErrInvalidMimeType.Error(), "invalid mime") {
		t.Error("ErrInvalidMimeType should mention 'invalid mime'")
	}
}

func TestConstants(t *testing.T) {
	// Test that constants have expected values
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"CITATIONS", CITATIONS, "citations"},
		{"SAFETY", SAFETY, "safety"},
		{"RoleSystem", RoleSystem, "system"},
		{"RoleModel", RoleModel, "model"},
		{"RoleUser", RoleUser, "user"},
		{"RoleTool", RoleTool, "tool"},
		{"ResponseMIMETypeJson", ResponseMIMETypeJson, "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("expected %s to be %q, got %q", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestConvertAndStreamFromIterator(t *testing.T) {
	// Skip actual implementation tests since we can't create a real iterator
	// These tests would need integration with the actual genai package
	t.Skip("Skipping iterator tests - requires real genai.GenerateContentResponseIterator")
}

// Test that Vertex implements llms.Model interface
func TestVertexImplementsModel(t *testing.T) {
	var _ llms.Model = &Vertex{}
}
