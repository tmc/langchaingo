// Set the GOOGLE_API_KEY env var to your API key taken from ai.google.dev
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/googleai"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	// See https://ai.google.dev/gemini-api/docs/models/gemini for possible models
	llm, err := googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-1.5-pro"))
	if err != nil {
		log.Fatal(err)
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name for a comapny that produces Go-backed LLM tools?"),
	}
	completion, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		fmt.Print(string(chunk))
		return nil
	}))
	if err != nil {
		log.Fatal(err)
	}
	_ = completion
}
