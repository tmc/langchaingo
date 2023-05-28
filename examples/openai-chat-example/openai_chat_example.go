package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
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
		schema.SystemChatMessage{Text: "Hello, I am a friendly chatbot. I love to talk about movies, books and music. Answer in long form yaml."},
		schema.HumanChatMessage{Text: "What would be a good company name a company that makes colorful socks?"},
	}, llms.WithModel("gpt-4"),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion.Message.Text)
}
