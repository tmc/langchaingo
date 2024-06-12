package llms

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

type unknownContent struct{}

func (unknownContent) isPart() {}

func TestUnmarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    MessageContent
		wantErr bool
	}{
		{
			name: "Single text part",
			input: `
role: user
text: Hello, world!
`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple parts",
			input: `role: user
parts:
- type: text
  text: Hello!, world!
- type: image_url
  url: http://example.com/image.png
- type: binary
  mime_type: application/octet-stream
  data: SGVsbG8sIHdvcmxkIQ==
- tool_response:
    toolcallid: "123"
    name: hammer
    content: hit
  type: tool_response
`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello!, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
				},
			},
			wantErr: false,
		},
		{
			name: "Unknown content type",
			input: `
role: user
parts:
  - type: unknown
    data: some data
`,
			want: MessageContent{
				Role: "user",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var mc MessageContent
			err := yaml.Unmarshal([]byte(tt.input), &mc)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, mc); diff != "" {
				t.Errorf("UnmarshalYAML() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalYAML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   MessageContent
		want    string
		wantErr bool
	}{
		{
			name: "Single text part",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
				},
			},
			want: `role: user
text: Hello, world!
`,
			wantErr: false,
		},
		{
			name: "Multiple parts",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
					ToolCallResponse{
						ToolCallID: "123",
						Name:       "hammer",
						Content:    "hit",
					},
				},
			},
			want: `parts:
    - text: Hello, world!
      type: text
    - type: image_url
      url: http://example.com/image.png
    - data: SGVsbG8sIHdvcmxkIQ==
      mime_type: application/octet-stream
      type: binary
    - tool_response:
        toolcallid: "123"
        name: hammer
        content: hit
      type: tool_response
role: user
`,
			wantErr: false,
		},
		{
			name: "Unknown content type",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					unknownContent{},
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.input.MarshalYAML()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				gotBytes, err := yaml.Marshal(got)
				if err != nil {
					t.Errorf("yaml.Marshal() error = %v", err)
					return
				}
				gotStr := string(gotBytes)
				if diff := cmp.Diff(tt.want, gotStr); diff != "" {
					t.Errorf("MarshalYAML() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestUnmarshalJSONMessageContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    MessageContent
		wantErr bool
	}{
		{
			name:  "Single text part",
			input: `{"role":"user","text":"Hello, world!"}`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{

					TextContent{Text: "Hello, world!"},
				},
			},

			wantErr: false,
		},
		{
			name:  "Multiple parts",
			input: `{"role":"user","parts":[{"type":"text","text":"Hello!, world!"},{"type":"image_url","image_url":{"url":"http://example.com/image.png"}},{"type":"binary","binary":{"mime_type":"application/octet-stream","data":"SGVsbG8sIHdvcmxkIQ=="}}]}`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello!, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "Unknown content type",
			input: `{"role":"user","parts":[{"type":"unknown","data":"some data"}]}`,
			want: MessageContent{
				Role: "user",
			},
			wantErr: true,
		},
		{
			name:  "tool use",
			input: `{"role":"assistant","parts":[{"type":"text","text":"Hello there!"},{"type":"tool_call","tool_call":{"id":"t42","type":"function","function":{"name":"get_current_weather","arguments":"{ \"location\": \"New York\" }"}}}]}`,
			want: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{
					TextContent{Text: "Hello there!"},
					ToolCall{
						ID:           "t42",
						Type:         "function",
						FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "New York" }`}},
				},
			},
			wantErr: false,
		},
		{
			name:  "tool response",
			input: `{"role":"user","parts":[{"type":"tool_response","tool_response":{"tool_call_id":"123","name":"hammer","content":"hit"}}]}`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()
			var mc MessageContent
			err := mc.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, mc); diff != "" {
				t.Errorf("UnmarshalJSON() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMarshalJSONMessageContent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   MessageContent
		want    string
		wantErr bool
	}{
		{
			name: "Single text part",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
				},
			},
			want:    `{"role":"user","text":"Hello, world!"}`,
			wantErr: false,
		},
		{
			name: "Multiple parts",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
				},
			},
			want:    `{"role":"user","parts":[{"text":"Hello, world!","type":"text"},{"image_url":{"url":"http://example.com/image.png"},"type":"image_url"},{"binary":{"data":"SGVsbG8sIHdvcmxkIQ==","mime_type":"application/octet-stream"},"type":"binary"}]}`,
			wantErr: false,
		},
		{
			name: "Unknown content type",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					unknownContent{},
				},
			},
			want:    `{"role":"user","parts":[{}]}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.input.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				gotStr := string(got)
				if diff := cmp.Diff(tt.want, gotStr); diff != "" {
					t.Errorf("MarshalJSON() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

// Test roundtripping for both JSON and YAML

func TestRoundtripping(t *testing.T) {
	//t.Parallel()
	tests := []struct {
		name string
		in   any
	}{
		{
			name: "Single text part",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
				},
			},
		},
		{
			name: "Multiple parts",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello!, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
				},
			},
		},
		{
			name: "tool use",
			in: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{
					ToolCall{Type: "function", ID: "t01", FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "New York" }`}},
				},
			},
		},
		// multiple tool uses:
		{
			name: "multiple tool uses",
			in: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{
					ToolCall{Type: "function", FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "New York" }`}},
					ToolCall{Type: "function", FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "Berlin" }`}},
				},
			},
		},
		{
			name: "tool use with arguments",
			in: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{
					ToolCall{Type: "hammer", FunctionCall: &FunctionCall{Name: "hit", Arguments: `{ "force": 10 }`}},
				},
			},
		},
		{
			name: "tool use with multiple arguments",
			in: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{

					ToolCall{Type: "hammer", FunctionCall: &FunctionCall{Name: "hit", Arguments: `{ "force": 10, "direction": "down" }`}},
				},
			},
		},
		{
			name: "tool response",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
				},
			},
		},
		{
			name: "multi-tool response",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
					ToolCallResponse{ToolCallID: "456", Name: "screwdriver", Content: "turn"},
				},
			},
		},
		{
			name: "tool response with arguments",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
				},
			},
		},
		{

			name: "multi-tool response with arguments",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{

					ToolCallResponse{ToolCallID: "123", Name: "hammer", Content: "hit"},
					ToolCallResponse{ToolCallID: "456", Name: "screwdriver", Content: "turn"},
				},
			},
		},
	}

	// round-trip both JSON and YAML:
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//t.Parallel()
			// JSON
			jsonBytes, err := json.Marshal(tt.in)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			var mc MessageContent
			err = mc.UnmarshalJSON(jsonBytes)
			if err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.in, mc); diff != "" {
				t.Errorf("Roundtrip JSON mismatch (-want +got):\n%s", diff)
				t.Logf("JSON: %s", jsonBytes)
			}

			// YAML
			yamlBytes, err := yaml.Marshal(tt.in)
			if err != nil {
				t.Errorf("MarshalYAML() error = %v", err)
				return
			}
			mc = MessageContent{}
			err = yaml.Unmarshal(yamlBytes, &mc)
			if err != nil {
				t.Errorf("UnmarshalYAML() error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.in, mc); diff != "" {
				t.Errorf("Roundtrip YAML mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
