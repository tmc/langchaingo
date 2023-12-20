// Package agents contains the standard interface all agents must implement,
// implementations of this interface, and an agent executor.
//
// An Agent is a wrapper around a model, which takes in user input and returns
// a response corresponding to an “action” to take and a corresponding
// “action input”. Alternatively the agent can return a finish with the
// finished answer to the query. This package contains and standard interface
// for such agents.
//
// Package agents provides and implementation of the agent interface called
// OneShotZeroAgent. This agent uses the ReAct Framework (based on the
// descriptions of tools) to decide what action to take. This agent is
// optimized to be used with LLMs.
//
// To make agents more powerful we need to make them iterative, i.e. call the
// model multiple times until they arrive at the final answer. That's the job of
// the Executor. The Executor is an Agent and set of Tools. The agent executor is
// responsible for calling the agent, getting back and action and action input,
// calling the tool that the action references with the corresponding input,
// getting the output of the tool, and then passing all that information back
// into the Agent to get the next action it should take.
package agents
