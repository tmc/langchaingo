package llms

import "encoding/json"

type ContentPart interface {
	isPart()
}

type TextContent struct {
	Text string
}

func (tc TextContent) MarshalJSON() ([]byte, error) {
	m := map[string]string{
		"type": "text",
		"text": tc.Text,
	}
	return json.Marshal(m)
}

func (_ TextContent) isPart() {}

type ImageURLContent struct {
	URL string
}

func (iuc ImageURLContent) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"type": "image_url",
		"image_url": map[string]string{
			"url": iuc.URL,
		},
	}
	return json.Marshal(m)
}

func (_ ImageURLContent) isPart() {}

type BinaryContent struct {
	MIMEType string
	Data     []byte
}

func (_ BinaryContent) isPart() {}

type ContentResponse struct {
	Choices []*ContentChoice
}

type ContentChoice struct {
	Content        string
	StopReason     string
	GenerationInfo map[string]any
}
