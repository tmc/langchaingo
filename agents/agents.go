package agents

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

// Agent is the interface all agents must implement.
type Agent interface {
	// Plan Given an input and previous steps decide what to do next. Returns
	// either actions or a finish.
	Plan(ctx context.Context, intermediateSteps []schema.AgentStep, inputs map[string]string, intermediateMessages []llms.ChatMessage) ([]schema.AgentAction, *schema.AgentFinish, []llms.ChatMessage, error) //nolint:lll
	GetInputKeys() []string
	GetOutputKeys() []string
	GetTools() []tools.Tool
}
