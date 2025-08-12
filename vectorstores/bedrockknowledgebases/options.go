package bedrockknowledgebases

import (
	"github.com/vendasta/langchaingo/vectorstores"
)

func (kb *KnowledgeBase) getOptions(options ...vectorstores.Option) *vectorstores.Options {
	opts := &vectorstores.Options{}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}

// Filter is the common interface for representing filters applied during
// document retrieval from a Bedrock knowledge base.
//
// Filters can be combined using logical operators: AllFilter (AND semantics)
// and AnyFilter (OR semantics). This allows the construction of complex,
// nested filter conditions.
//
// The interface is sealed by the unexported isFilter method to prevent
// external implementations, ensuring forward compatibility.
type Filter interface {
	isFilter()
}

type EqualsFilter struct {
	Key   string
	Value string
}

func (f EqualsFilter) isFilter() {}

type NotEqualsFilter struct {
	Key   string
	Value string
}

func (f NotEqualsFilter) isFilter() {}

type ContainsFilter struct {
	Key   string
	Value string
}

func (f ContainsFilter) isFilter() {}

type AllFilter struct {
	Filters []Filter
}

func (f AllFilter) isFilter() {}

type AnyFilter struct {
	Filters []Filter
}

func (f AnyFilter) isFilter() {}
