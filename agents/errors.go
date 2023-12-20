package agents

import "errors"

var (
	// ErrExecutorInputNotString is returned if an input to the executor call function is not a string.
	ErrExecutorInputNotString = errors.New("input to executor not string")
	// ErrAgentNoReturn is returned if the agent returns no actions and no finish.
	ErrAgentNoReturn = errors.New("no actions or finish was returned by the agent")
	// ErrNotFinished is returned if the agent does not give a finish before  the number of iterations
	// is larger than max iterations.
	ErrNotFinished = errors.New("agent not finished before max iterations")
	// ErrUnknownAgentType is returned if the type given to the initializer is invalid.
	ErrUnknownAgentType = errors.New("unknown agent type")
	// ErrInvalidOptions is returned if the options given to the initializer is invalid.
	ErrInvalidOptions = errors.New("invalid options")

	// ErrUnableToParseOutput is returned if the output of the llm is unparsable.
	ErrUnableToParseOutput = errors.New("unable to parse agent output")
	// ErrInvalidChainReturnType is returned if the internal chain of the agent returns a value in the
	// "text" filed that is not a string.
	ErrInvalidChainReturnType = errors.New("agent chain did not return a string")
)

// ParserErrorHandler is the struct used to handle parse errors from the agent in the executor. If
// an executor have a ParserErrorHandler, parsing errors will be formatted using the formatter
// function and added as an observation. In the next executor step the agent will then have the
// possibility to fix the error.
type ParserErrorHandler struct {
	// The formatter function can be used to format the parsing error. If nil the error will be given
	// as an observation directly.
	Formatter func(err string) string
}

// NewParserErrorHandler creates a new parser error handler.
func NewParserErrorHandler(formatFunc func(string) string) *ParserErrorHandler {
	return &ParserErrorHandler{
		Formatter: formatFunc,
	}
}
