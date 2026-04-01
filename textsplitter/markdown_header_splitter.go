package textsplitter

import (
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

var headerRegex = regexp.MustCompile(`^(#+)\s+(.*)`)

// MarkdownHeaderTextSplitter splits markdown by headers, storing hierarchy in metadata.
type MarkdownHeaderTextSplitter struct {
	HeaderPathSeparator string
	IncludeMetadata     bool
}

type headerInfo struct {
	level int
	text  string
}

func NewMarkdownHeaderTextSplitter(opts ...Option) *MarkdownHeaderTextSplitter {
	options := DefaultOptions()
	for _, o := range opts {
		o(&options)
	}

	return &MarkdownHeaderTextSplitter{
		HeaderPathSeparator: "/",
		IncludeMetadata:     true,
	}
}

func (sp *MarkdownHeaderTextSplitter) SplitText(text string) ([]string, error) {
	docs, err := sp.SplitTextToDocuments(text)
	if err != nil {
		return nil, err
	}

	chunks := make([]string, len(docs))
	for i, doc := range docs {
		chunks[i] = doc.PageContent
	}
	return chunks, nil
}

func (sp *MarkdownHeaderTextSplitter) SplitTextToDocuments(text string) ([]schema.Document, error) {
	var documents []schema.Document
	lines := strings.Split(text, "\n")
	var currentSection strings.Builder
	headerStack := []headerInfo{}
	inCodeBlock := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "```") || strings.HasPrefix(trimmedLine, "~~~") {
			inCodeBlock = !inCodeBlock
			currentSection.WriteString(line + "\n")
			continue
		}

		if !inCodeBlock {
			matches := headerRegex.FindStringSubmatch(line)
			if matches != nil {
				if strings.TrimSpace(currentSection.String()) != "" {
					doc := sp.buildDocumentFromSection(
						strings.TrimSpace(currentSection.String()),
						headerStack,
					)
					documents = append(documents, doc)
					currentSection.Reset()
				}

				headerLevel := len(matches[1])
				headerText := matches[2]

				for len(headerStack) > 0 && headerStack[len(headerStack)-1].level >= headerLevel {
					headerStack = headerStack[:len(headerStack)-1]
				}

				headerStack = append(headerStack, headerInfo{
					level: headerLevel,
					text:  headerText,
				})

				currentSection.WriteString(strings.Repeat("#", headerLevel) + " " + headerText + "\n")
				continue
			}
		}

		currentSection.WriteString(line + "\n")
	}

	if strings.TrimSpace(currentSection.String()) != "" {
		doc := sp.buildDocumentFromSection(
			strings.TrimSpace(currentSection.String()),
			headerStack,
		)
		documents = append(documents, doc)
	}

	return documents, nil
}

func (sp *MarkdownHeaderTextSplitter) SplitDocuments(docs []schema.Document) ([]schema.Document, error) {
	var result []schema.Document
	for _, doc := range docs {
		splitDocs, err := sp.SplitTextToDocuments(doc.PageContent)
		if err != nil {
			return nil, err
		}

		for i := range splitDocs {
			splitDocs[i].Metadata = mergeMetadata(doc.Metadata, splitDocs[i].Metadata)
			splitDocs[i].Score = doc.Score
		}
		result = append(result, splitDocs...)
	}
	return result, nil
}

func (sp *MarkdownHeaderTextSplitter) buildDocumentFromSection(
	textSplit string,
	headerStack []headerInfo,
) schema.Document {
	doc := schema.Document{
		PageContent: textSplit,
		Metadata:    make(map[string]any),
	}

	if sp.IncludeMetadata {
		doc.Metadata["header_path"] = sp.getHeaderPath(headerStack)
	}
	return doc
}

func (sp *MarkdownHeaderTextSplitter) getHeaderPath(headerStack []headerInfo) string {
	if len(headerStack) == 0 {
		return sp.HeaderPathSeparator
	}

	parentHeaders := headerStack
	if len(headerStack) > 0 {
		parentHeaders = headerStack[:len(headerStack)-1]
	}

	if len(parentHeaders) == 0 {
		return sp.HeaderPathSeparator
	}

	var pathBuilder strings.Builder
	pathBuilder.WriteString(sp.HeaderPathSeparator)

	for i, header := range parentHeaders {
		safeText := strings.ReplaceAll(header.text, sp.HeaderPathSeparator, "-")
		pathBuilder.WriteString(safeText)
		if i < len(parentHeaders)-1 {
			pathBuilder.WriteString(sp.HeaderPathSeparator)
		}
	}

	pathBuilder.WriteString(sp.HeaderPathSeparator)
	return pathBuilder.String()
}

func mergeMetadata(original, new map[string]any) map[string]any {
	result := make(map[string]any)
	for k, v := range original {
		result[k] = v
	}
	for k, v := range new {
		result[k] = v
	}
	return result
}
