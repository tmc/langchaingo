package agentExecutor

type AgentExecutor interface {
	Run(query string) (string, error)
}
