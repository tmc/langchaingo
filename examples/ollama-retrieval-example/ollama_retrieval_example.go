package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("mistral"))
	if err != nil {
		log.Fatal(err)
	}
	prompt := prompts.NewPromptTemplate(
		`
		What is {{.something}}? Answer this question using the provided context first if it is directly relevant.
		If the question is unrelated, ignore the provided context and answer based solely on the user's query.
		And don't mention any provided context.
		`,
		[]string{"something"},
	)
	chain := chains.NewRetrievalQAFromLLM(llm, retriever{})

	testdata := []string{"foodpanda", "foo", "f4", "panda"}
	var wg sync.WaitGroup
	wg.Add(len(testdata))
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	for i, q := range testdata {
		go func() {
			defer wg.Done()
			query, err := prompt.Format(map[string]any{"something": q})
			if err != nil {
				log.Fatal(err)
			}
			res, err := chain.Call(ctx, map[string]any{
				"query": query,
			})
			if err != nil {
				log.Fatal(err)
			}
			if t, ok := res["text"]; ok {
				fmt.Printf("\n#%d %s\n", i+1, q)
				fmt.Println(t)
			}
		}()
	}
	wg.Wait()
}

type retriever struct{}

func (r retriever) GetRelevantDocuments(_ context.Context, _ string) ([]schema.Document, error) {
	return []schema.Document{
		{PageContent: "foo means 9527."},
		{PageContent: "F4 (Flower Four) was a Taiwanese boy band. The group F4 was formed in 2001 after the Taiwanese drama Meteor Garden that they starred in was widely successful."},
	}, nil
}
