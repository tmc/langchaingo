package documentloaders

import (
	"bytes"
	"context"
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
func (l Text) Load(_ context.Context) ([]schema.Document, error) {
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
func (l Text) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
