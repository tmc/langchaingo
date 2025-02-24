package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/i18n"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// ConversationalAgent is a struct that represents an agent responsible for deciding
// what to do or give the final output if the task is finished given a set of inputs
// and previous steps taken.
//
// Other agents are often optimized for using tools to figure out the best response,
// which is not ideal in a conversational setting where you may want the agent to be
// able to chat with the user as well.
type ConversationalAgent struct {
	// Chain is the chain used to call with the values. The chain should have an
	// input called "agent_scratchpad" for the agent to put its thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
	// FinalAnswer is the final answer in various languages.
	FinalAnswer string
	// Lang is the language the prompt will use.
	Lang i18n.Lang
	// CallbacksHandler is the handler for callbacks.
	CallbacksHandler callbacks.Handler
}

var _ Agent = (*ConversationalAgent)(nil)

func NewConversationalAgent(llm llms.Model, tools []tools.Tool, opts ...Option) *ConversationalAgent {
	options := conversationalDefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	options.loadConversationalTranslatable()

	return &ConversationalAgent{
		Chain: chains.NewLLMChain(
			llm,
			options.getConversationalPrompt(tools),
			chains.WithCallback(options.callbacksHandler),
		),
		Tools:            tools,
		OutputKey:        options.outputKey,
		FinalAnswer:      i18n.AgentsMustPhrase(options.lang, "conversational final answer"),
		Lang:             options.lang,
		CallbacksHandler: options.callbacksHandler,
	}
}

// Plan decides what action to take or returns the final result of the input.
func (a *ConversationalAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}

	fullInputs["agent_scratchpad"] = constructScratchPad(intermediateSteps, a.Lang)

	var stream func(ctx context.Context, chunk []byte) error

	if a.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk []byte) error {
			a.CallbacksHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	output, err := chains.Predict(
		ctx,
		a.Chain,
		fullInputs,
		chains.WithStopWords([]string{
			fmt.Sprintf("\n%s", i18n.AgentsMustPhrase(a.Lang, "observation")),
			fmt.Sprintf("\n\t%s", i18n.AgentsMustPhrase(a.Lang, "observation")),
		}),
		chains.WithStreamingFunc(stream),
	)
	if err != nil {
		return nil, nil, err
	}

	return a.parseOutput(output)
}

func (a *ConversationalAgent) GetInputKeys() []string {
	chainInputs := a.Chain.GetInputKeys()

	// Remove inputs given in plan.
	agentInput := make([]string, 0, len(chainInputs))
	for _, v := range chainInputs {
		if v == "agent_scratchpad" {
			continue
		}
		agentInput = append(agentInput, v)
	}

	return agentInput
}

func (a *ConversationalAgent) GetOutputKeys() []string {
	return []string{a.OutputKey}
}

func (a *ConversationalAgent) GetTools() []tools.Tool {
	return a.Tools
}

func constructScratchPad(steps []schema.AgentStep, lang i18n.Lang) string {
	var scratchPad string
	if len(steps) > 0 {
		for _, step := range steps {
			scratchPad += step.Action.Log
			scratchPad += fmt.Sprintf("\n%s %s", i18n.AgentsMustPhrase(lang, "observation"), step.Observation)
		}
		scratchPad += fmt.Sprintf("\n%s", i18n.AgentsMustPhrase(lang, "thought"))
	}

	return scratchPad
}

func (a *ConversationalAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	if strings.Contains(output, a.FinalAnswer) {
		splits := strings.Split(output, a.FinalAnswer)

		finishAction := &schema.AgentFinish{
			ReturnValues: map[string]any{
				a.OutputKey: splits[len(splits)-1],
			},
			Log: output,
		}

		return nil, finishAction, nil
	}

	action, actionInput := i18n.AgentsMustPhrase(a.Lang, "action"),
		i18n.AgentsMustPhrase(a.Lang, "action input")
	r := regexp.MustCompile(fmt.Sprintf(`%s (.*?)[\n]*%s (.*)`, action, actionInput))
	matches := r.FindStringSubmatch(output)
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", ErrUnableToParseOutput, output)
	}

	return []schema.AgentAction{
		{Tool: strings.TrimSpace(matches[1]), ToolInput: strings.TrimSpace(matches[2]), Log: output},
	}, nil, nil
}

func createConversationalPrompt(tools []tools.Tool, prefix, instructions, suffix string) prompts.PromptTemplate {
	template := strings.Join([]string{prefix, instructions, suffix}, "\n\n")

	return prompts.PromptTemplate{
		Template:       template,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: []string{"input", "agent_scratchpad"},
		PartialVariables: map[string]any{
			"tool_names":        toolNames(tools),
			"tool_descriptions": toolDescriptions(tools),
			"history":           "",
		},
	}
}
