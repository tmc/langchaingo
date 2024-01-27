package textsplitter

import (
	"log"
	"strings"

	"github.com/tmc/langchaingo/schema"
)

// SplitDocuments splits documents using a textsplitter.
func SplitDocuments(textSplitter TextSplitter, documents []schema.Document) ([]schema.Document, error) {
	splittedDocuments := make([]schema.Document, 0)

	for i := 0; i < len(documents); i++ {
		chunks, err := textSplitter.SplitText(documents[i].PageContent)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			metadata := map[string]any{}
			if documents[i].Metadata != nil {
				metadata = documents[i].Metadata // This is here for ensuring backwards compatibility.
			}
			splittedDocuments = append(splittedDocuments, schema.Document{
				PageContent: chunk,
				Metadata:    metadata,
				CustomID:    documents[i].CustomID,
			})
		}
	}

	return splittedDocuments, nil
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
