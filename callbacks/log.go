package callbacks

import (
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LogHandler is a callback handler that prints to the standard output.
type LogHandler struct{}

var _ CallbackHandler = LogHandler{}

func (l LogHandler) HandleText(text string) {
	fmt.Println(text)
}

func (l LogHandler) HandleLLMStart(prompts []string) {
	fmt.Println("Entering LLM with prompts", prompts)
}

func (l LogHandler) HandleLLMEnd(output llms.LLMResult) {
	results := ""
	for i := 0; i < len(output.Generations); i++ {
		for j := 0; j < len(output.Generations[i]); j++ {
			results += output.Generations[i][j].Text + "\n"
		}
	}
	fmt.Println("Exiting LLM with results: \n", results)
}

func (l LogHandler) HandleChainStart(inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", inputs)
}

func (l LogHandler) HandleChainEnd(outputs map[string]any) {
	fmt.Println("Exiting chain with outputs:", outputs)
}

func (l LogHandler) HandleToolStart(input string) {
	fmt.Println("Entering tool with input:", input)
}

func (l LogHandler) HandleToolEnd(output string) {
	fmt.Println("Exiting tool with output:", output)
}

func (l LogHandler) HandleAgentAction(action schema.AgentAction) {
	fmt.Println("Agent selected action:", action)
}

func (l LogHandler) HandleRetrieverStart(query string) {
	fmt.Println("Entering retriever with query:", query)
}

func (l LogHandler) HandleRetrieverEnd(documents []schema.Document) {
	fmt.Println("Exiting retirer with documents:", documents)
}
