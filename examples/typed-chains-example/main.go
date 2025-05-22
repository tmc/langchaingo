package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// Example struct for structured output
type PersonInfo struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func main() {
	ctx := context.Background()

	// Create a fake LLM for demonstration
	llm := fake.New()

	fmt.Println("=== Typed Chains Example ===\n")

	// Example 1: Simple String Chain
	fmt.Println("1. Simple String Chain:")
	stringChainExample(ctx, llm)

	// Example 2: Structured Output Chain  
	fmt.Println("\n2. Structured Output Chain:")
	structuredChainExample(ctx, llm)

	// Example 3: Legacy Compatibility
	fmt.Println("\n3. Legacy Compatibility:")
	legacyCompatibilityExample(ctx, llm)

	// Example 4: Type Safety Demonstration
	fmt.Println("\n4. Type Safety Benefits:")
	typeSafetyExample(ctx, llm)
}

func stringChainExample(ctx context.Context, llm *fake.LLM) {
	// Configure fake LLM response
	llm.SetResponse("Paris is the capital of France.")

	// Create a typed string chain
	prompt := prompts.NewPromptTemplate(
		"Question: {{.question}}\nAnswer:",
		[]string{"question"},
	)

	stringChain := chains.NewStringChain(llm, prompt)

	// Use with type safety - no casting needed!
	answer, metadata, err := stringChain.Call(ctx, map[string]any{
		"question": "What is the capital of France?",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Answer: %s\n", answer) // answer is already a string!
	fmt.Printf("Metadata keys: %v\n", getKeys(metadata))
	fmt.Printf("Output type: %s\n", stringChain.OutputType())
}

func structuredChainExample(ctx context.Context, llm *fake.LLM) {
	// Configure fake LLM response
	llm.SetResponse(`{"name": "John", "age": 30, "city": "Paris"}`)

	// Create a structured output parser
	structParser := outputparser.NewStructured[PersonInfo](
		"PersonInfo",
		"Extract person information",
		map[string]string{
			"name": "person's name",
			"age":  "person's age",
			"city": "person's city",
		},
	)

	// Create a typed chain with structured output
	prompt := prompts.NewPromptTemplate(
		"Extract person info from: {{.text}}\n{{.format_instructions}}",
		[]string{"text"},
	)

	structuredChain := chains.NewTypedLLMChain(llm, prompt, structParser)

	// Use with full type safety
	person, metadata, err := structuredChain.Call(ctx, map[string]any{
		"text": "John is 30 years old and lives in Paris",
	})
	if err != nil {
		log.Fatal(err)
	}

	// person is PersonInfo struct - no casting needed!
	fmt.Printf("Person: %+v\n", person)
	fmt.Printf("Name: %s, Age: %d, City: %s\n", person.Name, person.Age, person.City)
	fmt.Printf("Output type: %s\n", structuredChain.OutputType())
	fmt.Printf("Metadata keys: %v\n", getKeys(metadata))
}

func legacyCompatibilityExample(ctx context.Context, llm *fake.LLM) {
	llm.SetResponse("42")

	// Create a typed chain
	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	typedChain := chains.NewStringChain(llm, prompt)

	// Convert to legacy interface
	legacyChain := typedChain.AsLegacyChain()

	// Use with legacy Chain interface
	result, err := chains.Call(ctx, legacyChain, map[string]any{
		"input": "What is the answer to everything?",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Legacy usage requires manual casting
	answer, ok := result["result"].(string)
	if !ok {
		log.Fatal("unexpected type")
	}

	fmt.Printf("Legacy result: %s\n", answer)
	fmt.Printf("Result keys: %v\n", getKeys(result))

	// Compare with typed usage
	typedAnswer, _, err := chains.TypedCall(ctx, typedChain, map[string]any{
		"input": "What is the answer to everything?",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Typed result: %s (no casting needed!)\n", typedAnswer)
}

func typeSafetyExample(ctx context.Context, llm *fake.LLM) {
	llm.SetResponse("Hello, World!")

	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	chain := chains.NewStringChain(llm, prompt)

	// Type-safe usage with helper functions
	result, err := chains.TypedRun(ctx, chain, "Say hello")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TypedRun result: %s\n", result)

	// Type-safe prediction
	prediction, err := chains.TypedPredict(ctx, chain, map[string]any{
		"input": "Predict something",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TypedPredict result: %s\n", prediction)

	// Demonstrate compile-time type safety
	fmt.Println("\nCompile-time type safety:")
	fmt.Printf("Result type: %T\n", result)
	fmt.Printf("No runtime type assertions needed!\n")

	// Example of what you CAN'T do (would be compile error):
	// var wrongType int = result  // Compile error: cannot use result (string) as int
}

// Helper function to get map keys
func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Example of chain composition with type safety
func chainCompositionExample() {
	fmt.Println("\n=== Chain Composition Example ===")
	fmt.Println("This example shows how typed chains enable safer composition:")

	// Pseudo-code for typed chain composition
	fmt.Println(`
// Type-safe chain composition
func ComposeChains[T, U any](
    first chains.TypedChain[T], 
    second chains.TypedChain[U],
    mapper func(T, map[string]any) map[string]any,
) chains.TypedChain[U] {
    // Implementation would ensure type safety
}

// Usage:
stringChain := chains.NewStringChain(llm, prompt1)
structChain := chains.NewTypedLLMChain[PersonInfo](llm, prompt2, parser)

// Compose with type safety
composed := ComposeChains(stringChain, structChain, func(s string, meta map[string]any) map[string]any {
    return map[string]any{"text": s}
})

// Result is guaranteed to be PersonInfo
person, metadata, err := composed.Call(ctx, inputs)
`)
}