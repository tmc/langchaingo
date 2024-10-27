package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/starmvp/langchaingo/chains"
	"github.com/starmvp/langchaingo/documentloaders"
	"github.com/starmvp/langchaingo/llms/mistral"
	"github.com/starmvp/langchaingo/textsplitter"
)

func main() {
	ctx := context.Background()
	llm, err := mistral.New(mistral.WithAPIKey("API_KEY_GOES_HERE"), mistral.WithModel("open-mistral-7b"))
	if err != nil {
		log.Fatal(err)
	}

	llmSummarizationChain := chains.LoadRefineSummarization(llm)
	doc := `AI applications are summarizing articles, writing stories and 
	engaging in long conversations — and large language models are doing 
	the heavy lifting.
	
	A large language model, or LLM, is a deep learning model that can 
	understand, learn, summarize, translate, predict, and generate text and other 
	content based on knowledge gained from massive datasets.
	
	Large language models - successful applications of 
	transformer models. They aren’t just for teaching AIs human languages, 
	but for understanding proteins, writing software code, and much, much more.
	
	In addition to accelerating natural language processing applications — 
	like translation, chatbots, and AI assistants — large language models are 
	used in healthcare, software development, and use cases in many other fields.`
	docs, err := documentloaders.NewText(strings.NewReader(doc)).LoadAndSplit(ctx,
		textsplitter.NewRecursiveCharacter(),
	)
	if err != nil {
		return
	}
	outputValues, err := chains.Call(ctx, llmSummarizationChain, map[string]any{"input_documents": docs})
	if err != nil {
		log.Fatal(err)
	}
	out := outputValues["text"].(string)
	fmt.Println(out)

	// Output:
	// Large language models (LLMs), a type of transformer model, go beyond teaching AIs human languages.
	// They are used in various applications, including understanding proteins, writing software code, and
	// more. These models not only accelerate natural language processing for tasks like translation,
	// chatbots, and AI assistants but also contribute significantly to healthcare, software development,
	// and other fields.
}
