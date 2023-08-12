package metaphor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/metaphor/client"
)

type MetaphorContents struct {
	client *client.MetaphorClient
}

var _ tools.Tool = &MetaphorContents{}

func NewDocuments(options ...client.ClientOptions) (*MetaphorContents, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingToken
	}

	client, err := client.NewClient(apiKey, options...)
	if err != nil {
		return nil, err
	}

	return &MetaphorContents{
		client: client,
	}, nil

}

func (tool *MetaphorContents) Name() string {
	return "Metaphor Contents Extractor"
}

func (tool *MetaphorContents) Description() string {
	return `
	To be used with Metaphor Search and/or Metaphor Links Search Tool.
	Retrieve contents of web pages based on a list of ID strings.
	Input should be a list(or a single ID) of comma seperated IDs,
	obtained from either Metaphor Search or Metaphor Search Links tool.
	Expected input format:
	"8U71IlQ5DUTdsherhhYA,9segZCZGNjjQB2yD2uyK,..."`
}

func (tool *MetaphorContents) Call(ctx context.Context, input string) (string, error) {
	ids := strings.Split(input, ",")
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	contents, err := tool.client.GetContents(ctx, ids)
	if err != nil {
		if errors.Is(err, client.ErrNoContentExtracted) {
			return "Metaphor Extractor didn't return any results", nil
		}
		return "", err
	}

	return tool.formatContents(contents), nil
}

func (tool *MetaphorContents) formatContents(response *client.ContentsResponse) string {
	formattedResults := ""

	for _, result := range response.Contents {
		formattedResults += fmt.Sprintf("Title: %s\nContent: %s\nURL: %s\n\n", result.Title, result.Extract, result.Url)
	}

	return formattedResults
}
