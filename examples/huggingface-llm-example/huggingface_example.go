package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/huggingface"
)

func main() {
	llm, err := huggingface.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "What would be a good company name be for name a company that makes colorful socks?", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
