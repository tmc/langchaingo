package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
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
		opts = append(opts, openai.WithHTTPClient(&http.Client{
			Transport: &logTransport{
				Transport: http.DefaultTransport, // Use http.DefaultTransport as the underlying transport
			},
		}))
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

// logTransport wraps around an existing http.RoundTripper, allowing us to
// intercept and log the request and response.
type logTransport struct {
	Transport http.RoundTripper
}

// RoundTrip executes a single HTTP transaction and logs the request and response.
func (c *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Log the request
	requestDump, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		log.Println("Request:\n" + string(requestDump))
	} else {
		log.Println("Error dumping request:", err)
	}
	// Use the underlying Transport to execute the request
	resp, err := c.Transport.RoundTrip(req)
	if err != nil {
		return nil, err // Return early if there's an error
	}
	// Log the response
	responseDump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		log.Println("Response:\n" + string(responseDump))
	} else {
		log.Println("Error dumping response:", err)
	}
	return resp, err
}
