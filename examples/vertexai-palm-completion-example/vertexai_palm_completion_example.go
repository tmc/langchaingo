package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/vertexai"
)

func main() {
	llm, err := vertexai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "The first man to walk on the moon")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
