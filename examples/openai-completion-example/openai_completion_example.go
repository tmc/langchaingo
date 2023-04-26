package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call("The first man to walk on the moon", []string{"Armstrong"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
