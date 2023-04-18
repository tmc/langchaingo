package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/exp/chains"
	"github.com/tmc/langchaingo/exp/prompts"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// We can construct an LLMChain from a PromptTemplate and an LLM.
	model, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}

	template := "What is a good name for a company that makes {product}?"
	prompt, err := prompts.NewPromptTemplate(template, []string{"product"})
	if err != nil {
		log.Fatal(err)
	}

	chainA := chains.NewLLMChain(model, prompt)

	// We call chains using the chain.Call function, not the Call method
	resA, err := chains.Call(chainA, map[string]any{"product": "fake food"})
	if err != nil {
		log.Fatal(err)
	}

	// The result is an map with a `text` field containing the result.
	fmt.Println(resA["text"]) //Faux Cuisine.

	// We can also construct an LLMChain from a ChatPromptTemplate and a chat model.

	//Chat not implemented
	/* chat, err := openai.NewChat()
	if err != nil {
		log.Fatal(err)
	}

	systemTemplate, err := prompts.NewPromptTemplate(
		"You are a helpful assistant that translates {input_language} to {output_language}.",
		[]string{"input_language", "output_language"},
	)
	if err != nil {
		log.Fatal(err)
	}

	userTemplate, err := prompts.NewPromptTemplate("{text}", []string{"text"})
	if err != nil {
		log.Fatal(err)
	}

	chatPrompt, err := prompts.NewChatTemplate([]prompts.Message{
		prompts.NewSystemMessage(systemTemplate),
		prompts.NewHumanMessage(userTemplate),
	}, []string{"input_language", "output_language", "text"})

	chainB := chains.NewLLMChain(chat, chatPrompt)

	resB, err := chains.Call(chainB, map[string]any{
		"input_language":  "English",
		"output_language": "French",
		"text":            "I love programming",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resB) */
}
