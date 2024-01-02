package llms

import "encoding/json"

// ContentPart is an interface all parts of content have to implement.
type ContentPart interface {
	isPart()
}

// TextContent is content with some text.
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

func (TextContent) isPart() {}

// ImageURLContent is content with an URL pointing to an image.
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

func (ImageURLContent) isPart() {}

// BinaryContent is content holding some binary data with a MIME type.
type BinaryContent struct {
	MIMEType string
	Data     []byte
}

func (BinaryContent) isPart() {}

type ContentResponse struct {
	Choices []*ContentChoice
}

type ContentChoice struct {
	Content        string
	StopReason     string
	GenerationInfo map[string]any
}
