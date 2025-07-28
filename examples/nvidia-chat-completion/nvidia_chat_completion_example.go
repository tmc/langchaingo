package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/llms/openai"
)

func main() {
	key := os.Getenv("NVIDIA_API_KEY")
	llm, err := openai.New(
		openai.WithBaseURL("https://integrate.api.nvidia.com/v1/"),
		openai.WithModel("mistralai/mixtral-8x7b-instruct-v0.1"),
		openai.WithToken(key),
		// openai.WithHTTPClient(httputil.DebugHTTPClient),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a golang expert"),
		llms.TextParts(llms.ChatMessageTypeHuman, "explain why go is a great fit for ai based products"),
	}

	if _, err = llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(4096),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		})); err != nil {
		log.Fatal(err)
	}
}
