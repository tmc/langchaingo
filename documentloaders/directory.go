package documentloaders

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vendasta/langchaingo/schema"
	"github.com/vendasta/langchaingo/textsplitter"
)

type Option func(*RecursiveDirectoryLoader)

func WithRoot(root string) Option {
	return func(l *RecursiveDirectoryLoader) { l.root = root }
}

func WithMaxDepth(d int) Option {
	return func(l *RecursiveDirectoryLoader) { l.maxDepth = d }
}

func WithAllowExts(exts ...string) Option {
	return func(l *RecursiveDirectoryLoader) {
		l.allowExt = make(map[string]struct{}, len(exts))
		for _, e := range exts {
			e = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(e), "."))
			l.allowExt["."+e] = struct{}{}
		}
	}
}

func WithCSVOpts(cols []string) Option {
	return func(c *RecursiveDirectoryLoader) {
		c.Columns = cols
	}
}

func WithPDFOpts(pwd string) Option {
	return func(c *RecursiveDirectoryLoader) { c.PDFPassword = pwd }
}

// RecursiveDirectoryLoader is a document loader that loads documents with allowed extensions from a directory.
type RecursiveDirectoryLoader struct {
	root     string
	maxDepth int
	allowExt map[string]struct{}

	Columns []string // CSV Columns

	PDFPassword string // PDF password
}

var _ Loader = (*RecursiveDirectoryLoader)(nil)

func NewRecursiveDirLoader(opts ...Option) *RecursiveDirectoryLoader {
	l := &RecursiveDirectoryLoader{
		root:     ".",
		maxDepth: 1,
		allowExt: map[string]struct{}{},
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func (l *RecursiveDirectoryLoader) newLoader(f *os.File) (Loader, error) {
	ext := filepath.Ext(f.Name())
	switch ext {
	case ".txt", ".md":
		return NewText(f), nil
	case ".csv":
		return NewCSV(f, l.Columns...), nil
	case ".pdf":
		finfo, err := f.Stat()
		if err != nil {
			return nil, err
		}
		if l.PDFPassword != "" {
			return NewPDF(f, finfo.Size(), WithPassword(l.PDFPassword)), nil
		}
		return NewPDF(f, finfo.Size()), nil
	case ".html", ".htm":
		return NewHTML(f), nil
	default:
		return nil, fmt.Errorf("unsupported file extension %q", ext)
	}
}

// Load retrieves data from a Notion directory and returns a list of schema.Document objects.
func (l *RecursiveDirectoryLoader) Load(ctx context.Context) ([]schema.Document, error) {
	var docs []schema.Document

	err := filepath.WalkDir(l.root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, _ := filepath.Rel(l.root, path)
			depth := strings.Count(rel, string(os.PathSeparator))
			if depth >= l.maxDepth {
				return fs.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if len(l.allowExt) > 0 {
			if _, ok := l.allowExt[ext]; !ok {
				return nil
			}
		}

		var fileDocs []schema.Document
		if err := func() error {
			f, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer f.Close()

			loader, err := l.newLoader(f)
			if err != nil {
				return err
			}

			fileDocs, err = loader.Load(ctx)
			if err != nil {
				return err
			}
			for i := range fileDocs {
				if fileDocs[i].Metadata == nil {
					fileDocs[i].Metadata = make(map[string]any)
				}
				fileDocs[i].Metadata["source"] = path
			}
			return nil
		}(); err != nil {
			log.Printf("skip %s: %v", path, err)
		} else {
			docs = append(docs, fileDocs...)
		}

		return nil
	})
	return docs, err
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l *RecursiveDirectoryLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
