package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/averikitsch/langchaingo/llms"
	"github.com/averikitsch/langchaingo/llms/cache"
	"github.com/averikitsch/langchaingo/llms/cache/inmemory"
	"github.com/averikitsch/langchaingo/llms/ollama"
	"github.com/mitchellh/go-wordwrap"
)

const WIDTH = 80

func main() {
	ctx := context.Background()

	var llm llms.Model

	// base LLM is ollama/llama2.
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatal(err)
	}

	// create a new inmemory cache backend that caches results for one minute.
	mem, err := inmemory.New(ctx, inmemory.WithExpiration(time.Minute))
	if err != nil {
		log.Fatal(err)
	}

	// wrap the base LLM to make it caching.
	llm = cache.New(llm, mem)

	// repeat the same query a few times. The first time it'll query the base LLM but
	// subsequent times the result is returned from the cache.
	for i := 0; i < 3; i++ {
		start := time.Now()

		completion, err := llms.GenerateFromSinglePrompt(ctx, llm,
			"Human: Who was the first man to walk on the moon?\nAssistant:",
			llms.WithTemperature(0.8),
		)
		if err != nil {
			log.Fatal(err)
		}

		if i > 0 {
			fmt.Println(strings.Repeat("=", WIDTH))
		}

		completion = wordwrap.WrapString(completion, WIDTH)

		fmt.Printf("## Iteration #%d\n\n%s\n\n(took %v)\n",
			i, completion, time.Since(start))
	}
}
