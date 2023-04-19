package main

import (
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms/local"
)

func main() {
	llm, err := local.New()
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call("How many sides does a square have?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
