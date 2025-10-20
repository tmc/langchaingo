package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

// UserInfo represents the structured output we want from the model
type UserInfo struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

// Recipe represents a cooking recipe
type Recipe struct {
	Title       string   `json:"title"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
	PrepTime    int      `json:"prep_time_minutes"`
}

func main() {
	// Example 1: Extract structured user information
	fmt.Println("=== Example 1: Extract User Information ===")
	extractUserInfo()

	fmt.Println("\n=== Example 2: Generate Recipe ===")
	generateRecipe()
}

func extractUserInfo() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-5-sonnet-20240620"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Define the schema for user information
	userSchema := &llms.StructuredOutputDefinition{
		Name:        "user_info",
		Description: "Extract user information from the text",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"name": {
					Type:        llms.SchemaTypeString,
					Description: "User's full name",
				},
				"age": {
					Type:        llms.SchemaTypeInteger,
					Description: "User's age in years",
				},
				"email": {
					Type:        llms.SchemaTypeString,
					Description: "User's email address",
				},
			},
			Required:             []string{"name", "age"},
			AdditionalProperties: false,
		},
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman,
			"My name is Alice Johnson, I'm 28 years old, and you can reach me at alice.j@example.com"),
	}

	// Request structured output
	// NOTE: For Anthropic, this is simulated via tool calling, but the API is identical to OpenAI/Google
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(userSchema),
		llms.WithMaxTokens(1024),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the JSON response
	var user UserInfo
	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &user); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Extracted User Info:\n")
	fmt.Printf("  Name: %s\n", user.Name)
	fmt.Printf("  Age: %d\n", user.Age)
	fmt.Printf("  Email: %s\n", user.Email)

	// Show token usage
	if genInfo := resp.Choices[0].GenerationInfo; genInfo != nil {
		fmt.Printf("\nToken Usage:\n")
		if inputTokens, ok := genInfo["InputTokens"].(int); ok {
			fmt.Printf("  Input: %d\n", inputTokens)
		}
		if outputTokens, ok := genInfo["OutputTokens"].(int); ok {
			fmt.Printf("  Output: %d\n", outputTokens)
		}
	}
}

func generateRecipe() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-5-sonnet-20240620"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Define the schema for a recipe
	recipeSchema := &llms.StructuredOutputDefinition{
		Name:        "recipe",
		Description: "A cooking recipe",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"title": {
					Type:        llms.SchemaTypeString,
					Description: "The recipe title",
				},
				"ingredients": {
					Type:        llms.SchemaTypeArray,
					Description: "List of ingredients",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
				"steps": {
					Type:        llms.SchemaTypeArray,
					Description: "Cooking steps in order",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
				"prep_time_minutes": {
					Type:        llms.SchemaTypeInteger,
					Description: "Preparation time in minutes",
				},
			},
			Required:             []string{"title", "ingredients", "steps"},
			AdditionalProperties: false,
		},
	}

	ctx := context.Background()
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman,
			"Give me a simple recipe for chocolate chip cookies"),
	}

	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(recipeSchema),
		llms.WithMaxTokens(2048),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the JSON response
	var recipe Recipe
	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &recipe); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Recipe: %s\n", recipe.Title)
	fmt.Printf("Prep Time: %d minutes\n\n", recipe.PrepTime)

	fmt.Println("Ingredients:")
	for i, ingredient := range recipe.Ingredients {
		fmt.Printf("  %d. %s\n", i+1, ingredient)
	}

	fmt.Println("\nSteps:")
	for i, step := range recipe.Steps {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
}
