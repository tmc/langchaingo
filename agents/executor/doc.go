// Package executor provides a standard chain for executing agent queries.
//
// The Executor is an Agent and set of Tools. The agent executor is
// responsible for calling the agent, getting back and action and action input,
// calling the tool that the action references with the corresponding input,
// getting the output of the tool, and then passing all that information back
// into the Agent to get the next action it should take.
//
// The package also contains functions to initialize executors with agents and
// supports different agent types and options to customize the agent's behavior.
//
// AgentType is a string type representing the type of agent to create.
// The package currently supports the "zeroShotReactDescription" agent type.
//
// Options is a type alias for a map of string keys to any value,
// representing the options for the agent and the executor.
//
// Option is a function type that can be used to modify the Options,
// such as the WithVerbosity function, which sets the verbosity option
// for the agent.
package executor
