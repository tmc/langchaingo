// Package agents defines the types for langchaingo Agetns.
package agents

type Agent interface {
	Run(query string) (string, error)
}

func Run(agent Agent, query string) (string, error) {
	return agent.Run(query)
}
