package documentloaders

import (
	"context"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// HTML loads parses and sanitizes html content from an io.Reader.
type HTML struct {
	r io.Reader
}

var _ Loader = HTML{}

// NewHTML creates a new html loader with an io.Reader.
func NewHTML(r io.Reader) HTML {
	return HTML{r}
}

// Load reads from the io.Reader and returns a single document with the data.
func (h HTML) Load(_ context.Context) ([]schema.Document, error) {
	doc, err := goquery.NewDocumentFromReader(h.r)
	if err != nil {
		return nil, err
	}

	var sel *goquery.Selection
	if doc.Has("body") != nil {
		sel = doc.Find("body").Contents()
	} else {
		sel = doc.Contents()
	}

	sanitized := bluemonday.UGCPolicy().Sanitize(sel.Text())
	pagecontent := strings.TrimSpace(sanitized)

	return []schema.Document{
		{
			PageContent: pagecontent,
			Metadata:    map[string]any{},
		},
	}, nil
}

// LoadAndSplit reads text data from the io.Reader and splits it into multiple
// documents using a text splitter.
func (h HTML) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := h.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
