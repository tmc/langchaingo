package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// UserInfo represents structured user information
type UserInfo struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func main() {
	// Define schema using the unified structured output API
	// This works consistently across OpenAI, Google, and Anthropic providers
	schema := &llms.StructuredOutputDefinition{
		Name:        "user_info",
		Description: "Extract user information from text",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"name": {
					Type:        llms.SchemaTypeString,
					Description: "The user's full name",
				},
				"age": {
					Type:        llms.SchemaTypeInteger,
					Description: "The user's age in years",
				},
				"email": {
					Type:        llms.SchemaTypeString,
					Description: "The user's email address",
				},
				"role": {
					Type:        llms.SchemaTypeString,
					Description: "The user's role or job title",
				},
			},
			Required:             []string{"name", "age", "email", "role"},
			AdditionalProperties: false,
		},
		Strict: true, // OpenAI strict mode: enforce exact schema match
	}

	// Create OpenAI LLM client
	llm, err := openai.New(
		openai.WithModel("gpt-4o-2024-08-06"),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Prepare messages
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem,
			"You are an expert at structured data extraction. Extract user information from the given text."),
		llms.TextParts(llms.ChatMessageTypeHuman,
			"John Smith is a 35-year-old software engineer. His email is john.smith@example.com and he works as a Senior Developer."),
	}

	// Generate content with structured output using the unified API
	// This call option works the same way across all providers
	completion, err := llm.GenerateContent(ctx, content,
		llms.WithStructuredOutput(schema),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the structured response
	responseText := completion.Choices[0].Content
	fmt.Println("Raw JSON response:")
	fmt.Println(responseText)
	fmt.Println()

	// Unmarshal into struct
	var user UserInfo
	if err := json.Unmarshal([]byte(responseText), &user); err != nil {
		log.Fatal(err)
	}

	// Display extracted information
	fmt.Println("Extracted User Information:")
	fmt.Printf("  Name:  %s\n", user.Name)
	fmt.Printf("  Age:   %d\n", user.Age)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Role:  %s\n", user.Role)
}
