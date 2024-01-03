package selfquery

import (
	"fmt"

	queryconstrutor_prompts "github.com/tmc/langchaingo/exp/tools/query_construtor/prompts"
)

func FromLLM() {
	fmt.Printf("queryconstrutor_prompts.DefaultSchema: %v\n", queryconstrutor_prompts.DefaultSchema)
}
