package agentExecutor

import "github.com/tmc/langchaingo/schema"

type AgentExecutor interface {
	Run(query string) (*schema.AgentFinish, error)
}
