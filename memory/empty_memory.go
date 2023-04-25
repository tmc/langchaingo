package memory

import (
	"github.com/tmc/langchaingo/schema"
)

// Empty memory is a class that implement the memory interface, but does nothing.
// The class is used as default in multiple chains.
type Empty struct{}

func NewEmpty() Empty {
	return Empty{}
}

// Statically assert that EmptyMemory implement the memory interface.
var _ schema.Memory = Empty{}

func (m Empty) MemoryVariables() []string {
	return nil
}

func (m Empty) LoadMemoryVariables(map[string]any) (map[string]any, error) {
	return make(map[string]any, 0), nil
}

func (m Empty) SaveContext(map[string]any, map[string]any) error {
	return nil
}

func (m Empty) Clear() error {
	return nil
}
