package llms

import (
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
