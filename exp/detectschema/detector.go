package detectschema

import "github.com/tmc/langchaingo/llms"

type Detector struct {
	llm llms.Model
}

func New(llm llms.Model) *Detector {
	return &Detector{
		llm: llm,
	}
}
