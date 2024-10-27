package main

import (
	"context"
	"fmt"
	"github.com/getzep/zep-go"
	zepClient "github.com/getzep/zep-go/client"
	zepOption "github.com/getzep/zep-go/option"
	"github.com/starmvp/langchaingo/chains"
	"github.com/starmvp/langchaingo/llms/openai"
	zepLangchainMemory "github.com/starmvp/langchaingo/memory/zep"
	"os"
)

func main() {
	ctx := context.Background()

	client := zepClient.NewClient(zepOption.WithAPIKey(os.Getenv("ZEP_API_KEY")))
	sessionID := os.Getenv("ZEP_SESSION_ID")

	llm, err := openai.New()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	c := chains.NewConversation(
		llm,
		zepLangchainMemory.NewMemory(
			client,
			sessionID,
			zepLangchainMemory.WithMemoryType(zep.MemoryGetRequestMemoryTypePerpetual),
			zepLangchainMemory.WithReturnMessages(true),
			zepLangchainMemory.WithAIPrefix("Robot"),
			zepLangchainMemory.WithHumanPrefix("Joe"),
		),
	)
	res, err := chains.Run(ctx, c, "Hi! I'm John Doe")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Response: %s\n", res)

	res, err = chains.Run(ctx, c, "What is my name?")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Printf("Response: %s\n", res)
}
