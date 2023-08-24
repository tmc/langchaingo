package metaphor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/metaphorsystems/metaphor-go"
	"github.com/tmc/langchaingo/tools"
)

// Documents defines a tool implementation for the Metaphor Web scrapper.
type Documents struct {
	client  *metaphor.Client
	options []metaphor.ClientOptions
}

var _ tools.Tool = &Documents{}

// NewDocuments creates a new instance of the Documents struct.
//
// The function takes in optional metaphorm.ClientOptions as parameters.
// It returns a pointer to a Documents struct and an error.
func NewDocuments(options ...metaphor.ClientOptions) (*Documents, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(apiKey, options...)
	if err != nil {
		return nil, err
	}

	return &Documents{
		client:  client,
		options: options,
	}, nil
}

// SetOptions sets the options for the Documents struct.
//
// It takes in variadic parameter(s) of type `metaphor.ClientOptions`.
func (tool *Documents) SetOptions(options ...metaphor.ClientOptions) {
	tool.options = options
}

// Name returns the name of the Documents tool.
//
// It does not take any parameters.
// It returns a string, which is the name of the tool.
func (tool *Documents) Name() string {
	return "Metaphor Contents Extractor"
}

// Description returns the contents of web pages based on a list of ID strings.
//
// It is designed to be used with Metaphor Search and/or Metaphor Links Search Tool.
// The expected input format is a list of ID strings obtained from either Metaphor Search or Metaphor Search Links tool.
// The function returns a string.
func (tool *Documents) Description() string {
	return `
	To be used with Metaphor Search and/or Metaphor Links Search Tool.
	Retrieve contents of web pages based on a list of ID strings.
	obtained from either Metaphor Search or Metaphor Search Links tool.
	Expected input format:
	"8U71IlQ5DUTdsherhhYA,9segZCZGNjjQB2yD2uyK,..."`
}

// Call calls the Documents API with the given input and returns the formatted contents.
//
// The input is a string that contains a comma-separated list of IDs.
//
// It returns a string which represents the formatted contents and an error if any.
func (tool *Documents) Call(ctx context.Context, input string) (string, error) {
	ids := strings.Split(input, ",")
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	contents, err := tool.client.GetContents(ctx, ids)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoContentExtracted) {
			return "Metaphor Extractor didn't return any results", nil
		}
		return "", err
	}

	return tool.formatContents(contents), nil
}

func (tool *Documents) formatContents(response *metaphor.ContentsResponse) string {
	formattedResults := ""

	for _, result := range response.Contents {
		formattedResults += fmt.Sprintf("Title: %s\nContent: %s\nURL: %s\n\n", result.Title, result.Extract, result.URL)
	}

	return formattedResults
}
