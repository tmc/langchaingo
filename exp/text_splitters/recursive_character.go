package text_splitters

import (
	"strings"
)

type RecursiveCharactersSplitter struct {
	Separators   []string
	ChunkSize    int
	ChunkOverlap int
}

func NewRecursiveCharactersSplitter() RecursiveCharactersSplitter {
	return RecursiveCharactersSplitter{
		Separators:   []string{"\n\n", "\n", " ", ""},
		ChunkSize:    1000,
		ChunkOverlap: 200,
	}
}

func (s RecursiveCharactersSplitter) SplitText(text string) ([]string, error) {

	//Find the appropriate separator
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

	//Split the text
	splits := strings.Split(text, separator)

	//Merge
	finalChunks := make([]string, 0)
	goodSplits := make([]string, 0)

	for _, split := range splits {

		if len(split) < s.ChunkSize {
			goodSplits = append(goodSplits, split)
			continue
		}

		if len(goodSplits) > 0 {
			mergedText := MergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap)

			finalChunks = append(finalChunks, mergedText...)
			goodSplits = make([]string, 0)
		}

		otherInfo, err := s.SplitText(split)
		if err != nil {
			return []string{}, err
		}

		finalChunks = append(finalChunks, otherInfo...)
	}

	if len(goodSplits) > 0 {
		mergedText := MergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap)
		finalChunks = append(finalChunks, mergedText...)
	}

	return finalChunks, nil
}
