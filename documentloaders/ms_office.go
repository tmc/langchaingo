package documentloaders

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/richardlehane/mscfb"
	"github.com/tealeg/xlsx"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

var _ Loader = Office{}

// Office loads text data from an io.Reader.
type Office struct {
	reader   io.ReaderAt
	size     int64
	fileType string
}

// NewOffice creates a new text loader with an io.Reader, filename and file size.
func NewOffice(reader io.ReaderAt, filename string, size int64) Office {
	return Office{
		reader:   reader,
		size:     size,
		fileType: strings.ToLower(filepath.Ext(filename)),
	}
}

// Load reads from the io.Reader for the MS Office data and returns the raw document data.
func (loader Office) Load(ctx context.Context) ([]schema.Document, error) {
	switch loader.fileType {
	case ".doc":
		return loader.loadDoc()
	case ".docx":
		return loader.loadDocx()
	case ".xls", ".xlsx":
		return loader.loadExcel()
	case ".ppt":
		return loader.loadPPT()
	case ".pptx":
		return loader.loadPPTX()
	default:
		return nil, fmt.Errorf("unsupported file type: %s", loader.fileType)
	}
}

// LoadAndSplit reads from the io.Reader for the MS Office data and returns the raw document data
// and splits it into multiple documents using a text splitter.
func (loader Office) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}

func (loader Office) loadDoc() ([]schema.Document, error) {
	doc, err := mscfb.New(io.NewSectionReader(loader.reader, 0, loader.size))
	if err != nil {
		return nil, fmt.Errorf("failed to read DOC file: %w", err)
	}

	var text strings.Builder
	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		if entry.Name == "WordDocument" {
			buf := make([]byte, entry.Size)
			i, err := doc.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("error reading WordDocument stream: %w", err)
			}
			if i > 0 {
				// Process the binary content
				for j := 0; j < i; j++ {
					// Extract readable ASCII text
					if buf[j] >= 32 && buf[j] <= 126 {
						text.WriteByte(buf[j])
					} else if buf[j] == 13 || buf[j] == 10 {
						text.WriteByte('\n')
					}
				}
			}
		}
	}

	documents := []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}

	return documents, nil
}

func (loader Office) loadExcel() ([]schema.Document, error) {
	buf := bytes.NewBuffer(make([]byte, 0, loader.size))
	if _, err := io.Copy(buf, io.NewSectionReader(loader.reader, 0, loader.size)); err != nil {
		return nil, fmt.Errorf("failed to copy Excel content: %w", err)
	}

	xlFile, err := xlsx.OpenBinary(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %w", err)
	}

	var docs []schema.Document
	for i, sheet := range xlFile.Sheets {
		var text strings.Builder
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				text.WriteString(cell.String() + "\t")
			}
			text.WriteString("\n")
		}

		docs = append(docs, schema.Document{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType":   loader.fileType,
				"sheetName":  sheet.Name,
				"sheetIndex": i,
			},
		})
	}

	return docs, nil
}

func (loader Office) loadPPT() ([]schema.Document, error) {
	doc, err := mscfb.New(io.NewSectionReader(loader.reader, 0, loader.size))
	if err != nil {
		return nil, fmt.Errorf("failed to read PPT file: %w", err)
	}

	var text strings.Builder
	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		if entry.Name == "PowerPoint Document" {
			buf := make([]byte, entry.Size)
			i, err := doc.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("error reading PowerPoint stream: %w", err)
			}
			if i > 0 {
				// Process the binary content
				for j := 0; j < i; j++ {
					// Extract readable ASCII text
					if buf[j] >= 32 && buf[j] <= 126 {
						text.WriteByte(buf[j])
					} else if buf[j] == 13 || buf[j] == 10 {
						text.WriteByte('\n')
					}
				}
			}
		}
	}

	documents := []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}

	return documents, nil
}

func (loader Office) loadPPTX() ([]schema.Document, error) {
	buf := bytes.NewBuffer(make([]byte, 0, loader.size))
	if _, err := io.Copy(buf, io.NewSectionReader(loader.reader, 0, loader.size)); err != nil {
		return nil, fmt.Errorf("failed to copy content: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), loader.size)
	if err != nil {
		return nil, fmt.Errorf("failed to read PPTX file as ZIP: %w", err)
	}

	var text strings.Builder
	for _, file := range zipReader.File {
		// PPTX stores slide content in ppt/slides/slide*.xml files
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening slide XML: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("error reading content: %w", err)
			}

			content = bytes.ReplaceAll(content, []byte("<"), []byte(" <"))
			content = bytes.ReplaceAll(content, []byte(">"), []byte("> "))
			text.Write(content)
			text.WriteString("\n--- Next Slide ---\n")
		}
	}

	documents := []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}

	return documents, nil
}

func (loader Office) loadDocx() ([]schema.Document, error) {
	buf := bytes.NewBuffer(make([]byte, 0, loader.size))
	if _, err := io.Copy(buf, io.NewSectionReader(loader.reader, 0, loader.size)); err != nil {
		return nil, fmt.Errorf("failed to copy content: %w", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), loader.size)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX file as ZIP: %w", err)
	}

	var text strings.Builder
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening document.xml: %w", err)
			}
			defer rc.Close()

			content, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("error reading content: %w", err)
			}

			content = bytes.ReplaceAll(content, []byte("<"), []byte(" <"))
			content = bytes.ReplaceAll(content, []byte(">"), []byte("> "))
			text.Write(content)
		}
	}

	documents := []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}

	return documents, nil
}
