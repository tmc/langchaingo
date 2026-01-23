package llms

import (
	"log"

	"github.com/pkoukk/tiktoken-go"
)

const (
	_tokenApproximation = 4
)

const (
	_gpt35TurboContextSize    = 16385  // gpt-3.5-turbo default context
	_gpt35Turbo16KContextSize = 16385  // gpt-3.5-turbo-16k
	_gpt4ContextSize          = 8192   // gpt-4
	_gpt432KContextSize       = 32768  // gpt-4-32k
	_gpt4TurboContextSize     = 128000 // gpt-4-turbo models
	_gpt4oContextSize         = 128000 // gpt-4o models
	_gpt4oMiniContextSize     = 128000 // gpt-4o-mini
	_textDavinci3ContextSize  = 4097
	_textBabbage1ContextSize  = 2048
	_textAda1ContextSize      = 2048
	_textCurie1ContextSize    = 2048
	_codeDavinci2ContextSize  = 8000
	_codeCushman1ContextSize  = 2048
	_defaultContextSize       = 2048
)

// nolint:gochecknoglobals
var modelToContextSize = map[string]int{
	// GPT-3.5 models
	"gpt-3.5-turbo":      _gpt35TurboContextSize,
	"gpt-3.5-turbo-16k":  _gpt35Turbo16KContextSize,
	"gpt-3.5-turbo-0125": _gpt35TurboContextSize,
	"gpt-3.5-turbo-1106": _gpt35TurboContextSize,
	// GPT-4 models
	"gpt-4":          _gpt4ContextSize,
	"gpt-4-32k":      _gpt432KContextSize,
	"gpt-4-0613":     _gpt4ContextSize,
	"gpt-4-32k-0613": _gpt432KContextSize,
	// GPT-4 Turbo models
	"gpt-4-turbo":            _gpt4TurboContextSize,
	"gpt-4-turbo-preview":    _gpt4TurboContextSize,
	"gpt-4-turbo-2024-04-09": _gpt4TurboContextSize,
	"gpt-4-1106-preview":     _gpt4TurboContextSize,
	"gpt-4-0125-preview":     _gpt4TurboContextSize,
	// GPT-4o models
	"gpt-4o":                 _gpt4oContextSize,
	"gpt-4o-2024-05-13":      _gpt4oContextSize,
	"gpt-4o-2024-08-06":      _gpt4oContextSize,
	"gpt-4o-mini":            _gpt4oMiniContextSize,
	"gpt-4o-mini-2024-07-18": _gpt4oMiniContextSize,
	// Legacy models
	"text-davinci-003": _textDavinci3ContextSize,
	"text-curie-001":   _textCurie1ContextSize,
	"text-babbage-001": _textBabbage1ContextSize,
	"text-ada-001":     _textAda1ContextSize,
	"code-davinci-002": _codeDavinci2ContextSize,
	"code-cushman-001": _codeCushman1ContextSize,
}

// GetModelContextSize gets the max number of tokens for a language model. If the model
// name isn't recognized the default value 2048 is returned.
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
