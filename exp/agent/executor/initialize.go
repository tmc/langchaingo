package executor

import (
	"github.com/tmc/langchaingo/exp/agent"
	"github.com/tmc/langchaingo/exp/agent/mrkl"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/llms"
)

// AgentType is a string type representing the type of agent to create.
type AgentType string

const (
	// ZeroShotReactDescription is an AgentType constant that represents
	// the "zeroShotReactDescription" agent type.
	ZeroShotReactDescription AgentType = "zeroShotReactDescription"
)

// Options is a type alias for a map of string keys to any value,
// representing the options for the agent and the executor.
type Options map[string]any

// Option is a function type that can be used to modify the Options.
type Option func(p *Options)

func defaultOptions() Options {
	return Options{
		"verbose":       false,
		"maxIterations": 3,
	}
}

// WithVerbosity is a function that sets the verbosity option for the agent.
func WithVerbosity() Option {
	return func(p *Options) {
		(*p)["verbose"] = true
	}
}

// WithMaxIterations is a function that sets the max iterations for the executor.
func WithMaxIterations(maxIterations int) Option {
	return func(p *Options) {
		(*p)["maxIterations"] = maxIterations
	}
}

// New is a function that creates a new agent with the specified LLM model,
// tools, agent type, and options. It returns an Executor or an error
// if there is any issue during the creation process.
func Initialize(
	llm llms.LLM,
	tools []tools.Tool,
	agentType AgentType,
	opts ...Option,
) (Executor, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}
	var agent agent.Agent

	switch agentType {
	case ZeroShotReactDescription:
		agent = mrkl.NewOneShotAgent(llm, tools, options)
	default:
		return Executor{}, ErrUnknownAgentType
	}

	return Executor{
		Agent:         agent,
		Tools:         tools,
		MaxIterations: options["maxIterations"].(int),
	}, nil
}
