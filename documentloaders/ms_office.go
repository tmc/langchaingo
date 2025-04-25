package documentloaders

import (
	"archive/zip"
	"context"
	"errors"
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

func (loader Office) Load(_ context.Context) ([]schema.Document, error) {
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

// extractTextFromBuffer processes a byte buffer to extract readable ASCII text.
func extractTextFromBuffer(buf []byte, size int) string {
	var text strings.Builder
	for j := 0; j < size; j++ {
		if buf[j] >= txtASCIIRangeMin && buf[j] <= txtASCIIRangeMax {
			text.WriteByte(buf[j])
		} else if buf[j] == CR || buf[j] == LF {
			text.WriteByte('\n')
		}
	}
	return text.String()
}

func (loader Office) loadDoc() ([]schema.Document, error) {
	doc, err := mscfb.New(io.NewSectionReader(loader.reader, 0, loader.size))
	if err != nil {
		return nil, fmt.Errorf("failed to read DOC file: %w", err)
	}

	var text strings.Builder
	for {
		entry, err := doc.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading DOC stream: %w", err)
		}

		if entry.Name == "WordDocument" {
			content, err := loader.readEntryContent(doc, entry)
			if err != nil {
				return nil, err
			}
			text.WriteString(content)
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

// readEntryContent reads the content of an entry and returns it as a string.
func (loader Office) readEntryContent(doc *mscfb.Reader, entry *mscfb.File) (string, error) {
	buf := make([]byte, entry.Size)
	i, err := doc.Read(buf)
	if err != nil {
		return "", fmt.Errorf("error reading content stream: %w", err)
	}

	if i > 0 {
		return extractTextFromBuffer(buf, i), nil
	}
	return "", nil
}

func (loader Office) loadDocx() ([]schema.Document, error) {
	zipReader, err := zip.NewReader(loader.reader, loader.size)
	if err != nil {
		return nil, fmt.Errorf("failed to read DOCX file as ZIP: %w", err)
	}

	var textContent string
	for _, file := range zipReader.File {
		if file.Name == "word/document.xml" {
			content, err := loader.readZipFileContent(file)
			if err != nil {
				return nil, err
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

// readZipFileContent reads the content of a zip file entry.
func (loader Office) readZipFileContent(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", file.Name, err)
	}
	defer rc.Close()

	content, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("error reading content: %w", err)
	}

	return content, nil
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

	docs := make([]schema.Document, 0, len(xlFile.Sheets))
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
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading PPT stream: %w", err)
		}

		if entry.Name == "PowerPoint Document" {
			content, err := loader.readEntryContent(doc, entry)
			if err != nil {
				return nil, err
			}
			text.WriteString(content)
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
		if !slideFilePattern.MatchString(file.Name) {
			continue
		}

		content, err := loader.readZipFileContent(file)
		if err != nil {
			return nil, err
		}

		slideText := loader.parsePPTX(content)
		if slideText != "" {
			if allText.Len() > 0 {
				allText.WriteString("\n\n") // Add paragraph breaks between slides
			}
			allText.WriteString(slideText)
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
