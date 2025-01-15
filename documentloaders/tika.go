package documentloaders

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-tika/tika"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// TikaOption sets and option on a Tika document loader.
type TikaOption func(tika *Tika)

// WithHTTPClient sets the http client to be used by the Tika document loader.
func WithHTTPClient(cli *http.Client) TikaOption {
	return func(tika *Tika) {
		tika.httpcli = cli
	}
}

// WithHTTPHeaders sets custom http headers to be used when requiring Tika to
// parse a file. The header "Acccept" can't be set as this client always sets
// it to "text/plain".
func WithHTTPHeaders(key, value string) TikaOption {
	return func(tika *Tika) {
		tika.headers.Set(key, value)
	}
}

// Tika uses an Apache Tika Server to parse files. Tika toolkit detects and
// extracts metadata and text from over a thousand different file types (such
// as PPT, XLS, and PDF).
type Tika struct {
	reader  io.Reader
	address string
	httpcli *http.Client
	headers http.Header
}

// LoadAndSplit reads the document data and splits it into multple documents
// using the provided text splitter.
func (t Tika) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	client := tika.NewClient(t.httpcli, t.address)

	t.headers.Set("Accept", "text/plain")
	text, err := client.ParseWithHeader(ctx, t.reader, t.headers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	slices, err := splitter.SplitText(text)
	if err != nil {
		return nil, fmt.Errorf("failed to split text: %w", err)
	}

	docs := make([]schema.Document, len(slices))
	for i, slice := range slices {
		docs[i] = schema.Document{PageContent: slice}
	}
	return docs, nil
}

// NewTika returns a new Tika document loader. Uses the address of the Tika
// server and the reader to read the file data.
func NewTika(addr string, r io.Reader, opts ...TikaOption) *Tika {
	tika := &Tika{
		address: addr,
		reader:  r,
		headers: make(http.Header),
	}
	for _, opt := range opts {
		opt(tika)
	}
	return tika
}
