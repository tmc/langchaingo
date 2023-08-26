package callbacks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Handler is the interface that allows for hooking into specific parts of an
// LLM application.
type Handler interface {
	HandleText(ctx context.Context, text string)
	HandleLLMStart(ctx context.Context, prompts []string)
	HandleLLMEnd(ctx context.Context, output llms.LLMResult)
	HandleChainStart(ctx context.Context, inputs map[string]any)
	HandleChainEnd(ctx context.Context, outputs map[string]any)
	HandleToolStart(ctx context.Context, input string)
	HandleToolEnd(ctx context.Context, output string)
	HandleAgentAction(ctx context.Context, action schema.AgentAction)
	HandleRetrieverStart(ctx context.Context, query string)
	HandleRetrieverEnd(ctx context.Context, documents []schema.Document)
}

// HandlerHaver is an interface used to get callbacks handler.
type HandlerHaver interface {
	GetCallbackHandler() Handler
}
