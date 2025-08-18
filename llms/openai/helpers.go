package openai

import "github.com/tmc/langchaingo/llms"

// IsOpenAIModel returns true if the provided model is an OpenAI LLM instance.
// This is useful for type checking when you have a generic llms.Model.
func IsOpenAIModel(model llms.Model) bool {
	_, ok := model.(*LLM)
	return ok
}

// GetOpenAIBaseURL returns the base URL if the model is an OpenAI LLM, otherwise returns empty string.
// This is a convenience function for working with generic llms.Model interfaces.
func GetOpenAIBaseURL(model llms.Model) string {
	if openaiLLM, ok := model.(*LLM); ok {
		return openaiLLM.GetBaseURL()
	}
	return ""
}

// IsOpenAIUsingCustomBaseURL returns true if the model is an OpenAI LLM using a custom base URL.
// Returns false if the model is not an OpenAI LLM or is using the default OpenAI URL.
func IsOpenAIUsingCustomBaseURL(model llms.Model) bool {
	if openaiLLM, ok := model.(*LLM); ok {
		return openaiLLM.IsUsingCustomBaseURL()
	}
	return false
}
