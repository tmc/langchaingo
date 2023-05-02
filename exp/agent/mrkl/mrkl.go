package mrkl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/exp/agent/executor"
	"github.com/tmc/langchaingo/exp/chains"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/logger"
	"github.com/tmc/langchaingo/schema"
)

// OneShotZeroAgent is a struct that represents an agent responsible for executing a query
// and returning the result using the LLM model, tools, and an internal chain.
type OneShotZeroAgent struct {
	llm        llms.LLM
	query      string
	chain      chains.Chain
	tools      []tools.Tool
	maxRetries int
}

// OneShotZeroAgentOptions is a type alias for a map of string keys to any value,
// representing the options for the OneShotZeroAgent.
type OneShotZeroAgentOptions map[string]any

var _ executor.AgentExecutor = (*OneShotZeroAgent)(nil)

const FinalAnswerAction = "Final Answer:"

func checkOptions(opts OneShotZeroAgentOptions) OneShotZeroAgentOptions {
	if _, ok := opts["maxRetries"].(int); !ok {
		opts["maxRetries"] = 3
	}
	return opts
}

// NewOneShotAgent creates a new OneShotZeroAgent with the given LLM model, tools,
// and options. It returns a pointer to the created agent and an error if there is any
// issue during the creation process.
func NewOneShotAgent(llm llms.LLM, tools []tools.Tool, opts map[string]any) (*OneShotZeroAgent, error) {
	// Validate opts
	opts = checkOptions(opts)

	return &OneShotZeroAgent{
		llm:        llm,
		query:      "",
		chain:      chains.NewLLMChain(llm, createPrompt()),
		tools:      tools,
		maxRetries: opts["maxRetries"].(int),
	}, nil
}

// Run is an implementation of the AgentExecutor interface. It takes a query as input
// and executes it, returning an AgentFinish object containing the result, or an error
// if the execution fails.
func (a *OneShotZeroAgent) Run(ctx context.Context, query string) (*schema.AgentFinish, error) {
	var attempts int
	a.query = query

	// Call the chain
	resp, _ := a.chain.Call(ctx, map[string]interface{}{
		"today":             time.Now().Format("January 02, 2006"),
		"tool_names":        toolNames(a.tools),
		"tool_descriptions": toolDescriptions(a.tools),
		"input":             a.query,
		"agent_scratchpad":  "",
		"stop":              []string{"\nObservation:", "\n\tObservation:"},
	})

	// Validate the response
	output, ok := resp["text"].(string)
	if !ok {
		return nil, errors.New("Agent did not return a string")
	}

	for output != "" || attempts < a.maxRetries {
		var err error
		action, finish := a.plan(output)
		if finish != nil {
			return finish, nil
		}
		output, err = a.nextStep(ctx, *action)
		if err != nil {
			return nil, err
		}

		attempts++
	}

	return nil, fmt.Errorf("Agent did not finish after %d attempts", attempts)
}

func (a *OneShotZeroAgent) nextStep(ctx context.Context, action schema.AgentAction) (string, error) {
	var scratchpad []string
	// Perform your desired operation with the text value
	observation, err := runTool(action.Tool, action.ToolInput.(string), &a.tools)
	if err != nil {
		return "", err
	}
	scratchpad = append(scratchpad, action.Log+observation)
	logger.AgentThought(getCurrentScratchpad(scratchpad))

	// Update resp using a.chain.Call()
	newResp, err := a.chain.Call(ctx, map[string]interface{}{
		"input":            a.query,
		"agent_scratchpad": strings.Join(scratchpad, "\n"),
		"stop":             []string{"\nObservation:", "\n\tObservation:"},
	})
	if err != nil {
		return "", err
	}

	// Validate the response from the chain
	if _, ok := newResp["text"].(string); !ok {
		return "", errors.New("Agent did not return a string")
	}

	// Use the updated response in the next iteration
	return newResp["text"].(string), nil
}

func (a *OneShotZeroAgent) plan(info string) (*schema.AgentAction, *schema.AgentFinish) {
	action := getAgentAction(info)
	if aswer := getFinalAnswer(action.Log); aswer != "" {
		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				"answer": aswer,
			},
			Log: action.Log,
		}
	}
	return &action, nil
}

func getCurrentScratchpad(scratchpad []string) string {
	if len(scratchpad) == 0 {
		return ""
	}
	lastThought := scratchpad[len(scratchpad)-1]
	if len(scratchpad) == 1 {
		lastThought = lastThought
	}

	// The first line of the scratchpad string should be the agent's thought.
	// The second line of the scratchpad string should be the action that the agent chooses.
	// The third line of the scratchpad string should be the action input.
	// The fourth line of the scratchpad string (if any) should be the observation.
	// YMMV depending on how well your particular LLM adheres to the initial instructions.

	return lastThought
}

func getFinalAnswer(text string) string {
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

func getAgentAction(input string) schema.AgentAction {
	var agentAction schema.AgentAction
	agentAction.Log = input
	fields := strings.Split(input, "\n")
	for _, field := range fields {
		if strings.HasPrefix(field, "Action: ") {
			agentAction.Tool = strings.TrimPrefix(field, "Action: ")
		} else if strings.HasPrefix(field, "Action Input: ") {
			agentAction.ToolInput = strings.TrimPrefix(field, "Action Input: ")
		}
	}
	return agentAction
}

func runTool(action string, actionInput string, tools *[]tools.Tool) (string, error) {
	// Sanitize the action
	action = strings.ToLower(strings.TrimSpace(action))

	// Sanitize the action input
	actionInput = strings.TrimSpace(actionInput)

	// Find the tool that matches the action
	var observation string
	for _, tool := range *tools {
		if tool.Name != action {
			continue
		}

		// Run the tool
		toolOutput, err := tool.Run(actionInput)
		if err != nil {
			return "", err
		}

		// Add the tool's output to the observation
		observation = "\nObservation: " + toolOutput + "\n"
		break
	}
	return observation, nil
}
