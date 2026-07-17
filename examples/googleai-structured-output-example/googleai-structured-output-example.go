// Set the GOOGLE_API_KEY env var to your API key taken from ai.google.dev
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
)

// MoonLanding is the typed result we ask Gemini to produce. Its shape mirrors the
// JSON Schema below.
type MoonLanding struct {
	Mission    string   `json:"mission"`
	Year       int      `json:"year"`
	Astronauts []string `json:"astronauts"`
}

func main() {
	ctx := context.Background()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	llm, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-2.5-flash"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// A raw JSON Schema (Draft 2020-12) describing exactly the shape we want back.
	// The SDK sends it to Gemini and validates the response against it.
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"mission":    {"type": "string"},
			"year":       {"type": "integer"},
			"astronauts": {"type": "array", "items": {"type": "string"}}
		},
		"required": ["mission", "year", "astronauts"]
	}`)

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Give the details of the first crewed Moon landing."),
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(llms.StructuredOutputConfig{
			Name:        "moon_landing",
			Description: "Details of a crewed Moon landing",
			Schema:      schema,
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// On success the content is a single JSON value that matches the schema and
	// has already been validated by the SDK, so it unmarshals straight into a
	// typed struct — no prompt-engineering or manual parsing needed.
	var landing MoonLanding
	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &landing); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mission:    %s\n", landing.Mission)
	fmt.Printf("Year:       %d\n", landing.Year)
	fmt.Printf("Astronauts: %v\n", landing.Astronauts)
}
