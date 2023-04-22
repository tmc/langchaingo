// Package executor provides interfaces and implementations for executing
// agent queries.
//
// An agent query is a string that describes a task to be performed by an agent,
// which is a software program that acts on behalf of a user or another program.
// The query can specify various parameters, such as input data, processing
// options, and output format.
//
// To execute an agent query, an application needs to create an object that
// implements the AgentExecutor interface, which defines a method for running
// a query and returning the result. The AgentFinish type from the langchaingo
// schema package is used to represent the result of a successful execution.
// If the execution fails, an error is returned.
//
// The executor package provides a sample implementation of the AgentExecutor
// interface, called BasicExecutor, which executes queries using a simple
// command-line interface. This implementation is intended for testing and
// demonstration purposes only and should not be used in production environments.
package executor
