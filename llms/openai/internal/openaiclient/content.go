package openaiclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/tmc/langchaingo/llms"
	"regexp"
	"strings"
)

type ContentPart struct {
	llms.ContentPart
}

func (cp ContentPart) MarshalJSON() ([]byte, error) {
	switch cp.ContentPart.(type) {
	case llms.BinaryContent:
		// we have to handle binary content separately
		return marshalBinaryContent(cp.ContentPart.(llms.BinaryContent))
	default:
		// for all other content types, we can use the default JSON marshaling
		return json.Marshal(cp.ContentPart)
	}
}

func marshalBinaryContent(content llms.BinaryContent) ([]byte, error) {
	if strings.HasPrefix(content.MIMEType, "audio/") {
		// in case of audio content, we marshal it differently than other binary content
		return marshalAudioContent(content)
	}

	// openai expects binary content to be marshaled as a file with a filename
	// see: https://platform.openai.com/docs/guides/pdf-files#base64-encoded-files
	return json.Marshal(map[string]any{
		"type": "file",
		"file": map[string]any{
			"filename": extractFilename(content.MIMEType),
			"file_data": Base64Data{
				MIMEType: content.MIMEType,
				Data:     content.Data,
			},
		},
	})
}

var filenameRegex = regexp.MustCompile(`filename="?([^";]+)"?`)

func extractFilename(mimeType string) string {
	filename := "langchaingo"
	if matches := filenameRegex.FindStringSubmatch(mimeType); len(matches) > 1 {
		filename = matches[1]
	}
	return filename
}

func marshalAudioContent(content llms.BinaryContent) ([]byte, error) {
	format := ""
	if strings.HasPrefix(content.MIMEType, "audio/wav") {
		format = "wav"
	} else if strings.HasPrefix(content.MIMEType, "audio/mpeg") {
		format = "mp3"
	}

	// see: https://platform.openai.com/docs/guides/audio?example=audio-in#add-audio-to-your-existing-application
	return json.Marshal(map[string]any{
		"type": "input_audio",
		"input_audio": map[string]any{
			"format": format,
			"data":   content.Data,
		},
	})
}

type Base64Data struct {
	MIMEType string
	Data     []byte
}

func (bd Base64Data) MarshalJSON() ([]byte, error) {
	var r []byte

	mimeType := strings.Split(bd.MIMEType, ";")[0]

	r = append(r, []byte(`"data:`)...)
	r = append(r, []byte(mimeType)...)
	r = append(r, []byte(";base64,")...)

	b := bytes.NewBuffer(r)

	_, err := base64.NewEncoder(base64.StdEncoding, b).Write(bd.Data)
	if err != nil {
		return nil, err
	}

	r = b.Bytes()
	r = append(r, '"')

	return r, nil
}
