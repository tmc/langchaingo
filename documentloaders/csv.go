package documentloaders

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"golang.org/x/exp/slices"
)

// CSV represents a CSV document loader.
type CSV struct {
	r                  io.Reader
	columns            []string
	rowPropsInMetadata bool
}

var _ Loader = CSV{}

type CSVOption func(*CSV)

// NewCSV creates a new csv loader with an io.Reader and optional column names for filtering.
func NewCSV(r io.Reader, opts ...CSVOption) CSV {
	csv := CSV{
		r:                  r,
		rowPropsInMetadata: false,
	}

	for _, opt := range opts {
		opt(&csv)
	}

	return csv
}

// Load reads from the io.Reader and returns a single document with the data.
func (c CSV) Load(_ context.Context) ([]schema.Document, error) {
	var header []string
	var docs []schema.Document
	var rown int

	rd := csv.NewReader(c.r)
	for {
		row, err := rd.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(header) == 0 {
			header = append(header, row...)
			continue
		}

		var content []string
		metadata := map[string]any{}
		for i, value := range row {
			if c.columns != nil &&
				len(c.columns) > 0 &&
				!slices.Contains(c.columns, header[i]) {
				continue
			}
			if c.rowPropsInMetadata {
				metadata[header[i]] = value
			}

			line := fmt.Sprintf("%s: %s", header[i], value)
			content = append(content, line)
		}

		rown++
		metadata["row"] = rown
		docs = append(docs, schema.Document{
			PageContent: strings.Join(content, "\n"),
			Metadata:    metadata,
		})
	}

	return docs, nil
}

// LoadAndSplit reads text data from the io.Reader and splits it into multiple
// documents using a text splitter.
func (c CSV) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := c.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}

// WithColumnFilter defines the columns filtered.
func WithColumnFilter(columns ...string) CSVOption {
	return func(c *CSV) {
		c.columns = columns
	}
}

// WithRowPropertiesInMetadata defines if row should be added to metadata.
func WithRowPropertiesInMetadata() CSVOption {
	return func(c *CSV) {
		c.rowPropsInMetadata = true
	}
}
