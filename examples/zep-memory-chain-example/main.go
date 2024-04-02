package main

import (
	"context"
	"fmt"
	"github.com/getzep/zep-go"
	zepClient "github.com/getzep/zep-go/client"
	"github.com/getzep/zep-go/option"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/openai"
	zepLangchainMemory "github.com/tmc/langchaingo/memory/zep"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	client := zepClient.NewClient(
		option.WithAPIKey(fmt.Sprintf("Api-Key %s", os.Getenv("ZEP_API_KEY"))),
	)
	sessionID := os.Getenv("ZEP_SESSION_ID")

	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Error: %s", err)
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
		log.Fatal(err)
	}
	log.Print(res)

	res, err = chains.Run(ctx, c, "What is my name?")
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res)
}
