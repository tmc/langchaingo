package schema

// Memory is the interface for memory in chains.
type Memory interface {
	// GetMemoryKey getter for memory key.
	GetMemoryKey() string
	// MemoryVariables Input keys this memory class will load dynamically.
	MemoryVariables() []string
	// LoadMemoryVariables Return key-value pairs given the text input to the chain.
	// If None, return all memories
	LoadMemoryVariables(inputs map[string]any) (map[string]any, error)
	// SaveContext Save the context of this model run to memory.
	SaveContext(inputs map[string]any, outputs map[string]any) error
	// Clear memory contents.
	Clear() error
}
