//nolint:forbidigo
package callbacks

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type SimpleHandler struct{}

var _ Handler = SimpleHandler{}

func (SimpleHandler) HandleText(context.Context, string)                                   {}
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
