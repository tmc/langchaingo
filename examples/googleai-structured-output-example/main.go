package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func main() {
	ctx := context.Background()

	// Get API key from environment
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	// Create Google AI client
	llm, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithDefaultModel("gemini-2.0-flash"),
	)
	if err != nil {
		log.Fatalf("Failed to create Google AI client: %v", err)
	}
	defer llm.Close()

	fmt.Println("Google AI Structured Output Example")
	fmt.Println("====================================\n")

	// Example 1: Simple user information extraction
	fmt.Println("Example 1: User Information Extraction")
	fmt.Println("---------------------------------------")
	runUserInfoExample(ctx, llm)

	fmt.Println()

	// Example 2: Recipe extraction with nested arrays
	fmt.Println("Example 2: Recipe Extraction")
	fmt.Println("-----------------------------")
	runRecipeExample(ctx, llm)

	fmt.Println()

	// Example 3: Nested objects (company and CEO info)
	fmt.Println("Example 3: Nested Objects")
	fmt.Println("--------------------------")
	runNestedObjectExample(ctx, llm)

	fmt.Println()

	// Example 4: Enum validation (sentiment analysis)
	fmt.Println("Example 4: Enum Validation (Sentiment)")
	fmt.Println("---------------------------------------")
	runSentimentExample(ctx, llm)
}

func runUserInfoExample(ctx context.Context, llm *googleai.GoogleAI) {
	// Define schema for user information
	schema := &llms.StructuredOutputDefinition{
		Name:        "user_info",
		Description: "Extract user information from text",
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
				"occupation": {
					Type:        llms.SchemaTypeString,
					Description: "User's occupation or job title",
				},
			},
			Required: []string{"name", "age"},
		},
	}

	// Create prompt
	prompt := "Extract information: Sarah Johnson is a 28-year-old software engineer. Her email is sarah.j@techcorp.com"

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	// Generate with structured output
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(schema),
		llms.WithMaxTokens(200),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// Parse JSON response
	var userInfo struct {
		Name       string `json:"name"`
		Age        int    `json:"age"`
		Email      string `json:"email"`
		Occupation string `json:"occupation"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &userInfo); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Display result
	fmt.Printf("Input: %s\n\n", prompt)
	fmt.Printf("Extracted Information:\n")
	fmt.Printf("  Name: %s\n", userInfo.Name)
	fmt.Printf("  Age: %d\n", userInfo.Age)
	fmt.Printf("  Email: %s\n", userInfo.Email)
	fmt.Printf("  Occupation: %s\n", userInfo.Occupation)
}

func runRecipeExample(ctx context.Context, llm *googleai.GoogleAI) {
	// Define schema for recipe
	schema := &llms.StructuredOutputDefinition{
		Name:        "recipe",
		Description: "Extract or generate recipe information",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"title": {
					Type:        llms.SchemaTypeString,
					Description: "Recipe title",
				},
				"servings": {
					Type:        llms.SchemaTypeInteger,
					Description: "Number of servings",
				},
				"prep_time": {
					Type:        llms.SchemaTypeString,
					Description: "Preparation time",
				},
				"ingredients": {
					Type:        llms.SchemaTypeArray,
					Description: "List of ingredients with amounts",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
				"steps": {
					Type:        llms.SchemaTypeArray,
					Description: "Cooking instructions in order",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
			},
			Required: []string{"title", "ingredients", "steps"},
		},
	}

	// Create prompt
	prompt := "Give me a simple recipe for chocolate chip cookies"

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	// Generate with structured output
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(schema),
		llms.WithMaxTokens(500),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// Parse JSON response
	var recipe struct {
		Title       string   `json:"title"`
		Servings    int      `json:"servings,omitempty"`
		PrepTime    string   `json:"prep_time,omitempty"`
		Ingredients []string `json:"ingredients"`
		Steps       []string `json:"steps"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &recipe); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Display result
	fmt.Printf("Input: %s\n\n", prompt)
	fmt.Printf("Recipe: %s\n", recipe.Title)
	if recipe.Servings > 0 {
		fmt.Printf("Servings: %d\n", recipe.Servings)
	}
	if recipe.PrepTime != "" {
		fmt.Printf("Prep Time: %s\n", recipe.PrepTime)
	}
	fmt.Println("\nIngredients:")
	for i, ingredient := range recipe.Ingredients {
		fmt.Printf("  %d. %s\n", i+1, ingredient)
	}
	fmt.Println("\nSteps:")
	for i, step := range recipe.Steps {
		fmt.Printf("  %d. %s\n", i+1, step)
	}
}

func runNestedObjectExample(ctx context.Context, llm *googleai.GoogleAI) {
	// Define schema with nested objects
	schema := &llms.StructuredOutputDefinition{
		Name:        "company_info",
		Description: "Extract company and leadership information",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"company_name": {
					Type:        llms.SchemaTypeString,
					Description: "Name of the company",
				},
				"founded_year": {
					Type:        llms.SchemaTypeInteger,
					Description: "Year the company was founded",
				},
				"ceo": {
					Type:        llms.SchemaTypeObject,
					Description: "CEO information",
					Properties: map[string]*llms.StructuredOutputSchema{
						"name": {
							Type:        llms.SchemaTypeString,
							Description: "CEO's full name",
						},
						"age": {
							Type:        llms.SchemaTypeInteger,
							Description: "CEO's age",
						},
						"education": {
							Type:        llms.SchemaTypeString,
							Description: "Educational background",
						},
					},
					Required: []string{"name"},
				},
				"headquarters": {
					Type:        llms.SchemaTypeObject,
					Description: "Headquarters location",
					Properties: map[string]*llms.StructuredOutputSchema{
						"city": {Type: llms.SchemaTypeString},
						"state": {Type: llms.SchemaTypeString},
						"country": {Type: llms.SchemaTypeString},
					},
					Required: []string{"city", "country"},
				},
			},
			Required: []string{"company_name", "ceo"},
		},
	}

	// Create prompt
	prompt := "Tell me about Anthropic. The company was founded in 2021 and is led by Dario Amodei. They are based in San Francisco, California, USA."

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	// Generate with structured output
	resp, err := llm.GenerateContent(ctx, messages,
		llms.WithStructuredOutput(schema),
		llms.WithMaxTokens(300),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	// Parse JSON response
	var companyInfo struct {
		CompanyName  string `json:"company_name"`
		FoundedYear  int    `json:"founded_year,omitempty"`
		CEO          struct {
			Name      string `json:"name"`
			Age       int    `json:"age,omitempty"`
			Education string `json:"education,omitempty"`
		} `json:"ceo"`
		Headquarters struct {
			City    string `json:"city"`
			State   string `json:"state,omitempty"`
			Country string `json:"country"`
		} `json:"headquarters,omitempty"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Content), &companyInfo); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Display result
	fmt.Printf("Input: %s\n\n", prompt)
	fmt.Printf("Company: %s\n", companyInfo.CompanyName)
	if companyInfo.FoundedYear > 0 {
		fmt.Printf("Founded: %d\n", companyInfo.FoundedYear)
	}
	fmt.Printf("\nCEO:\n")
	fmt.Printf("  Name: %s\n", companyInfo.CEO.Name)
	if companyInfo.CEO.Age > 0 {
		fmt.Printf("  Age: %d\n", companyInfo.CEO.Age)
	}
	if companyInfo.CEO.Education != "" {
		fmt.Printf("  Education: %s\n", companyInfo.CEO.Education)
	}
	if companyInfo.Headquarters.City != "" {
		fmt.Printf("\nHeadquarters:\n")
		fmt.Printf("  City: %s\n", companyInfo.Headquarters.City)
		if companyInfo.Headquarters.State != "" {
			fmt.Printf("  State: %s\n", companyInfo.Headquarters.State)
		}
		fmt.Printf("  Country: %s\n", companyInfo.Headquarters.Country)
	}
}

func runSentimentExample(ctx context.Context, llm *googleai.GoogleAI) {
	// Define schema with enum validation
	schema := &llms.StructuredOutputDefinition{
		Name:        "sentiment_analysis",
		Description: "Analyze sentiment of text",
		Schema: &llms.StructuredOutputSchema{
			Type: llms.SchemaTypeObject,
			Properties: map[string]*llms.StructuredOutputSchema{
				"text": {
					Type:        llms.SchemaTypeString,
					Description: "The analyzed text",
				},
				"sentiment": {
					Type:        llms.SchemaTypeString,
					Description: "Overall sentiment",
					Enum:        []string{"positive", "negative", "neutral"},
				},
				"confidence": {
					Type:        llms.SchemaTypeNumber,
					Description: "Confidence score between 0 and 1",
				},
				"key_phrases": {
					Type:        llms.SchemaTypeArray,
					Description: "Key phrases that influenced the sentiment",
					Items: &llms.StructuredOutputSchema{
						Type: llms.SchemaTypeString,
					},
				},
			},
			Required: []string{"sentiment", "confidence"},
		},
	}

	// Test multiple sentiments
	texts := []string{
		"This product is absolutely amazing! Best purchase ever!",
		"Terrible experience. Would not recommend to anyone.",
		"It's okay, nothing special but it works as expected.",
	}

	for _, text := range texts {
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman,
				fmt.Sprintf("Analyze the sentiment of this text: %s", text)),
		}

		// Generate with structured output
		resp, err := llm.GenerateContent(ctx, messages,
			llms.WithStructuredOutput(schema),
			llms.WithMaxTokens(200),
		)
		if err != nil {
			log.Printf("Failed to analyze sentiment: %v", err)
			continue
		}

		// Parse JSON response
		var analysis struct {
			Text        string   `json:"text,omitempty"`
			Sentiment   string   `json:"sentiment"`
			Confidence  float64  `json:"confidence"`
			KeyPhrases  []string `json:"key_phrases,omitempty"`
		}

		if err := json.Unmarshal([]byte(resp.Choices[0].Content), &analysis); err != nil {
			log.Printf("Failed to parse response: %v", err)
			continue
		}

		// Display result
		fmt.Printf("Text: \"%s\"\n", text)
		fmt.Printf("  Sentiment: %s (confidence: %.2f)\n", analysis.Sentiment, analysis.Confidence)
		if len(analysis.KeyPhrases) > 0 {
			fmt.Printf("  Key Phrases: %v\n", analysis.KeyPhrases)
		}
		fmt.Println()
	}
}
