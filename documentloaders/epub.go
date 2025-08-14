package documentloaders

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/0xDezzy/langchaingo/schema"
	"github.com/0xDezzy/langchaingo/textsplitter"
	"github.com/jaytaylor/html2text"
	"github.com/timsims/pamphlet"
)

// EPUB loads text data from EPUB files.
type EPUB struct {
	filePath string
	reader   io.Reader
	data     []byte
	mode     string
}

var _ Loader = EPUB{}

// EPUBOptions are options for the EPUB loader.
type EPUBOptions func(epub *EPUB)

// WithMode sets the parsing mode for the EPUB loader.
// "single" mode returns all content as a single document.
// "elements" mode returns each chapter as a separate document.
func WithMode(mode string) EPUBOptions {
	return func(epub *EPUB) {
		epub.mode = mode
	}
}

// NewEPUB creates a new EPUB loader with a file path.
func NewEPUB(filePath string, opts ...EPUBOptions) EPUB {
	epub := EPUB{
		filePath: filePath,
		mode:     "single",
	}
	for _, opt := range opts {
		opt(&epub)
	}
	return epub
}

// NewEPUBFromReader creates a new EPUB loader with an io.Reader.
func NewEPUBFromReader(reader io.Reader, opts ...EPUBOptions) EPUB {
	epub := EPUB{
		reader: reader,
		mode:   "single",
	}
	for _, opt := range opts {
		opt(&epub)
	}
	return epub
}

// NewEPUBFromBytes creates a new EPUB loader with a byte slice.
func NewEPUBFromBytes(data []byte, opts ...EPUBOptions) EPUB {
	epub := EPUB{
		data: data,
		mode: "single",
	}
	for _, opt := range opts {
		opt(&epub)
	}
	return epub
}

// Load reads from the EPUB and returns documents with extracted text content.
// In "single" mode, returns one document with all text content.
// In "elements" mode, returns one document per chapter.
func (e EPUB) Load(_ context.Context) ([]schema.Document, error) {
	parser, err := e.openEPUB()
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}

	book := parser.GetBook()
	chapters := book.Chapters

	if e.mode == "single" {
		return e.loadSingle(book, chapters)
	}

	return e.loadElements(book, chapters)
}

// openEPUB opens the EPUB file using the appropriate method.
func (e EPUB) openEPUB() (*pamphlet.Parser, error) {
	if e.filePath != "" {
		return pamphlet.Open(e.filePath)
	}

	if e.data != nil {
		return pamphlet.OpenBytes(e.data)
	}

	if e.reader != nil {
		data, err := io.ReadAll(e.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read from reader: %w", err)
		}
		return pamphlet.OpenBytes(data)
	}

	return nil, fmt.Errorf("no valid input provided")
}

// loadSingle loads all chapters into a single document.
func (e EPUB) loadSingle(book *pamphlet.Book, chapters []pamphlet.Chapter) ([]schema.Document, error) {
	var contentBuilder strings.Builder

	for i, chapter := range chapters {
		htmlContent, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get content for chapter %d: %w", i+1, err)
		}
		plainText, err := html2text.FromString(htmlContent, html2text.Options{
			PrettyTables: false,
			OmitLinks:    true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTML to text for chapter %d: %w", i+1, err)
		}

		if contentBuilder.Len() > 0 {
			contentBuilder.WriteString("\n\n")
		}
		contentBuilder.WriteString(strings.TrimSpace(plainText))
	}

	metadata := map[string]any{
		"title":    book.Title,
		"author":   book.Author,
		"language": book.Language,
		"chapters": len(chapters),
		"mode":     e.mode,
	}

	if e.filePath != "" {
		metadata["source"] = e.filePath
	}

	return []schema.Document{
		{
			PageContent: contentBuilder.String(),
			Metadata:    metadata,
		},
	}, nil
}

// loadElements loads each chapter as a separate document.
func (e EPUB) loadElements(book *pamphlet.Book, chapters []pamphlet.Chapter) ([]schema.Document, error) {
	var docs []schema.Document

	for i, chapter := range chapters {
		htmlContent, err := chapter.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to get content for chapter %d: %w", i+1, err)
		}
		plainText, err := html2text.FromString(htmlContent, html2text.Options{
			PrettyTables: false,
			OmitLinks:    true,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to convert HTML to text for chapter %d: %w", i+1, err)
		}

		metadata := map[string]any{
			"title":          book.Title,
			"author":         book.Author,
			"language":       book.Language,
			"chapter":        i + 1,
			"total_chapters": len(chapters),
			"chapter_title":  chapter.Title,
			"mode":           e.mode,
		}

		if e.filePath != "" {
			metadata["source"] = e.filePath
		}

		docs = append(docs, schema.Document{
			PageContent: strings.TrimSpace(plainText),
			Metadata:    metadata,
		})
	}

	return docs, nil
}

// LoadAndSplit loads EPUB data and splits it into multiple documents using a text splitter.
func (e EPUB) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := e.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}
