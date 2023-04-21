// Package agents defines the types for langchaingo Agent.
package agent

import (
	"errors"

	"github.com/tmc/langchaingo/exp/agent/executor"
	"github.com/tmc/langchaingo/exp/agent/mrkl"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/llms"
)

// AgentType is a string type representing the type of agent to create.
type AgentType string

const (
	// ZeroShotReactDescription is an AgentType constant that represents the "zeroShotReactDescription" agent type.
	ZeroShotReactDescription AgentType = "zeroShotReactDescription"
)

// AgentOption is a type alias for a map of string keys to any value, representing the options for the agent.
type AgentOption map[string]any

// Options is a function type that can be used to modify the AgentOption.
type Options func(p *AgentOption)

func defaultOptions() AgentOption {
	return AgentOption{
		"verbose": false,
	}
}

// WithVerbosity is a function that sets the verbosity option for the agent.
func WithVerbosity() Options {
	return func(p *AgentOption) {
		(*p)["verbose"] = true
	}
}

// New is a function that creates a new agent with the specified LLM model, tools, agent type, and options.
// It returns an AgentExecutor interface or an error if there is any issue during the creation process.
func New(
	llm llms.LLM,
	tools []tools.Tool,
	agentType AgentType,
	opts ...Options,

) (executor.AgentExecutor, error) {

	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	switch agentType {
	case ZeroShotReactDescription:
		return mrkl.NewOneShotAgent(llm, tools, options)
	default:
		return nil, errors.New("Unknown agent type")
	}
}
