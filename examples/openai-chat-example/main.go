package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Chat(ctx, []openai.ChatMessage{{Role: "user", Content: "What would be a good company name a company that makes colorful socks?"}})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Content)
}
