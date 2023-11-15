package main

import (
	"context"
	"fmt"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie"
	"log"

	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := ernie.NewChat(
		ernie.WithModelName(ernie.ModelNameERNIEBot),
		// Fill in your AK and SK here.
		ernie.WithAKSK("ak", "sk"),
		// Use an external cache for the access token.
		ernie.WithAccessToken("accesstoken"),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, []schema.ChatMessage{
		schema.SystemChatMessage{Content: "Hello, I am a friendly chatbot. I love to talk about movies, books and music. Answer in long form yaml."},
		schema.HumanChatMessage{Content: "What would be a good company name a company that makes colorful socks?"},
	}, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		log.Println(string(chunk))
		return nil
	}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
