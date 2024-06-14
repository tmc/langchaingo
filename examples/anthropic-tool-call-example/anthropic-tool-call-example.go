package main

import (
	"github.com/tmc/langchaingo/llms/anthropic"
	"log"
)

func main() {
	_, err := anthropic.New(
		anthropic.WithModel("claude-3-haiku-20240307"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
