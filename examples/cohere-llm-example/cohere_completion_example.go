package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/cohere"
)

func main() {
	llm, err := cohere.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	input := "The first man to walk on the moon"
	completion, err := llm.Call(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)

	inputToken := llms.CountTokens("", input)
	outputToken := llms.CountTokens("", completion)

	fmt.Printf("%v/%v\n", inputToken, outputToken)
}
