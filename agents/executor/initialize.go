package executor

import (
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/agents/mrkl"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// AgentType is a string type representing the type of agent to create.
type AgentType string

const (
	// ZeroShotReactDescription is an AgentType constant that represents
	// the "zeroShotReactDescription" agent type.
	ZeroShotReactDescription AgentType = "zeroShotReactDescription"
)

// options is a type alias for a map of string keys to any value,
// representing the options for the agent and the executor.
type options map[string]any

// Option is a function type that can be used to modify the creation of the
// executor and agent.
type Option func(p *options)

func defaultOptions() options {
	return options{
		"verbose":       false,
		"maxIterations": 3,
	}
}

// WithVerbosity is a function that sets the verbosity option for the agent.
func WithVerbosity() Option {
	return func(p *options) {
		(*p)["verbose"] = true
	}
}

// WithMaxIterations is a function that sets the max iterations for the executor.
func WithMaxIterations(maxIterations int) Option {
	return func(p *options) {
		(*p)["maxIterations"] = maxIterations
	}
}

// Initialize is a function that creates a new executor with the specified LLM
// model, tools, agent type, and options. It returns an Executor or an error
// if there is any issues during the creation process.
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
	var agent agents.Agent

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
