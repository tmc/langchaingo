package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/starmvp/langchaingo/llms"
	"github.com/starmvp/langchaingo/llms/openai"
)

var flagModel = flag.String("model", "o1-preview", "model to use (e.g. 'o1-preview', 'o1-mini')")

func main() {
	flag.Parse()
	llm, err := openai.New(
		openai.WithModel(*flagModel),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, `
I want to build a Go app that takes user questions and looks them up in a 
database where they are mapped to answers. If there is a close match, it retrieves
the matched answer. If there isn't, it asks the user to provide an answer and 
stores the question/answer pair in the database. Make a plan for the directory 
structure you'll need, then return each file in full. Only supply your reasoning 
at the beginning and end, not throughout the code.`),
	}
	fmt.Println("Generating content...")
	output, err := llm.GenerateContent(ctx, content,
		llms.WithMaxTokens(4000),
		llms.WithTemperature(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output.Choices[0].Content)

	fmt.Println("\ngeneration info:")
	json.NewEncoder(os.Stdout).Encode(output.Choices[0].GenerationInfo)
}
