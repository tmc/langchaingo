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

func (l LogHandler) HandleLLMGenerateContentStart(_ context.Context, ms []llms.MessageContent) {
	fmt.Println("Entering LLM with messages:")
	for _, m := range ms {
		// TODO: Implement logging of other content types
		var buf strings.Builder
		for _, t := range m.Parts {
			if t, ok := t.(llms.TextContent); ok {
				buf.WriteString(t.Text)
			}
		}
		fmt.Println("Role:", m.Role)
		fmt.Println("Text:", buf.String())
	}
}

func (l LogHandler) HandleLLMGenerateContentEnd(_ context.Context, res *llms.ContentResponse) {
	fmt.Println("Exiting LLM with response:")
	for _, c := range res.Choices {
		if c.Content != "" {
			fmt.Println("Content:", c.Content)
		}
		if c.StopReason != "" {
			fmt.Println("StopReason:", c.StopReason)
		}
		if len(c.GenerationInfo) > 0 {
			fmt.Println("GenerationInfo:")
			for k, v := range c.GenerationInfo {
				fmt.Printf("%20s: %v\n", k, v)
			}
		}
		if c.FuncCall != nil {
			fmt.Println("FuncCall: ", c.FuncCall.Name, c.FuncCall.Arguments)
		}
	}
}

func (l LogHandler) HandleStreamingFunc(_ context.Context, chunk []byte) {
	fmt.Println(string(chunk))
}

func (l LogHandler) HandleText(_ context.Context, text string) {
	fmt.Println(text)
}

func (l LogHandler) HandleLLMStart(_ context.Context, prompts []string) {
	fmt.Println("Entering LLM with prompts:", prompts)
}

func (l LogHandler) HandleLLMError(_ context.Context, err error) {
	fmt.Println("Exiting LLM with error:", err)
}

func (l LogHandler) HandleChainStart(_ context.Context, inputs map[string]any) {
	fmt.Println("Entering chain with inputs:", formatChainValues(inputs))
}

func (l LogHandler) HandleChainEnd(_ context.Context, outputs map[string]any) {
	fmt.Println("Exiting chain with outputs:", formatChainValues(outputs))
}

func (l LogHandler) HandleChainError(_ context.Context, err error) {
	fmt.Println("Exiting chain with error:", err)
}

func (l LogHandler) HandleToolStart(_ context.Context, input string) {
	fmt.Println("Entering tool with input:", removeNewLines(input))
}

func (l LogHandler) HandleToolEnd(_ context.Context, output string) {
	fmt.Println("Exiting tool with output:", removeNewLines(output))
}

func (l LogHandler) HandleToolError(_ context.Context, err error) {
	fmt.Println("Exiting tool with error:", err)
}

func (l LogHandler) HandleAgentAction(_ context.Context, action schema.AgentAction) {
	fmt.Println("Agent selected action:", formatAgentAction(action))
}

func (l LogHandler) HandleAgentFinish(_ context.Context, finish schema.AgentFinish) {
	fmt.Printf("Agent finish: %v \n", finish)
}

func (l LogHandler) HandleRetrieverStart(_ context.Context, query string) {
	fmt.Println("Entering retriever with query:", removeNewLines(query))
}

func (l LogHandler) HandleRetrieverEnd(_ context.Context, query string, documents []schema.Document) {
	fmt.Println("Exiting retriever with documents for query:", documents, query)
}

func formatChainValues(values map[string]any) string {
	output := ""
	for key, value := range values {
		output += fmt.Sprintf("\"%s\" : \"%s\", ", removeNewLines(key), removeNewLines(value))
	}

	return output
}

func formatAgentAction(action schema.AgentAction) string {
	return fmt.Sprintf("\"%s\" with input \"%s\"", removeNewLines(action.Tool), removeNewLines(action.ToolInput))
}

func removeNewLines(s any) string {
	return strings.ReplaceAll(fmt.Sprint(s), "\n", " ")
}
