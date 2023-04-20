package prompts

// ExampleSelector is an interface for example selectors. It is equivalent to
// BaseExampleSelector in langchain and langchainjs.
type ExampleSelector interface {
	AddExample(example map[string]string) string
	SelectExamples(inputVariables map[string]string) []map[string]string
}
