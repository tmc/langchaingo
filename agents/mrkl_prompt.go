package agents

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

const (
	_defaultMrklPrefix = `Today is {{.today}}.
Answer the following questions as best you can. You have access to the following tools:

{{.tool_descriptions}}`

	_defaultMrklFormatInstructions = `Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [ {{.tool_names}} ]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question`

	_defaultMrklSuffix = `Begin!

Question: {{.input}}
{{.agent_scratchpad}}`
)

type mrklTemplateBase struct {
	Template       string
	InputVariables []string
}

func createMRKLPrompt(tools []tools.Tool, prefix, instructions, suffix mrklTemplateBase) prompts.PromptTemplate {
	template := strings.Join([]string{prefix.Template, instructions.Template, suffix.Template}, "\n\n")
	inputVariables := make([]string, 0, len(prefix.InputVariables)+
		len(instructions.InputVariables)+
		len(suffix.InputVariables))
	inputVariables = append(inputVariables, prefix.InputVariables...)
	inputVariables = append(inputVariables, instructions.InputVariables...)
	inputVariables = append(inputVariables, suffix.InputVariables...)
	return prompts.PromptTemplate{
		Template:       template,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: inputVariables,
		PartialVariables: map[string]any{
			"tool_names":        toolNames(tools),
			"tool_descriptions": toolDescriptions(tools),
		},
	}
}

func toolNames(tools []tools.Tool) string {
	var tn strings.Builder
	for i, tool := range tools {
		if i > 0 {
			tn.WriteString(", ")
		}
		tn.WriteString(tool.Name())
	}

	return tn.String()
}

func toolDescriptions(tools []tools.Tool) string {
	var ts strings.Builder
	for _, tool := range tools {
		ts.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}

	return ts.String()
}
