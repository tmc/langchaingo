package executor

import "errors"

var (
	// ErrExecutorInputNotString is returned if an input to the executor call function is not a string.
	ErrExecutorInputNotString = errors.New("input to executor not string")
	// ErrAgentNoReturn is returned if the agent returns no actions and no finish.
	ErrAgentNoReturn = errors.New("no actions or finish was returned by the agent")
	// ErrNotFinished is returned if the agent does not give a finish before  the number of iterations
	// is larger then max iterations.
	ErrNotFinished = errors.New("agent not finished before max iterations")
	// ErrUnknownAgentType is returned if the type given to the initializer is invalid.
	ErrUnknownAgentType = errors.New("unknown agent type")
	// ErrInvalidOptions is returned if the options given to the initializer is invalid.
	ErrInvalidOptions = errors.New("invalid options")
)
