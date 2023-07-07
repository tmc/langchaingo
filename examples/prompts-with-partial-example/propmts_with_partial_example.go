package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/prompts"
)

func main() {
	prompt := prompts.PromptTemplate{
		Template:       "{{.foo}}{{.bar}}",
		InputVariables: []string{"bar"},
		PartialVariables: map[string]any{
			"foo": "foo",
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	result, err := prompt.Format(map[string]any{
		"bar": "baz",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
