package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	// Define the appropriate schema that the LLM should use to format its response.
	type dish struct {
		Name      string
		NumPieces int `describe:"number of pieces in an order"`
	}
	type dishes struct {
		Dishes []dish
	}

	// Create the outputparser.
	var schema dishes
	definedParser, parseErr := outputparser.NewDefined(schema)
	if parseErr != nil {
		log.Fatal(parseErr)
	}

	ctx := context.TODO()

	// Initialize an LLM.
	opts := []googleai.Option{
		googleai.WithAPIKey(os.Getenv("GOOGLE_API_KEY")),
	}
	llm, llmErr := googleai.New(context.TODO(), opts...)
	if llmErr != nil {
		log.Fatal(llmErr)
	}

	// Configure the prompt.
	question := "List the top 10 most popular restaurant dishes in the world"
	prompt := fmt.Sprintf("%s\n%s", question, definedParser.GetFormatInstructions())
	tmpl := prompts.NewPromptTemplate(prompt, []string{})

	// Create the LLM chain.
	chain := chains.NewLLMChain(llm, tmpl)
	chain.OutputParser = definedParser // Set the chain to use Defined outputparser
	output, chainErr := chain.Call(ctx, map[string]any{})
	if chainErr != nil {
		log.Fatal(chainErr)
	}
	result, ok := output["text"].(dishes)
	if !ok {
		log.Fatal("output is not a dishes struct")
	}

	// Verify that the output was extracted into the proper structure.
	fmt.Println(question)
	for i, d := range result.Dishes {
		fmt.Printf("%d. %s (%d)\n", i+1, d.Name, d.NumPieces)
	}

	// Sample output:
	// List the top 10 most popular restaurant dishes in the world
	// 1. Pizza (8)
	// 2. Hamburger (1)
	// 3. Sushi (6)
	// 4. Pasta (1)
	// 5. Fried Chicken (4)
	// 6. Tacos (3)
	// 7. Burritos (1)
	// 8. Noodles (1)
	// 9. Soup (1)
	// 10. Salad (1)
}
