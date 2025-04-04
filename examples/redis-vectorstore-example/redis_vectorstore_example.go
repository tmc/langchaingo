package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/averikitsch/langchaingo/chains"
	"github.com/averikitsch/langchaingo/embeddings"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/ollama"
	"github.com/averikitsch/langchaingo/schema"
	"github.com/averikitsch/langchaingo/vectorstores"
	"github.com/averikitsch/langchaingo/vectorstores/redisvector"
)

func main() {
	redisURL := "redis://127.0.0.1:6379"
	index := "test_redis_vectorstore"

	llm, e := getEmbedding("gemma:2b", "http://127.0.0.1:11434")
	ctx := context.Background()

	store, err := redisvector.New(ctx,
		redisvector.WithConnectionURL(redisURL),
		redisvector.WithIndexName(index, true),
		redisvector.WithEmbedder(e),
	)
	if err != nil {
		log.Fatalln(err)
	}

	data := []schema.Document{
		{PageContent: "Tokyo", Metadata: map[string]any{"population": 9.7, "area": 622}},
		{PageContent: "Kyoto", Metadata: map[string]any{"population": 1.46, "area": 828}},
		{PageContent: "Hiroshima", Metadata: map[string]any{"population": 1.2, "area": 905}},
		{PageContent: "Kazuno", Metadata: map[string]any{"population": 0.04, "area": 707}},
		{PageContent: "Nagoya", Metadata: map[string]any{"population": 2.3, "area": 326}},
		{PageContent: "Toyota", Metadata: map[string]any{"population": 0.42, "area": 918}},
		{PageContent: "Fukuoka", Metadata: map[string]any{"population": 1.59, "area": 341}},
		{PageContent: "Paris", Metadata: map[string]any{"population": 11, "area": 105}},
		{PageContent: "London", Metadata: map[string]any{"population": 9.5, "area": 1572}},
		{PageContent: "Santiago", Metadata: map[string]any{"population": 6.9, "area": 641}},
		{PageContent: "Buenos Aires", Metadata: map[string]any{"population": 15.5, "area": 203}},
		{PageContent: "Rio de Janeiro", Metadata: map[string]any{"population": 13.7, "area": 1200}},
		{PageContent: "Sao Paulo", Metadata: map[string]any{"population": 22.6, "area": 1523}},
	}

	_, err = store.AddDocuments(ctx, data)
	docs, err := store.SimilaritySearch(ctx, "Tokyo", 2,
		vectorstores.WithScoreThreshold(0.5),
	)
	fmt.Println(docs)

	result, err := chains.Run(
		ctx,
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	fmt.Println(result)
}

func getEmbedding(model string, connectionStr ...string) (llms.Model, *embeddings.EmbedderImpl) {
	opts := []ollama.Option{ollama.WithModel(model)}
	if len(connectionStr) > 0 {
		opts = append(opts, ollama.WithServerURL(connectionStr[0]))
	}
	llm, err := ollama.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	e, err := embeddings.NewEmbedder(llm)
	if err != nil {
		log.Fatal(err)
	}
	return llms.Model(llm), e
}
