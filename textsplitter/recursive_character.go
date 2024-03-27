package textsplitter

import (
	"strings"
)

// RecursiveCharacter is a text splitter that will split texts recursively by different
// characters.
type RecursiveCharacter struct {
	Separators    []string
	ChunkSize     int
	ChunkOverlap  int
	LenFunc       func(string) int
	KeepSeparator bool
}

// NewRecursiveCharacter creates a new recursive character splitter with default values. By
// default, the separators used are "\n\n", "\n", " " and "". The chunk size is set to 4000
// and chunk overlap is set to 200.
func NewRecursiveCharacter(opts ...Option) RecursiveCharacter {
	options := DefaultOptions()
	for _, o := range opts {
		o(&options)
	}

	s := RecursiveCharacter{
		Separators:    options.Separators,
		ChunkSize:     options.ChunkSize,
		ChunkOverlap:  options.ChunkOverlap,
		LenFunc:       options.LenFunc,
		KeepSeparator: options.KeepSeparator,
	}

	return s
}

// SplitText splits a text into multiple text.
func (s RecursiveCharacter) SplitText(text string) ([]string, error) {
	return s.splitText(text, s.Separators)
}

// addSeparatorInSplits adds the separator in each of splits.
func (s RecursiveCharacter) addSeparatorInSplits(splits []string, separator string) []string {
	splitsWithSeparator := make([]string, 0, len(splits))
	for i, s := range splits {
		if i > 0 {
			s = separator + s
		}
		splitsWithSeparator = append(splitsWithSeparator, s)
	}
	return splitsWithSeparator
}

func (s RecursiveCharacter) splitText(text string, separators []string) ([]string, error) {
	finalChunks := make([]string, 0)

	// Find the appropriate separator.
	separator := separators[len(separators)-1]
	newSeparators := []string{}
	for i, c := range separators {
		if c == "" || strings.Contains(text, c) {
			separator = c
			newSeparators = separators[i+1:]
			break
		}
	}

	splits := strings.Split(text, separator)
	if s.KeepSeparator {
		splits = s.addSeparatorInSplits(splits, separator)
		separator = ""
	}
	goodSplits := make([]string, 0)

	// Merge the splits, recursively splitting larger texts.
	for _, split := range splits {
		if s.LenFunc(split) < s.ChunkSize {
			goodSplits = append(goodSplits, split)
			continue
		}

		if len(goodSplits) > 0 {
			mergedText := mergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap, s.LenFunc)

			finalChunks = append(finalChunks, mergedText...)
			goodSplits = make([]string, 0)
		}

		if len(newSeparators) == 0 {
			finalChunks = append(finalChunks, split)
		} else {
			otherInfo, err := s.splitText(split, newSeparators)
			if err != nil {
				return nil, err
			}
			finalChunks = append(finalChunks, otherInfo...)
		}
	}

	if len(goodSplits) > 0 {
		mergedText := mergeSplits(goodSplits, separator, s.ChunkSize, s.ChunkOverlap, s.LenFunc)
		finalChunks = append(finalChunks, mergedText...)
	}

	return finalChunks, nil
}
