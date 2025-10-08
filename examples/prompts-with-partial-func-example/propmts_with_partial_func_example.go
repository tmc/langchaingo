package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tmc/langchaingo/prompts"
)

func main() {
	prompt := prompts.PromptTemplate{
		Template:       "Tell me a {{.adjective}} joke about the day {{.date}}",
		InputVariables: []string{"adjective"},
		PartialVariables: map[string]any{
			"date": func() string {
				return time.Now().Format("January 02, 2006")
			},
		},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	result, err := prompt.Format(map[string]any{
		"adjective": "funny",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}
