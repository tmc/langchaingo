package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
)

func main() {

	ctx := context.Background()

	//load data from PDF to add context on prompt
	docsFromPdf, err := fetchDocumentsFromPdf(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	log.Default().Printf("Documents find on the PDF : %d", len(docsFromPdf))

	//initialization of the llm model (with tinydolphin)
	llm, err := ollama.New(ollama.WithModel("tinydolphin"))
	if err != nil {
		log.Fatal(err)
	}
	stuffQAChain := chains.LoadStuffQA(llm)

	prompt := "What is Langchain ? A short response"

	log.Default().Printf("Prompt: %s \n\r", prompt)
	//request the llm with prompt and context
	answer, err := chains.Call(ctx, stuffQAChain, map[string]any{
		"input_documents": docsFromPdf,
		"question":        prompt,
		"temperature":     0.5,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Default().Printf("Answer: %s", answer["text"])

}

// Fetch documents from pdf file (example.pdf).
// The function open and load the pdf, extract documents from pdf pages
// and return documents will use to enrich prompt with context
func fetchDocumentsFromPdf(context context.Context) ([]schema.Document, error) {
	file, err := os.Open("example.pdf")
	if err != nil {
		log.Default().Println(err)
		return nil, errors.New("error opening file")
	}

	fileStat, err := file.Stat()
	if err != nil {
		log.Default().Println(err)
		return nil, errors.New("error to retrieve file stats")
	}

	pdf := documentloaders.NewPDF(file, fileStat.Size())

	docs, err := pdf.Load(context)
	if err != nil {
		log.Default().Println(err)
		return nil, errors.New("error to load pdf file")
	}

	return docs, nil
}
