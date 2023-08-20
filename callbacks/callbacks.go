package callbacks

import (
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// Handler is the interface that allows for hooking
type Handler interface {
	HandleText(text string)
	HandleLLMStart(prompts []string)
	HandleLLMEnd(output llms.LLMResult)
	HandleChainStart(inputs map[string]any)
	HandleChainEnd(outputs map[string]any)
	HandleToolStart(input string)
	HandleToolEnd(output string)
	HandleAgentAction(action schema.AgentAction)
	HandleRetrieverStart(query string)
	HandleRetrieverEnd(documents []schema.Document)
}

// HandlerHaver is an interface used to get callbacks handler.
type HandlerHaver interface {
	GetCallbackHandler() Handler
}
