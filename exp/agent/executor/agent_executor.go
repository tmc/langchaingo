package executor

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// AgentExecutor is an interface that defines methods for executing agent
// queries. Implementations of this interface are responsible for executing
// a query and returning the result in the form of an AgentFinish object
// or an error if the execution fails.
type AgentExecutor interface {
	Run(ctx context.Context, query string) (*schema.AgentFinish, error)
}
