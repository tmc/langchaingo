//nolint:forbidigo
package callbacks

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LogHandler is a callback handler that prints to the standard output.
type LogHandler struct{}

var _ Handler = LogHandler{}

func (l LogHandler) HandleText(_ context.Context, text string) {
	fmt.Println(text)
}

func (l LogHandler) HandleLLMStart(_ context.Context, prompts []string) {
	fmt.Println("Entering LLM with prompts:", prompts)
}

func (l LogHandler) HandleLLMEnd(_ context.Context, output llms.LLMResult) {
	fmt.Println("Exiting LLM with results:", formatLLMResult(output))
}

func (l LogHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", formatChainValues(inputs))
}

func (l LogHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	fmt.Println("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l LogHandler) HandleToolStart(_ context.Context, input string) {
	fmt.Println("Entering tool with input:", removeNewLines(input))
}

func (l LogHandler) HandleToolEnd(_ context.Context, output string) {
	fmt.Println("Exiting tool with output:", removeNewLines(output))
}

func (l LogHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	fmt.Println("Agent selected action:", formatAgentAction(action))
}

func (l LogHandler) HandleRetrieverStart(_ context.Context, query string) {
	fmt.Println("Entering retriever with query:", removeNewLines(query))
}

func (l LogHandler) HandleRetrieverEnd(_ context.Context, documents []schema.Document) {
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
