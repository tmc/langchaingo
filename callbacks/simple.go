//nolint:forbidigo
package callbacks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type SimpleHandler struct{}

var _ Handler = SimpleHandler{}

func (SimpleHandler) HandleText(context.Context, string) {}
func (SimpleHandler) HandleLLM(ctx context.Context, _ []llms.MessageContent, _ llms.CallOptions, next func(context.Context) (*llms.ContentResponse, error)) (*llms.ContentResponse, error) {
	return next(ctx)
}

func (SimpleHandler) HandleChain(ctx context.Context, _ map[string]any, _ schema.ChainInfo, next func(context.Context) (map[string]any, error)) (map[string]any, error) {
	return next(ctx)
}
func (SimpleHandler) HandleLLMStart(context.Context, []string)                             {}
func (SimpleHandler) HandleLLMGenerateContentStart(context.Context, []llms.MessageContent) {}
func (SimpleHandler) HandleLLMGenerateContentEnd(context.Context, *llms.ContentResponse)   {}
func (SimpleHandler) HandleLLMError(context.Context, error)                                {}
func (SimpleHandler) HandleChainStart(context.Context, map[string]any)                     {}
func (SimpleHandler) HandleChainEnd(context.Context, map[string]any)                       {}
func (SimpleHandler) HandleChainError(context.Context, error)                              {}
func (SimpleHandler) HandleToolStart(context.Context, string)                              {}
func (SimpleHandler) HandleToolEnd(context.Context, string)                                {}
func (SimpleHandler) HandleToolError(context.Context, error)                               {}
func (SimpleHandler) HandleAgentAction(context.Context, schema.AgentAction)                {}
func (SimpleHandler) HandleAgentFinish(context.Context, schema.AgentFinish)                {}
func (SimpleHandler) HandleRetrieverStart(context.Context, string)                         {}
func (SimpleHandler) HandleRetrieverEnd(context.Context, string, []schema.Document)        {}
func (SimpleHandler) HandleStreamingFunc(context.Context, []byte)                          {}
