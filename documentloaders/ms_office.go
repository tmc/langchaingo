package documentloaders

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/richardlehane/mscfb"
	"github.com/tealeg/xlsx"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	txtASCIIRangeMin = 32
	txtASCIIRangeMax = 126
	CR               = 13
	LF               = 10
)

type Office struct {
	reader   io.ReaderAt
	size     int64
	fileType string
}

var _ Loader = Office{}

func NewOffice(reader io.ReaderAt, filename string, size int64) Office {
	return Office{
		reader:   reader,
		size:     size,
		fileType: strings.ToLower(filepath.Ext(filename)),
	}
}

func (loader Office) Load(ctx context.Context) ([]schema.Document, error) {

	switch loader.fileType {
	case ".doc", ".docx":
		return loader.loadWord()
	case ".xls", ".xlsx":
		return loader.loadExcel()
	case ".ppt", ".pptx":
		return loader.loadPowerPoint()
	default:
		return nil, fmt.Errorf("unsupported file type: %s", loader.fileType)
	}
}

func (loader Office) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	return textsplitter.SplitDocuments(splitter, docs)
}

func (loader Office) loadWord() ([]schema.Document, error) {
	if loader.fileType == ".docx" {
		return loader.loadDocx()
	}

	return loader.loadDoc()
}

func (loader Office) loadDoc() ([]schema.Document, error) {
	doc, err := mscfb.New(io.NewSectionReader(loader.reader, 0, loader.size))
	if err != nil {
		return nil, fmt.Errorf("failed to read DOC file: %w", err)
	}

	var text strings.Builder
	for {
		entry, err := doc.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading DOC stream: %w", err)
		}

		if entry.Name == "WordDocument" {
			buf := make([]byte, entry.Size)
			i, err := doc.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("error reading WordDocument stream: %w", err)
			}

			if i > 0 {
				for j := 0; j < i; j++ {
					if buf[j] >= txtASCIIRangeMin && buf[j] <= txtASCIIRangeMax {
						text.WriteByte(buf[j])
					} else if buf[j] == CR || buf[j] == LF {
						text.WriteByte('\n')
					}
				}
			}
		}
	}

	return []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}, nil
}

func (loader Office) loadDocx() ([]schema.Document, error) {
	zipReader, err := zip.NewReader(loader.reader, loader.size)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX file as ZIP: %w", err)
	}

	var textContent string
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening document.xml: %w", err)
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("error reading content: %w", err)
			}

			textContent = loader.parseDocx(content)
			break
		}
	}

	return []schema.Document{
		{
			PageContent: textContent,
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}, nil
}

func (loader Office) parseDocx(xmlContent []byte) string {
	re := regexp.MustCompile(`<w:t(?:\s+[^>]*)?>(.*?)</w:t>`)
	matches := re.FindAllSubmatch(xmlContent, -1)

	var textBuilder strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			if textBuilder.Len() > 0 {
				textBuilder.WriteString(" ")
			}
			textBuilder.Write(match[1])
		}
	}

	text := textBuilder.String()
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

func (loader Office) loadExcel() ([]schema.Document, error) {
	buff := make([]byte, loader.size)
	_, err := loader.reader.ReadAt(buff, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %w", err)
	}

	xlFile, err := xlsx.OpenBinary(buff)
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

func (loader Office) loadPowerPoint() ([]schema.Document, error) {
	if loader.fileType == ".pptx" {
		return loader.loadPptx()
	}
	return loader.loadPpt()
}

func (loader Office) loadPpt() ([]schema.Document, error) {
	doc, err := mscfb.New(io.NewSectionReader(loader.reader, 0, loader.size))
	if err != nil {
		return nil, fmt.Errorf("failed to read PPT file: %w", err)
	}

	var text strings.Builder
	for {
		entry, err := doc.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading PPT stream: %w", err)
		}

		if entry.Name == "PowerPoint Document" {
			buf := make([]byte, entry.Size)

			i, err := doc.Read(buf)
			if err != nil {
				return nil, fmt.Errorf("error reading PPT stream: %w", err)
			}

			if i > 0 {
				for j := 0; j < i; j++ {
					if buf[j] >= txtASCIIRangeMin && buf[j] <= txtASCIIRangeMax {
						text.WriteByte(buf[j])
					} else if buf[j] == CR || buf[j] == LF {
						text.WriteByte('\n')
					}
				}
			}
		}

	}

	return []schema.Document{
		{
			PageContent: text.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}, nil
}

func (loader Office) loadPptx() ([]schema.Document, error) {
	zipReader, err := zip.NewReader(loader.reader, loader.size)
	if err != nil {
		return nil, fmt.Errorf("failed to read PPTX file as ZIP: %w", err)
	}

	var allText strings.Builder
	slideFilePattern := regexp.MustCompile(`ppt/slides/slide[0-9]+\.xml`)

	for _, file := range zipReader.File {
		if slideFilePattern.MatchString(file.Name) {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("error opening slide %s: %w", file.Name, err)
			}

			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, fmt.Errorf("error reading slide content: %w", err)
			}

			slideText := loader.parsePPTX(content)
			if slideText != "" {
				if allText.Len() > 0 {
					allText.WriteString("\n\n") // Add paragraph breaks between slides
				}
				allText.WriteString(slideText)
			}
		}
	}

	return []schema.Document{
		{
			PageContent: allText.String(),
			Metadata: map[string]interface{}{
				"fileType": loader.fileType,
			},
		},
	}, nil
}

func (loader Office) parsePPTX(xmlContent []byte) string {
	re := regexp.MustCompile(`<a:t(?:\s+[^>]*)?>(.*?)</a:t>`)
	matches := re.FindAllSubmatch(xmlContent, -1)

	var textBuilder strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			if textBuilder.Len() > 0 {
				textBuilder.WriteString(" ")
			}
			textBuilder.Write(match[1])
		}
	}

	text := textBuilder.String()
	spaceRe := regexp.MustCompile(`\s+`)
	text = spaceRe.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}
