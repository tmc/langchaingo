package llms

import (
	"log"

	"github.com/pkoukk/tiktoken-go"
)

const (
	_tokenApproximation = 4
)

const (
	_gpt35TurboContextSize   = 4096
	_gpt432KContextSize      = 32768
	_gpt4ContextSize         = 8192
	_textDavinci3ContextSize = 4097
	_textBabbage1ContextSize = 2048
	_textAda1ContextSize     = 2048
	_textCurie1ContextSize   = 2048
	_codeDavinci2ContextSize = 8000
	_codeCushman1ContextSize = 2048
	_textBisonContextSize    = 2048
	_chatBisonContextSize    = 2048
	_defaultContextSize      = 2048
)

// ModelContextSize gets the max number of tokens for a language model. If the model
// name isn't recognized the default value 4097 is returned.
func GetModelContextSize(model string) int {
	switch model {
	case "gpt-3.5-turbo":
		return _gpt35TurboContextSize
	case "gpt-4-32k":
		return _gpt432KContextSize
	case "gpt-4":
		return _gpt4ContextSize
	case "text-davinci-003":
		return _textDavinci3ContextSize
	case "text-curie-001":
		return _textCurie1ContextSize
	case "text-babbage-001":
		return _textBabbage1ContextSize
	case "text-ada-001":
		return _textBabbage1ContextSize
	case "code-davinci-002":
		return _codeDavinci2ContextSize
	case "code-cushman-001":
		return _codeCushman1ContextSize
	default:
		return _defaultContextSize
	}
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
