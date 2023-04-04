package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	// We can construct an LLMChain from a PromptTemplate and an LLM.
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	template := "What is a good name for a company that makes {product}?"
	prompt, err := prompts.NewTemplate(template, []string{"product"})
	if err != nil {
		log.Fatal(err)
	}

	chainA := chains.NewLLMChain(llm, prompt)

	// We call chains using the chain.Call function, not the Call method
	resA, err := chains.Call(chainA, map[string]any{"product": "fake food"})

	// The result is an map with a `text` field.
	fmt.Println(resA["text"]) //Faux Cuisine.

}
