package compliance_test

import (
	"testing"

	"github.com/vendasta/langchaingo/llms/compliance"
	"github.com/vendasta/langchaingo/llms/fake"
)

// ExampleCompliance demonstrates how to use the compliance test suite
// with a provider. This example uses the fake provider.
func TestFakeProviderCompliance(t *testing.T) {
	// Skip this test in normal test runs
	if testing.Short() {
		t.Skip("Skipping compliance test in short mode")
	}

	// Create a fake model for testing with responses that pass compliance tests
	model := fake.NewFakeLLM([]string{
		"Hello, World!",              // For BasicGeneration
		"Your name is Alice.",        // For MultiMessage
		"42",                         // For Temperature (temp=0)
		"42",                         // For Temperature (temp=1)
		"1, 2, 3",                    // For MaxTokens (short response)
		"Monday, Tuesday, Wednesday", // For StopSequences
		"Extra response 1",           // Additional responses if needed
		"Extra response 2",           // Additional responses if needed
		"Extra response 3",           // Additional responses if needed
	})

	// Create and run the compliance suite
	suite := compliance.NewSuite("fake", model)

	// The fake provider doesn't support some features
	suite.SkipTests = map[string]bool{
		"ContextCancellation": true, // Fake provider doesn't check context
		"MaxTokensRespected":  true, // Fake provider doesn't respect max tokens
		"StopWordsRespected":  true, // Fake provider doesn't respect stop words
		"MultipleMessages":    true, // Fake provider doesn't handle conversation history
	}

	suite.Run(t)
}

// TestOpenAIComplianceExample shows how a real provider would implement compliance testing.
func TestOpenAIComplianceExample(t *testing.T) {
	t.Skip("Example only - requires real OpenAI credentials")

	// This is how you would test a real provider:
	//
	// import "github.com/vendasta/langchaingo/llms/openai"
	//
	// llm, err := openai.New()
	// if err != nil {
	//     t.Fatal(err)
	// }
	//
	// suite := compliance.NewSuite("openai", llm)
	// suite.Run(t)
}
