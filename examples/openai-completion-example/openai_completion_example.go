package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx,
		llm,
		"The first man to walk on the moon",
		llms.WithTemperature(0.8),
		llms.WithStopWords([]string{"Armstrong"}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The first man to walk on the moon:")
	fmt.Println(completion)
}
