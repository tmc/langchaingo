package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

func main() {
	fmt.Println("=== Google AI SDK Migration Example ===")
	
	// Check for required environment variable
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	ctx := context.Background()

	// Example 1: Default behavior (now uses official SDK)
	fmt.Println("\n1. Default Client (Official SDK)")
	defaultClientExample(ctx, apiKey)

	// Example 2: Explicitly using official SDK
	fmt.Println("\n2. Explicit Official SDK")
	officialSDKExample(ctx, apiKey)

	// Example 3: Using legacy SDK for backward compatibility
	fmt.Println("\n3. Legacy SDK for Backward Compatibility")
	legacySDKExample(ctx, apiKey)

	// Example 4: SDK comparison using migration helper
	fmt.Println("\n4. SDK Comparison Using Migration Helper")
	migrationHelperExample(ctx, apiKey)

	// Example 5: Conditional SDK selection
	fmt.Println("\n5. Conditional SDK Selection")
	conditionalSDKExample(ctx, apiKey)
}

func defaultClientExample(ctx context.Context, apiKey string) {
	// Default behavior - now uses official SDK
	client, err := googleai.New(ctx, googleai.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error creating default client: %v", err)
		return
	}
	defer client.Close()

	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the capital of France?"),
			},
		},
	})
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Default Client Response: %s\n", response.Choices[0].Content)
	}
}

func officialSDKExample(ctx context.Context, apiKey string) {
	// Explicitly use the official SDK
	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithSDKVersion(googleai.SDKOfficial),
	)
	if err != nil {
		log.Printf("Error creating official SDK client: %v", err)
		return
	}
	defer client.Close()

	// Alternative constructor
	// client, err := googleai.NewWithOfficialSDK(ctx, googleai.WithAPIKey(apiKey))

	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is 2+2?"),
			},
		},
	})
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Official SDK Response: %s\n", response.Choices[0].Content)
	}
}

func legacySDKExample(ctx context.Context, apiKey string) {
	// Use legacy SDK for backward compatibility
	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithSDKVersion(googleai.SDKLegacy),
	)
	if err != nil {
		log.Printf("Error creating legacy SDK client: %v", err)
		return
	}
	defer client.Close()

	// Alternative constructor
	// client, err := googleai.NewWithLegacySDK(ctx, googleai.WithAPIKey(apiKey))

	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("What is the largest planet in our solar system?"),
			},
		},
	})
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Legacy SDK Response: %s\n", response.Choices[0].Content)
	}
}

func migrationHelperExample(ctx context.Context, apiKey string) {
	// Use migration helper to compare SDK behavior
	helper, err := googleai.NewMigrationHelper(ctx, apiKey)
	if err != nil {
		log.Printf("Error creating migration helper: %v", err)
		return
	}
	defer helper.Close()

	// Note: This is a simplified example. In practice, you would implement
	// actual comparison logic here.
	fmt.Println("Migration helper created successfully for comparing SDKs")
	fmt.Println("Legacy client available:", helper.GetLegacyClient() != nil)
	fmt.Println("Official client available:", helper.GetOfficialClient() != nil)

	// You could implement comparison logic here:
	// - Send same request to both clients
	// - Compare response times
	// - Validate response equivalence
	// - Test error handling differences
}

func conditionalSDKExample(ctx context.Context, apiKey string) {
	// Choose SDK based on environment variable
	var sdkVersion googleai.SDKVersion
	
	if os.Getenv("USE_LEGACY_GOOGLE_SDK") == "true" {
		sdkVersion = googleai.SDKLegacy
		fmt.Println("Using Legacy SDK (via environment variable)")
	} else {
		sdkVersion = googleai.SDKOfficial
		fmt.Println("Using Official SDK (default)")
	}

	client, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithSDKVersion(sdkVersion),
	)
	if err != nil {
		log.Printf("Error creating conditional client: %v", err)
		return
	}
	defer client.Close()

	response, err := client.GenerateContent(ctx, []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart("Tell me a fun fact about AI."),
			},
		},
	})
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("Conditional SDK Response: %s\n", response.Choices[0].Content)
	}
}

// Example of how to implement proper SDK comparison testing
func compareSDKResponses(ctx context.Context, apiKey string, prompt string) error {
	// Create clients for both SDKs
	legacyClient, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithSDKVersion(googleai.SDKLegacy),
	)
	if err != nil {
		return fmt.Errorf("failed to create legacy client: %w", err)
	}
	defer legacyClient.Close()

	officialClient, err := googleai.New(ctx,
		googleai.WithAPIKey(apiKey),
		googleai.WithSDKVersion(googleai.SDKOfficial),
	)
	if err != nil {
		return fmt.Errorf("failed to create official client: %w", err)
	}
	defer officialClient.Close()

	// Prepare the same request
	content := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	}

	// Get responses from both SDKs
	legacyResponse, err := legacyClient.GenerateContent(ctx, content)
	if err != nil {
		return fmt.Errorf("legacy SDK error: %w", err)
	}

	officialResponse, err := officialClient.GenerateContent(ctx, content)
	if err != nil {
		return fmt.Errorf("official SDK error: %w", err)
	}

	// Compare responses (simplified comparison)
	fmt.Printf("Legacy SDK response length: %d\n", len(legacyResponse.Choices))
	fmt.Printf("Official SDK response length: %d\n", len(officialResponse.Choices))

	if len(legacyResponse.Choices) > 0 && len(officialResponse.Choices) > 0 {
		legacyText := legacyResponse.Choices[0].Content
		officialText := officialResponse.Choices[0].Content
		
		fmt.Printf("Response similarity: %s\n", 
			func() string {
				if len(legacyText) == len(officialText) {
					return "identical length"
				}
				return "different length"
			}())
	}

	return nil
}