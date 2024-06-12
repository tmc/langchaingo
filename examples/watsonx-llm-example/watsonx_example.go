package main

import (
	"context"
	"fmt"
	"log"

	wx "github.com/h0rv/go-watsonx/models"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/watsonx"
)

func main() {
	llm, err := watsonx.New(
		// Optional parameters:
		// wx.WithIBMCloudAPIKey("YOUR IBM CLOUD API KEY"),
		// wx.WithWatsonxProjectID("YOUR WATSONX PROJECT ID"),
		wx.WithModel(wx.LLAMA_2_70B_CHAT),
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
