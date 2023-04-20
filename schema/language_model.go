package schema

// LanguageModel is the base class for all language models.
type LanguageModel interface {
	// Take in a list of prompt values and return an LLMResult.
	GeneratePrompt(prompts []PromptValue, stop []string) LLMResult
	// Get the number of tokens present in the text.
	GetNumTokens(text string) int
	// Get the number of tokens in the message.
	GetNumTokensFromMessages(messages []ChatMessage) int
}
