package memory

import (
	"github.com/tmc/langchaingo/schema"
)

// Simple is a class that implement the memory interface, but does nothing.
// The class is used as default in multiple chains.
type Simple struct{}

func NewSimple() Simple {
	return Simple{}
}

// Statically assert that Simple implement the memory interface.
var _ schema.Memory = Simple{}

func (m Simple) MemoryVariables() []string {
	return nil
}

func (m Simple) LoadMemoryVariables(map[string]any) (map[string]any, error) {
	return make(map[string]any, 0), nil
}

func (m Simple) SaveContext(map[string]any, map[string]any) error {
	return nil
}

func (m Simple) Clear() error {
	return nil
}

func (m Simple) GetMemoryKey() string {
	return ""
}
