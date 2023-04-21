// Package agent provides an implementation of the AgentExecutor interface for
// langchaingo agents. The package includes the necessary types, structures,
// and functions to create and configure agents.
//
// The agent package provides the New function to create a new agent, and
// supports different agent types and options to customize the agent's behavior.
//
// AgentType is a string type representing the type of agent to create.
// The package currently supports the "zeroShotReactDescription" agent type.
//
// AgentOption is a type alias for a map of string keys to any value,
// representing the options for the agent.
//
// Options is a function type that can be used to modify the AgentOption,
// such as the WithVerbosity function, which sets the verbosity option for the agent.
package agent
