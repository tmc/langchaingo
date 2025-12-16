// Set the GOOGLE_API_KEY env var to your API key taken from ai.google.dev
//
// This example demonstrates the unified google package that automatically
// selects the appropriate underlying SDK based on the model:
// - gemini-3+ models use googleaiv2 (google.golang.org/genai SDK)
// - older models use googleai (github.com/google/generative-ai-go SDK)
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vendasta/langchaingo/llms"
	"github.com/vendasta/langchaingo/llms/google"
)

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")

	// Using gemini-2.0-flash - automatically uses googleai (original SDK)
	// When gemini-3+ models become available, they will use googleaiv2 automatically
	llm, err := google.New(ctx,
		google.WithAPIKey(apiKey),
		google.WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer llm.Close()

	prompt := "Who was the second person to walk on the moon?"
	answer, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}
