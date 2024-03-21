package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http/httputil"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

var flagDebugHTTP = flag.Bool("debug-http", true, "enable debugging of HTTP requests and responses")

func main() {
	// Demonstrates how to use a custom HTTP client to log requests and responses.
	flag.Parse()
	var opts []openai.Option
	if *flagDebugHTTP {
		opts = append(opts, openai.WithHTTPClient(httputil.DebugHTTPClient))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx,
		llm,
		"The first man to walk on the moon",
		llms.WithTemperature(0.0),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(completion)
}
