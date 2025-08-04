package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/huggingface"
)

func main() {
	// You may instantiate a client with a custom token and/or custom model
	// clientOptions := []huggingface.Option{
	// 	huggingface.WithToken("HF_1234"),
	// 	huggingface.WithModel("ZZZ"),
	// }
	// llm, err := huggingface.New(clientOptions...)
	
	// Or you may instantiate a client with a default model and use token from environment variable
	llm, err := huggingface.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	
	// Or override default model to another one
	generateOptions := []llms.CallOption{
		llms.WithModel("gpt2"),
		// llms.WithTopK(10),
		// llms.WithTopP(0.95),
		// llms.WithSeed(13),
	}
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"What would be a good company name be for name a company that makes colorful socks?",
		generateOptions...)
	// Check for errors
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
