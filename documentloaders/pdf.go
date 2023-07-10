package documentloaders

import (
	"code.sajari.com/docconv"
	"context"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"io"
	"strings"
)

type PDF struct {
	r io.ReadSeeker
}

func NewPDF(r io.ReadSeeker) PDF {
	return PDF{r}
}

// Load reads from the io.ReadSeeker and returns a single document with the data.
func (p PDF) Load(ctx context.Context) ([]schema.Document, error) {
	_, err := p.r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Convert the uploaded file to a human-readable text
	bodyResult, meta, err := docconv.ConvertPDF(p.r)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]any, len(meta))
	for k, v := range meta {
		metadata[k] = v
	}

	// Remove extra space and newlines
	contents := strings.TrimSpace(bodyResult)
	var docs = []schema.Document{
		{
			PageContent: contents,
			Metadata:    metadata,
		},
	}

	return docs, nil
}

// LoadAndSplit reads text data from the io.ReadSeeker and splits it into multiple
// documents using a text splitter.
func (p PDF) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := p.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
