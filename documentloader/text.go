package documentloader

import (
	"bytes"
	"io"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// Text loads text data from an io.Reader.
type Text struct {
	r io.Reader
}

var _ Loader = Text{}

// NewText creates a new text loader with an io.Reader.
func NewText(r io.Reader) Text {
	return Text{
		r: r,
	}
}

// Load reads from the io.Reader and returns a single document with the data.
func (l Text) Load() ([]schema.Document, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, l.r)
	if err != nil {
		return nil, err
	}

	return []schema.Document{
		{
			PageContent: buf.String(),
			Metadata:    map[string]any{},
		},
	}, nil
}

// LoadAndSplit reads text data from the io.Reader and splits it into multiple
// documents using a text splitter.
func (l Text) LoadAndSplit(splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load()
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
