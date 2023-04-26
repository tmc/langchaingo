package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "The first man to walk on the moon", []string{"Armstrong"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
