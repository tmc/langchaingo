package llms

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
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
			name: "single text part",
			input: `role: user
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
			name: "multiple parts",
			input: `role: user
parts:
- type: text
  text: Hello!, world!
- type: image_url
  image_url:
    url: http://example.com/image.png
- type: image_url
  image_url:
    url: http://example.com/image.png
    detail: high
- type: binary
  binary:
    mime_type: application/octet-stream
    data: SGVsbG8sIHdvcmxkIQ==
- tool_response:
    tool_call_id: "123"
    name: hammer
    content: hit
  type: tool_response
`,
			want: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello!, world!"},
					ImageURLContent{URL: "http://example.com/image.png"},
					ImageURLContent{URL: "http://example.com/image.png", Detail: "high"},
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
				t.Log("in:", tt.input)
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
			name: "single text part",
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
			name: "multiple parts",
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
- image_url:
    url: http://example.com/image.png
  type: image_url
- binary:
    data: SGVsbG8sIHdvcmxkIQ==
    mime_type: application/octet-stream
  type: binary
- tool_response:
    content: hit
    name: hammer
    tool_call_id: "123"
  type: tool_response
role: user
`,
			wantErr: false,
		},
		{
			name: "unknown content type",
			input: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					unknownContent{},
				},
			},
			want: "parts:\n- {}\nrole: user\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := yaml.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, string(got)); diff != "" {
				t.Errorf("MarshalYAML() mismatch (-want +got):\n%s", diff)
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
			name:  "single text part",
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
			name:  "multiple parts",
			input: `{"role":"user","parts":[{"text":"Hello, world!","type":"text"},{"type":"image_url","image_url":{"url":"http://example.com/image.png"}},{"type":"binary","binary":{"data":"SGVsbG8sIHdvcmxkIQ==","mime_type":"application/octet-stream"}}]}`,
			want: MessageContent{
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
						FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "New York" }`},
					},
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
			t.Parallel()
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
			name: "single text part",
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
			name: "multiple parts",
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
			want:    `{"role":"user","parts":[{"text":"Hello, world!","type":"text"},{"type":"image_url","image_url":{"url":"http://example.com/image.png"}},{"type":"binary","binary":{"data":"SGVsbG8sIHdvcmxkIQ==","mime_type":"application/octet-stream"}}]}`,
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
			got, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
			}
			gotStr := string(got)
			if diff := cmp.Diff(tt.want, gotStr); diff != "" {
				t.Errorf("MarshalJSON() mismatch (-want +got):\n%s", diff)
				t.Log("got:", gotStr)
			}
		})
	}
}

// Test roundtripping for both JSON and YAML representations.
func TestRoundtripping(t *testing.T) { // nolint:funlen // We make an exception given the number of test cases.
	t.Parallel()
	tests := []struct {
		name         string
		in           MessageContent
		assertedJSON string
		assertedYAML string
	}{
		{
			name: "single text part",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello, world!"},
				},
			},
			assertedJSON: `{"role":"user","text":"Hello, world!"}`,
			assertedYAML: "role: user\ntext: Hello, world!\n",
		},
		{
			name: "multiple parts",
			in: MessageContent{
				Role: "user",
				Parts: []ContentPart{
					TextContent{Text: "Hello!, world!"},
					ImageURLContent{URL: "http://example.com/image.png", Detail: "low"},
					BinaryContent{
						MIMEType: "application/octet-stream",
						Data:     []byte("Hello, world!"),
					},
				},
			},
			assertedYAML: `parts:
- text: Hello!, world!
  type: text
- image_url:
    detail: low
    url: http://example.com/image.png
  type: image_url
- binary:
    data: SGVsbG8sIHdvcmxkIQ==
    mime_type: application/octet-stream
  type: binary
role: user
`,
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
		{
			name: "multiple tool uses",
			in: MessageContent{
				Role: "assistant",
				Parts: []ContentPart{
					ToolCall{Type: "function", ID: "tc01", FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "New York" }`}},
					ToolCall{Type: "function", ID: "tc02", FunctionCall: &FunctionCall{Name: "get_current_weather", Arguments: `{ "location": "Berlin" }`}},
				},
			},
			assertedJSON: `{"role":"assistant","parts":[{"type":"tool_call","tool_call":{"function":{"name":"get_current_weather","arguments":"{ \"location\": \"New York\" }"},"id":"tc01","type":"function"}},{"type":"tool_call","tool_call":{"function":{"name":"get_current_weather","arguments":"{ \"location\": \"Berlin\" }"},"id":"tc02","type":"function"}}]}`,
			assertedYAML: `parts:
- tool_call:
    function:
      arguments: '{ "location": "New York" }'
      name: get_current_weather
    id: tc01
    type: function
  type: tool_call
- tool_call:
    function:
      arguments: '{ "location": "Berlin" }'
      name: get_current_weather
    id: tc02
    type: function
  type: tool_call
role: assistant
`,
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

	// Round-trip both JSON and YAML:
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// JSON
			jsonBytes, err := json.Marshal(tt.in)
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.assertedJSON, string(jsonBytes)); diff != "" && tt.assertedJSON != "" {
				t.Errorf("JSON mismatch (-want +got):\n%s", diff)
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
			if diff := cmp.Diff(tt.assertedYAML, string(yamlBytes)); diff != "" && tt.assertedYAML != "" {
				t.Errorf("YAML mismatch (-want +got):\n%s", diff)
				t.Log("got:", string(yamlBytes))
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
