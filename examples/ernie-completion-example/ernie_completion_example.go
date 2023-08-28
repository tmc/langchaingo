package main

import (
	"context"
	"fmt"
	"log"

	ernieembedding "github.com/tmc/langchaingo/embeddings/ernie"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ernie"
)

func main() {
	llm, err := ernie.New(ernie.WithModelName(ernie.ModelNameERNIEBot))
	// note:
	// You would include ernie.WithAKSK(apiKey,secretKey) to use specific auth info.
	// You would include ernie.WithModelName(ernie.ModelNameERNIEBot) to use the ERNIE-Bot model.
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llm.Call(ctx, "介绍一下你自己",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Println(string(chunk))
			return nil
		}),
	)

	if err != nil {
		log.Fatal(err)
	}

	_ = completion

	// embedding
	embedding, _ := ernieembedding.NewErnie()

	emb, err := embedding.EmbedDocuments(ctx, []string{"你好"})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Embedding-V1:", len(emb), len(emb[0]))
}
