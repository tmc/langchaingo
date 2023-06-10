package agents

import (
	"context"

	"github.com/tmc/langchaingo/schema"
)

// Agent is the interface all agents must implement.
type Agent interface {
	// Given an input and previous steps decide what to do next. Returns
	// either actions or a finish.
	Plan(ctx context.Context, intermediateSteps []schema.AgentStep, inputs map[string]string) ([]schema.AgentAction, *schema.AgentFinish, error) //nolint:lll
	GetInputKeys() []string
	GetOutputKeys() []string
}
