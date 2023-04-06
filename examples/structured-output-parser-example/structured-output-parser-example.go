package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/outputParsers"
	"github.com/tmc/langchaingo/prompts"
)

//The structured output parser can be used when you want to return multiple fields.

func main() {
	// With a `StructuredOutputParser` we can define a schema for the output.
	parser := outputParsers.NewStructuredFromNameAndDescription(map[string]string{
		"answer": "answer to the user's question",
		"source": "source used to answer the user's question, should be a website.",
	})

	formatInstructions := parser.GetFormatInstructions()
	template := "Answer the users question as best as possible.\n{format_instructions}\n{question}"

	prompt, err := prompts.NewPromptTemplate(
		template,
		[]string{"question"},
		prompts.WithPartialVariablesPrompt(map[string]any{"format_instructions": formatInstructions}),
	)

	if err != nil {
		log.Fatal(err)
	}

	model, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	input, err := prompt.Format(map[string]any{
		"question": "What is the capital of France?",
	})

	response, err := model.Call(input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(input)
	/*
		Answer the users question as best as possible.
		The output should be a markdown code snippet formatted in the following schema:
		```json
		{
			"answer": string // answer to the user's question
			"source": string // source used to answer the user's question, should be a website.
		}
		```
		What is the capital of France?
	*/

	fmt.Println(response)
	/*
		```json
		{
			"answer": "Paris",
			"source": "https://en.wikipedia.org/wiki/France"
		}
		```
	*/

	parsedResponse, err := parser.Parse(response)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(parsedResponse)
	//map[answer:Paris source:https://en.wikipedia.org/wiki/France]

}
