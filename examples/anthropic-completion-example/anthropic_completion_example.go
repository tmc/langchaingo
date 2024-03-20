package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
)

func main() {
	llm, err := anthropic.New(
		anthropic.WithModel("claude-3-opus-20240229"),
		// anthropic.WithHTTPClient(debugHttpClient),
	)
	// note: You would include anthropic.WithModel("claude-2") to use the claude-2 model.
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, "Human: Who was the first man to walk on the moon?\nAssistant:",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println(completion)
}

var debugHttpClient = &http.Client{
	Transport: &logTransport{http.DefaultTransport},
}

type logTransport struct {
	Transport http.RoundTripper
}

// RoundTrip logs the request and response with full contents using httputil.DumpRequest and httputil.DumpResponse
func (t *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", dump)
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", dump)
	return resp, nil
}
