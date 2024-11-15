package schema

// AgentAction is the agent's action to take.
type AgentAction struct {
	Tool      string
	ToolInput string
	Log       string
	ToolID    string
}

// AgentStep is a step of the agent.
type AgentStep struct {
	Action      AgentAction
	Observation string
}

// AgentFinish is the agent's return value.
type AgentFinish struct {
	ReturnValues map[string]any
	Log          string
}
