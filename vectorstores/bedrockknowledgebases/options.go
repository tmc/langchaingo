package bedrockknowledgebases

import (
	"github.com/tmc/langchaingo/vectorstores"
)

func (kb *KnowledgeBase) getOptions(options ...vectorstores.Option) *vectorstores.Options {
	opts := &vectorstores.Options{}
	for _, opt := range options {
		opt(opts)
	}
	return opts
}

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
