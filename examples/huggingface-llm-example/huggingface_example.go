package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/huggingface"
)

func main() {
	llm, err := huggingface.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// By default, library will use default model described in huggingface.defaultModel
	// completion, err := llm.Call(ctx, "What would be a good company name be for name a company that makes colorful socks?")

	// Or override default model to another one
	generateOptions := []llms.CallOption{
		llms.WithModel("gpt2"),
		//llms.WithTopK(10),
		//llms.WithTopP(0.95),
	}
	completion, err := llm.Call(ctx, "What would be a good company name be for name a company that makes colorful socks?", generateOptions...)

	// Check for errors
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
