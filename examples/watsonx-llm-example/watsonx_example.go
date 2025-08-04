package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/yincongcyincong/langchaingo/llms"
	"github.com/yincongcyincong/langchaingo/llms/watsonx"
)

func main() {
	llm, err := watsonx.New(
		"meta-llama/llama-3-70b-instruct",
		//// Optional parameters:
		// wx.WithWatsonxAPIKey("YOUR WATSONX API KEY"),
		// wx.WithWatsonxProjectID("YOUR WATSONX PROJECT ID"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	
	// Or override default model to another one
	prompt := "What would be a good company name be for name a company that makes colorful socks?"
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		prompt,
		llms.WithTopK(10),
		llms.WithTopP(0.95),
		llms.WithSeed(13),
	)
	// Check for errors
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
