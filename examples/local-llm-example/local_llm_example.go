package main

import (
	"context"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/local"
)

func main() {
	// You may instantiate a client with a default bin and args from environment variable
	llm, err := local.New()
	if err != nil {
		log.Fatal(err)
	}

	// Or instantiate a client with a custom bin and args options
	//clientOptions := []local.Option{
	//	local.WithBin("/usr/bin/echo"),
	//	local.WithArgs("--arg1=value1 --arg2=value2"),
	//	local.WithGlobalAsArgs(), // build key-value arguments from global llms.Options, then append to args
	//}
	//llm, err := local.New(clientOptions...)

	// Init context
	ctx := context.Background()

	// By default, library will use default bin and args
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, "How many sides does a square have?")
	// Or append to default args options from global llms.Options
	//generateOptions := []llms.CallOption{
	//	llms.WithTopK(10),
	//	llms.WithTopP(0.95),
	//	llms.WithSeed(13),
	//}
	// In that case command will look like: /path/to/bin --arg1=value1 --arg2=value2 --top_k=10 --top_p=0.95 --seed=13 "How many sides does a square have?"
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
