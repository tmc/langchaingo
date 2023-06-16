package llms

import (
	"log"

	"github.com/pkoukk/tiktoken-go"
)

const (
	_tokenApproximation = 4
	_defaultContextSize = 4096
)

var modelToContextSize = map[string]int{
	"gpt-3.5-turbo":    4096,
	"gpt-4-32k":        32768,
	"gpt-4":            8192,
	"text-davinci-003": 4097,
	"text-curie-001":   2048,
	"text-babbage-001": 2048,
	"text-ada-001":     2048,
	"code-davinci-002": 8000,
	"code-cushman-001": 2048,
	"text-bison":       8192,
	"chat-bison":       4096,
}

// ModelContextSize gets the max number of tokens for a model. If the model
// name isn't recognized the default value 4097 is returned.
func GetModelContextSize(model string) int {
	contextSize, ok := modelToContextSize[model]
	if !ok {
		return _defaultContextSize
	}
	return contextSize
}

// CountTokens gets the number of tokens the text contains.
func CountTokens(model, text string) int {
	e, err := tiktoken.EncodingForModel(model)
	if err != nil {
		e, err = tiktoken.GetEncoding("gpt2")
		if err != nil {
			log.Printf("[WARN] Failed to calculate number of tokens for model, falling back to approximate count")
			return len([]rune(text)) / _tokenApproximation
		}
	}
	return len(e.Encode(text, nil, nil))
}

// CalculateMaxTokens calculates the max number of tokens that could be added to a text.
func CalculateMaxTokens(model, text string) int {
	return GetModelContextSize(model) - CountTokens(model, text)
}
