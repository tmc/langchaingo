package callbacks

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LogHandler is a callback handler that prints to the standard output.
type LogHandler struct{}

var _ Handler = LogHandler{}

func (l LogHandler) HandleText(text string) {
	fmt.Println(text)
}

func (l LogHandler) HandleLLMStart(prompts []string) {
	fmt.Println("Entering LLM with prompts:", prompts)
}

func (l LogHandler) HandleLLMEnd(output llms.LLMResult) {
	fmt.Println("Exiting LLM with results:", formatLLMResult(output))
}

func (l LogHandler) HandleChainStart(inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", formatChainValues(inputs))
}

func (l LogHandler) HandleChainEnd(outputs map[string]any) {

	fmt.Println("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l LogHandler) HandleToolStart(input string) {
	fmt.Println("Entering tool with input:", removeNewLines(input))
}

func (l LogHandler) HandleToolEnd(output string) {
	fmt.Println("Exiting tool with output:", removeNewLines(output))
}

func (l LogHandler) HandleAgentAction(action schema.AgentAction) {
	fmt.Println("Agent selected action:", formatAgentAction(action))
}

func (l LogHandler) HandleRetrieverStart(query string) {
	fmt.Println("Entering retriever with query:", removeNewLines(query))
}

func (l LogHandler) HandleRetrieverEnd(documents []schema.Document) {
	fmt.Println("Exiting retirer with documents:", documents)
}

func formatChainValues(values map[string]any) string {
	output := ""
	for key, value := range values {
		output += fmt.Sprintf("\"%s\" : \"%s\", ", removeNewLines(key), removeNewLines(value))
	}

	return output
}

func formatLLMResult(output llms.LLMResult) string {
	results := "[ "
	for i := 0; i < len(output.Generations); i++ {
		for j := 0; j < len(output.Generations[i]); j++ {
			results += output.Generations[i][j].Text
		}
	}

	return results + " ]"
}

func formatAgentAction(action schema.AgentAction) string {
	return fmt.Sprintf("\"%s\" with input \"%s\"", removeNewLines(action.Tool), removeNewLines(action.ToolInput))
}

func removeNewLines(s any) string {
	return strings.ReplaceAll(fmt.Sprint(s), "\n", " ")
}
