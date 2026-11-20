package textsplitter

import (
	"errors"
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

const (
	// nolint:gosec
	_defaultTokenModelName    = "gpt-3.5-turbo"
	_defaultTokenEncoding     = "cl100k_base"
	_defaultTokenChunkSize    = 512
	_defaultTokenChunkOverlap = 100
)

// ErrInvalidChunkConfig is returned when a splitter's chunk size or chunk overlap
// cannot produce chunks.
var ErrInvalidChunkConfig = errors.New("invalid chunk configuration")

// TokenSplitter is a text splitter that will split texts by tokens.
type TokenSplitter struct {
	ChunkSize         int
	ChunkOverlap      int
	ModelName         string
	EncodingName      string
	AllowedSpecial    []string
	DisallowedSpecial []string
}

func NewTokenSplitter(opts ...Option) TokenSplitter {
	options := DefaultOptions()
	for _, o := range opts {
		o(&options)
	}

	s := TokenSplitter{
		ChunkSize:         options.ChunkSize,
		ChunkOverlap:      options.ChunkOverlap,
		ModelName:         options.ModelName,
		EncodingName:      options.EncodingName,
		AllowedSpecial:    options.AllowedSpecial,
		DisallowedSpecial: options.DisallowedSpecial,
	}

	return s
}

// SplitText splits a text into multiple text.
func (s TokenSplitter) SplitText(text string) ([]string, error) {
	// Each chunk must move past the previous one, or splitting never ends.
	if s.ChunkSize <= 0 {
		return nil, fmt.Errorf("%w: chunk size %d must be positive", ErrInvalidChunkConfig, s.ChunkSize)
	}
	if s.ChunkOverlap < 0 || s.ChunkOverlap >= s.ChunkSize {
		return nil, fmt.Errorf("%w: chunk overlap %d must be at least 0 and less than chunk size %d",
			ErrInvalidChunkConfig, s.ChunkOverlap, s.ChunkSize)
	}

	// Get the tokenizer
	var tk *tiktoken.Tiktoken
	var err error
	if s.EncodingName != "" {
		tk, err = tiktoken.GetEncoding(s.EncodingName)
	} else {
		tk, err = tiktoken.EncodingForModel(s.ModelName)
	}
	if err != nil {
		return nil, fmt.Errorf("tiktoken.GetEncoding: %w", err)
	}
	texts := s.splitText(text, tk)

	return texts, nil
}

func (s TokenSplitter) splitText(text string, tk *tiktoken.Tiktoken) []string {
	splits := make([]string, 0)
	inputIDs := tk.Encode(text, s.AllowedSpecial, s.DisallowedSpecial)

	startIdx := 0
	curIdx := len(inputIDs)
	if startIdx+s.ChunkSize < curIdx {
		curIdx = startIdx + s.ChunkSize
	}
	for startIdx < len(inputIDs) {
		chunkIDs := inputIDs[startIdx:curIdx]
		splits = append(splits, tk.Decode(chunkIDs))
		// Stop once a chunk reaches the end of the text: a later chunk would only
		// repeat tokens that are already in this one.
		if curIdx == len(inputIDs) {
			break
		}
		startIdx += s.ChunkSize - s.ChunkOverlap
		curIdx = startIdx + s.ChunkSize
		if curIdx > len(inputIDs) {
			curIdx = len(inputIDs)
		}
	}
	return splits
}
