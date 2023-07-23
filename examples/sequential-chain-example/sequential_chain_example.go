package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	simpleSequentialChainExample()
	sequentialChainExample()
}

func simpleSequentialChainExample() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	template1 := `
        You are a playwright. Given the title of play, it is your job to write a synopsis for that title.
        Title: {{.title}}
        Playwright: This is a synopsis for the above play:
    `
	chain1 := chains.NewLLMChain(llm, prompts.NewPromptTemplate(template1, []string{"title"}))

	template2 := `
        You are a play critic from the New York Times. Given the synopsis of play, it is your job to write a review for that play.
        Play Synopsis:
        {{.synopsis}}
        Review from a New York Times play critic of the above play:
    `
	chain2 := chains.NewLLMChain(llm, prompts.NewPromptTemplate(template2, []string{"synopsis"}))

	simpleSeqChain, err := chains.NewSimpleSequentialChain([]chains.Chain{chain1, chain2})
	if err != nil {
		log.Fatal(err)
	}

	res, err := chains.Run(context.Background(), simpleSeqChain, "Tragedy at sunset on the beach")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}

func sequentialChainExample() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	template1 := `
	You are a playwright. Given the title of play and the era it is set in, it is your job to write a synopsis for that title.
	Title: {{.title}}
	Era: {{.era}}
	Playwright: This is a synopsis for the above play:
	`
	chain1 := chains.NewLLMChain(llm, prompts.NewPromptTemplate(template1, []string{"title", "era"}))
	chain1.OutputKey = "synopsis"

	template2 := `
		You are a play critic from the New York Times. Given the synopsis of play, it is your job to write a review for that play.
		Play Synopsis:
		{{.synopsis}}
		Review from a New York Times play critic of the above play:
	`
	chain2 := chains.NewLLMChain(llm, prompts.NewPromptTemplate(template2, []string{"synopsis"}))
	chain2.OutputKey = "review"

	sequentialChain, err := chains.NewSequentialChain([]chains.Chain{chain1, chain2}, []string{"title", "era"}, []string{"review"})
	if err != nil {
		log.Fatal(err)
	}

	res, err := chains.Call(context.Background(), sequentialChain, map[string]any{"title": "Tragedy at sunset on the beach", "era": "Victorian"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res["review"])
}
