package textsplitter

import (
	"errors"
	"log"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

// ErrMismatchMetadatasAndTextAndCustomIDs is returned when the number of texts and metadatas
// given to CreateDocuments does not match. The function will not error if the
// length of the metadatas slice is zero.
var ErrMismatchMetadatasAndTextAndCustomIDs = errors.New("number of texts, metadatas and custom IDs do not match")

// SplitDocuments splits documents using a textsplitter.
func SplitDocuments(textSplitter TextSplitter, documents []schema.Document) ([]schema.Document, error) {
	texts := make([]string, 0)
	metadatas := make([]map[string]any, 0)
	customIDs := make([]string, 0)
	for _, document := range documents {
		texts = append(texts, document.PageContent)
		metadatas = append(metadatas, document.Metadata)
		customIDs = append(customIDs, document.CustomID)
	}

	return CreateDocuments(textSplitter, texts, metadatas, customIDs)
}

// CreateDocuments creates documents from texts and metadatas with a text splitter. If
// the length of the metadatas is zero, the result documents will contain no metadata.
// Otherwise, the numbers of texts and metadatas must match.
func CreateDocuments(
	textSplitter TextSplitter, texts []string, metadatas []map[string]any, customIDs []string,
) ([]schema.Document, error) {
	if len(metadatas) == 0 {
		metadatas = make([]map[string]any, len(texts))
	}

	if len(texts) != len(metadatas) {
		return nil, ErrMismatchMetadatasAndTextAndCustomIDs
	}

	if customIDs != nil && len(texts) != len(customIDs) {
		return nil, ErrMismatchMetadatasAndTextAndCustomIDs
	}

	documents := make([]schema.Document, 0)

	for i := 0; i < len(texts); i++ {
		chunks, err := textSplitter.SplitText(texts[i])
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			// Copy the document metadata
			curMetadata := make(map[string]any, len(metadatas[i]))
			for key, value := range metadatas[i] {
				curMetadata[key] = value
			}

			documents = append(documents, schema.Document{
				PageContent: chunk,
				Metadata:    curMetadata,
				CustomID:    customIDs[i],
			})
		}
	}

	return documents, nil
}

// joinDocs comines two documents with the separator used to split them.
func joinDocs(docs []string, separator string) string {
	return strings.TrimSpace(strings.Join(docs, separator))
}

// mergeSplits merges smaller splits into splits that are closer to the chunkSize.
func mergeSplits(splits []string, separator string, chunkSize int, chunkOverlap int) []string { //nolint:cyclop
	docs := make([]string, 0)
	currentDoc := make([]string, 0)
	total := 0

	for _, split := range splits {
		totalWithSplit := total + len(split)
		if len(currentDoc) != 0 {
			totalWithSplit += len(separator)
		}

		maybePrintWarning(total, chunkSize)
		if totalWithSplit > chunkSize && len(currentDoc) > 0 {
			doc := joinDocs(currentDoc, separator)
			if doc != "" {
				docs = append(docs, doc)
			}

			for shouldPop(chunkOverlap, chunkSize, total, len(split), len(separator), len(currentDoc)) {
				total -= len(currentDoc[0]) //nolint:gosec
				if len(currentDoc) > 1 {
					total -= len(separator)
				}
				currentDoc = currentDoc[1:] //nolint:gosec
			}
		}

		currentDoc = append(currentDoc, split)
		total += len(split)
		if len(currentDoc) > 1 {
			total += len(separator)
		}
	}

	doc := joinDocs(currentDoc, separator)
	if doc != "" {
		docs = append(docs, doc)
	}

	return docs
}

func maybePrintWarning(total, chunkSize int) {
	if total > chunkSize {
		log.Printf(
			"[WARN] created a chunk with size of %v, which is longer then the specified %v\n",
			total,
			chunkSize,
		)
	}
}

// Keep popping if:
//   - the chunk is larger than the chunk overlap
//   - or if there are any chunks and the length is long
func shouldPop(chunkOverlap, chunkSize, total, splitLen, separatorLen, currentDocLen int) bool {
	docsNeededToAddSep := 2
	if currentDocLen < docsNeededToAddSep {
		separatorLen = 0
	}

	return currentDocLen > 0 && (total > chunkOverlap || (total+splitLen+separatorLen > chunkSize && total > 0))
}
