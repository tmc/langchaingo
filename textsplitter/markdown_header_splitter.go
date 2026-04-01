package textsplitter

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

// headerRegex matches markdown headers: ^(#+)\s+(.*)
// Compiled at package level for performance.
var headerRegex = regexp.MustCompile(`^(#+)\s+(.*)`)

// MarkdownHeaderTextSplitter splits markdown documents by headers and preserves
// hierarchical context in metadata. Inspired by LlamaIndex's MarkdownNodeParser.
//
// Unlike MarkdownTextSplitter which prepends headers to content, this stores
// the header hierarchy in Document.Metadata["header_path"], enabling section-based
// filtering in vector databases for RAG applications.
//
// Example: For content under "# Chapter 1" -> "## Section 1.1", the document
// will have metadata["header_path"] = "/Chapter 1/" (parent headers only).
type MarkdownHeaderTextSplitter struct {
	// HeaderPathSeparator is the separator used in header paths (default: "/").
	HeaderPathSeparator string

	// IncludeMetadata determines whether to add header metadata (default: true).
	IncludeMetadata bool
}

// headerInfo stores information about a header in the document.
type headerInfo struct {
	level int    // Header level (1-6 for H1-H6)
	text  string // Header text without # symbols
}

// NewMarkdownHeaderTextSplitter creates a new MarkdownHeaderTextSplitter.
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

// SplitText splits markdown text into chunks, returning strings without metadata.
// For metadata support, use SplitTextToDocuments or SplitDocuments instead.
func (sp *MarkdownHeaderTextSplitter) SplitText(text string) ([]string, error) {
	docs, err := sp.SplitTextToDocuments(text)
	if err != nil {
		return nil, err
	}

	// Extract just the page content
	chunks := make([]string, len(docs))
	for i, doc := range docs {
		chunks[i] = doc.PageContent
	}

	return chunks, nil
}

// SplitTextToDocuments splits markdown text into Documents with header metadata.
// Creates one document per header section with metadata["header_path"] containing
// parent headers only (matching LlamaIndex's MarkdownNodeParser behavior).
func (sp *MarkdownHeaderTextSplitter) SplitTextToDocuments(text string) ([]schema.Document, error) {
	var documents []schema.Document

	lines := strings.Split(text, "\n")
	// Use strings.Builder for efficient string concatenation
	var currentSection strings.Builder
	headerStack := []headerInfo{}
	inCodeBlock := false

	for _, line := range lines {
		// Track code blocks to avoid parsing # symbols in code as headers
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			currentSection.WriteString(line + "\n")
			continue
		}

		// Only parse headers outside code blocks
		if !inCodeBlock {
			matches := headerRegex.FindStringSubmatch(line)
			if matches != nil {
				// Save previous section before starting new one
				if strings.TrimSpace(currentSection.String()) != "" {
					doc := sp.buildDocumentFromSection(
						strings.TrimSpace(currentSection.String()),
						headerStack,
					)
					documents = append(documents, doc)
					currentSection.Reset() // Clear builder for next section
				}

				// Extract header level and text
				headerLevel := len(matches[1])
				headerText := matches[2]

				// Pop headers of equal or higher level (handles non-sequential headers)
				for len(headerStack) > 0 && headerStack[len(headerStack)-1].level >= headerLevel {
					headerStack = headerStack[:len(headerStack)-1]
				}

				// Add new header to stack
				headerStack = append(headerStack, headerInfo{
					level: headerLevel,
					text:  headerText,
				})

				// Start new section with header
				currentSection.WriteString(strings.Repeat("#", headerLevel) + " " + headerText + "\n")
				continue
			}
		}

		// Add line to current section
		currentSection.WriteString(line + "\n")
	}

	// Add final section
	if strings.TrimSpace(currentSection.String()) != "" {
		doc := sp.buildDocumentFromSection(
			strings.TrimSpace(currentSection.String()),
			headerStack,
		)
		documents = append(documents, doc)
	}

	return documents, nil
}

// SplitDocuments splits Documents while preserving original metadata.
// Merges original metadata (e.g., source, author) with new header metadata.
func (sp *MarkdownHeaderTextSplitter) SplitDocuments(docs []schema.Document) ([]schema.Document, error) {
	var result []schema.Document

	for _, doc := range docs {
		splitDocs, err := sp.SplitTextToDocuments(doc.PageContent)
		if err != nil {
			return nil, err
		}

		// Merge original metadata with header metadata
		for i := range splitDocs {
			splitDocs[i].Metadata = mergeMetadata(doc.Metadata, splitDocs[i].Metadata)
			splitDocs[i].Score = doc.Score
		}

		result = append(result, splitDocs...)
	}

	return result, nil
}

// buildDocumentFromSection creates a Document from a text section with header metadata.
func (sp *MarkdownHeaderTextSplitter) buildDocumentFromSection(
	textSplit string,
	headerStack []headerInfo,
) schema.Document {
	doc := schema.Document{
		PageContent: textSplit,
		Metadata:    make(map[string]any),
	}

	if sp.IncludeMetadata {
		// Build header path from parent headers (excluding current header)
		headerPath := sp.getHeaderPath(headerStack)
		doc.Metadata["header_path"] = headerPath

		// Add individual header levels for filtering
		for i, header := range headerStack {
			doc.Metadata[formatHeaderKey(i+1)] = header.text
		}
	}

	return doc
}

// getHeaderPath builds the header path from the header stack.
// Returns format: "/header1/header2/" (parent headers only, excluding current).
// This matches LlamaIndex's MarkdownNodeParser behavior.
func (sp *MarkdownHeaderTextSplitter) getHeaderPath(headerStack []headerInfo) string {
	if len(headerStack) == 0 {
		return sp.HeaderPathSeparator
	}

	// Use all headers except the last one (current header)
	parentHeaders := headerStack
	if len(headerStack) > 0 {
		parentHeaders = headerStack[:len(headerStack)-1]
	}

	if len(parentHeaders) == 0 {
		return sp.HeaderPathSeparator
	}

	// Build path using strings.Builder
	var pathBuilder strings.Builder
	pathBuilder.WriteString(sp.HeaderPathSeparator)

	for i, header := range parentHeaders {
		pathBuilder.WriteString(header.text)
		if i < len(parentHeaders)-1 {
			pathBuilder.WriteString(sp.HeaderPathSeparator)
		}
	}

	pathBuilder.WriteString(sp.HeaderPathSeparator)
	return pathBuilder.String()
}

// formatHeaderKey formats a header level key for metadata (e.g., "Header_1", "Header_2").
func formatHeaderKey(level int) string {
	return "Header_" + strconv.Itoa(level)
}

// mergeMetadata merges two metadata maps, with new metadata taking precedence.
func mergeMetadata(original, new map[string]any) map[string]any {
	result := make(map[string]any)

	// Copy original metadata
	for k, v := range original {
		result[k] = v
	}

	// Overlay new metadata (takes precedence)
	for k, v := range new {
		result[k] = v
	}

	return result
}
