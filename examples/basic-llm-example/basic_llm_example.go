package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	// load .env with joho/godotenv
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	completion, err := llm.Call("What would be a good company name a company that makes colorful socks, besides socktastic?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(completion)
}
