// Package agents defines the types for langchaingo Agetns.
package agents

import (
	"errors"

	agentsExecutor "github.com/tmc/langchaingo/exp/agents/agentExecutor"
	"github.com/tmc/langchaingo/exp/agents/mrkl"
	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/llms"
)

type AgentType string

const (
	ZeroShotReactDescription AgentType = "zeroShotReactDescription"
)

type AgentOptions map[string]any
type Options func(p *AgentOptions)

func defaultOptions() AgentOptions {
	return AgentOptions{
		"verbose": false,
	}
}

func WithVerbosity() Options {
	return func(p *AgentOptions) {
		(*p)["verbose"] = true
	}
}

func InitializeAgent(
	llm llms.LLM,
	tools []tools.Tool,
	agentType AgentType,
	opts ...Options,

) (agentsExecutor.AgentExecutor, error) {
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
