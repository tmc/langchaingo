package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/vertexai"
	"github.com/tmc/langchaingo/textsplitter"
)

func main() {
	llm, err := vertexai.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
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
	outputValues, err := chains.Call(ctx, llmSummarizationChain, map[string]any{"input_documents": docs})
	if err != nil {
		log.Fatal(err)
	}
	out := outputValues["text"].(string)
	fmt.Println(out)

	// Output:
	// Large language models are a type of deep learning model that can understand, learn,
	// summarize, translate, predict, and generate text and other content based on knowledge
	// gained from massive datasets. They are used in a variety of applications, including
	// natural language processing, healthcare, and software development.
}
