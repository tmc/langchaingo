package llms

import (
	"log"

	"github.com/pkoukk/tiktoken-go"
)

const _tokenApproximation = 4

// GetModelContextSize gets the max number of tokens for a model.
func GetModelContextSize(model string) int {
	switch model {
	case "gpt-3.5-turbo":
		return 4096
	case "gpt-4-32k":
		return 32768
	case "gpt-4":
		return 8192
	case "text-davinci-003":
		return 4097
	case "text-curie-001":
		return 2048
	case "text-babbage-001":
		return 2048
	case "text-ada-001":
		return 2048
	case "code-davinci-002":
		return 8000
	case "code-cushman-001":
		return 2048
	case "text-bison":
		return 8192
	case "chat-bison":
		return 4096
	default:
		return 4097
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
