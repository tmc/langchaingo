package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/yzma"
)

const modelPath = "/home/ron/models/SmolLM2-135M-Instruct.Q2_K.gguf"

func main() {
	llm, err := yzma.New(yzma.WithModel(modelPath))
	if err != nil {
		log.Fatal(err)
	}

	// Init context
	ctx := context.Background()

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, "How many sides does a square have?")
	// Or append to default args options from global llms.Options
	//generateOptions := []llms.CallOption{
	//	llms.WithTopK(10),
	//	llms.WithTopP(0.95),
	//	llms.WithTemperature(0.25),
	//}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
