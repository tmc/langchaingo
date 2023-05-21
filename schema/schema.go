package schema

// AgentAction is the agent's action to take.
type AgentAction struct {
	Tool      string
	ToolInput string
	Log       string
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

// Generation is the output of a single generation.
type Generation struct {
	// Generated text output.
	Text string
	// Raw generation info response from the provider.
	// May include things like reason for finishing (e.g. in OpenAI).
	GenerationInfo map[string]any
}

// LLMResult is the class that contains all relevant information for an LLM Result.
type LLMResult struct {
	Generations [][]Generation
	LLMOutput   map[string]any
}
