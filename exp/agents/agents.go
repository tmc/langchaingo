// Package agents defines the types for langchaingo Agetns.
package agents

import ()

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
