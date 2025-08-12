package main

import (
	"context"
	"fmt"
	"os"

	"github.com/vendasta/langchaingo/llms/maritaca"
)

func main() {
	token := os.Getenv("MARITACA_KEY")

	opts := []maritaca.Option{
		maritaca.WithToken(token),
		maritaca.WithModel("sabia-2-medium"),
	}
	llm, err := maritaca.New(opts...)
	if err != nil {
		panic(err)
	}

	prompt := "How many people live in Brazil?"

	resp, err := llm.Call(context.Background(), prompt)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
