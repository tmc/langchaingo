package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/local"
)

func main() {
	llm, err := local.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "How many sides does a square have?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
