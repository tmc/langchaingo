package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/huggingface"
)

func main() {
	// load .env with joho/godotenv
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	llm, err := huggingface.New()
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call("What would be a good company name be for name a company that makes colorful socks that is not sock monkey?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
