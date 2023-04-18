package agent

import (
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

type Agent struct {
	Llm    llms.LLM
	Chain  chains.Chain
	Tools  []tools.Tool
	Prompt prompts.PromptTemplate
}

func FromLlmAndTools(llm llms.LLM, chain chains.Chain, tools []tools.Tool, prompt prompts.PromptTemplate) Agent {
	return Agent{
		llm,
		chain,
		tools,
		prompt,
	}
}

func GetActionAndInput(input string) (action, actionInput string) {
	fields := strings.Split(input, "\n")
	for _, field := range fields {
		if strings.HasPrefix(field, "Action: ") {
			action = strings.TrimPrefix(field, "Action: ")
		} else if strings.HasPrefix(field, "Action Input: ") {
			actionInput = strings.TrimPrefix(field, "Action Input: ")
		}
	}
	return action, actionInput
}

func GetObservation(action string, actionInput string, tools []tools.Tool) string {
	for _, tool := range tools {
		if tool.Name() == strings.Trim(action, " ") {
			observation := "\nObservation: " + tool.Run(actionInput) + "\n"
			return observation

		}
	}
	return ""
}

func GetFinalAnswer(text string) string {
	finalAnswerPrefix := "Final Answer:"
	startIndex := strings.Index(text, finalAnswerPrefix)

	if startIndex == -1 {
		return ""
	}
	startIndex += len(finalAnswerPrefix)
	text = text[startIndex:]
	trimmed := strings.Join(strings.Fields(text), " ")
	return trimmed
}

type AgentInterface interface {
	Run(input string) string
}
