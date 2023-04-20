package schema

// Memory is the interface for memory in chains.
type Memory interface {
	// Input keys this memory class will load dynamically.
	MemoryVariables() []string
	// Return key-value pairs given the text input to the chain.
	// If None, return all memories
	LoadMemoryVariables(inputs map[string]any) map[string]any
	// Save the context of this model run to memory.
	SaveContext(inputs map[string]any, outputs map[string]string) error
	// Clear memory contents.
	Clear() error
}
