package textsplitter

import (
	"strings"
)

// RecursiveCharacter is a text splitter that will split texts recursively by different
// characters.
type RecursiveCharacter struct {
	Separators   []string
	ChunkSize    int
	ChunkOverlap int
}

// NewRecursiveCharacter creates a new recursive character splitter with default values. By
// default the separators used are "\n\n", "\n", " " and "". The chunk size is set to 4000
// and chunk overlap is set to 200.
func NewRecursiveCharacter() RecursiveCharacter {
	return RecursiveCharacter{
		Separators:   []string{"\n\n", "\n", " ", ""},
		ChunkSize:    _defaultChunkSize,
		ChunkOverlap: _defaultChunkOverlap,
	}
}

// SplitText splits a text into multiple text.
func (s RecursiveCharacter) SplitText(text string) ([]string, error) {
	finalChunks := make([]string, 0)

	// Find the appropriate separator
	separator := s.Separators[len(s.Separators)-1]
	for _, s := range s.Separators {
		if s == "" {
			separator = s
			break
		}

		if strings.Contains(text, s) {
			separator = s
			break
		}
	}

	splits := strings.Split(text, separator)
	goodSplits := make([]string, 0)

	// Merge the splits, recursively splitting larger texts.
	for _, split := range splits {
		if len(split) < s.ChunkSize {
			goodSplits = append(goodSplits, split)
			continue
		}

		if len(goodSplits) > 0 {
			mergedText := mergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap)

			finalChunks = append(finalChunks, mergedText...)
			goodSplits = make([]string, 0)
		}

		otherInfo, err := s.SplitText(split)
		if err != nil {
			return nil, err
		}
		finalChunks = append(finalChunks, otherInfo...)
	}

	if len(goodSplits) > 0 {
		mergedText := mergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap)
		finalChunks = append(finalChunks, mergedText...)
	}

	return finalChunks, nil
}
