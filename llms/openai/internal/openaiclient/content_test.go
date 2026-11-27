package openaiclient

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"testing"
)

func TestContentPart_MarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		part     llms.ContentPart
		expected string
	}{
		{
			name:     "text content",
			part:     llms.TextPart("Hello, world!"),
			expected: `{"type":"text","text":"Hello, world!"}`,
		},
		{
			name:     "image URL content",
			part:     llms.ImageURLPart("https://example.com/image.png"),
			expected: `{"type":"image_url","image_url":{"url":"https://example.com/image.png"}}`,
		},
		{
			name:     "audio content: mp3",
			part:     llms.BinaryPart("audio/mpeg", []byte{0x01, 0x02, 0x03}),
			expected: `{"type":"input_audio","input_audio":{"format":"mp3","data":"AQID"}}`,
		},
		{
			name:     "audio content: wav",
			part:     llms.BinaryPart("audio/wav", []byte{0x01, 0x02, 0x03}),
			expected: `{"type":"input_audio","input_audio":{"format":"wav","data":"AQID"}}`,
		},
		{
			name:     "file content: pdf",
			part:     llms.BinaryPart("application/pdf", []byte{0x01, 0x02, 0x03}),
			expected: `{"type":"file","file":{"filename":"langchaingo","file_data":"data:application/pdf;base64,AQID"}}`,
		},
		{
			name:     "file content: pdf with filename",
			part:     llms.BinaryPart("application/pdf; filename=test.pdf", []byte{0x01, 0x02, 0x03}),
			expected: `{"type":"file","file":{"filename":"test.pdf","file_data":"data:application/pdf;base64,AQID"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(ContentPart{ContentPart: tt.part})
			assert.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}
