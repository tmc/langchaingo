package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Chat(ctx, []schema.ChatMessage{
		schema.HumanChatMessage{Text: "What would be a good company name a company that makes colorful socks?"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Message.Text)
}
