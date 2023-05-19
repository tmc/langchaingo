package serpapi

import (
	"strings"

	"github.com/tmc/langchaingo/exp/tools"
	"github.com/tmc/langchaingo/exp/tools/serpapi/internal"
)

type Tool struct {
	client *internal.SerpapiClient
}

var _ tools.Tool = Tool{}

// New creates a new serpapi tool to search on internet.
func New() (*Tool, error) {
	client, err := internal.New()
	if err != nil {
		return nil, err
	}

	return &Tool{
		client: client,
	}, nil
}

func (t Tool) Name() string {
	return "Google Search"
}

func (t Tool) Description() string {
	return `"A wrapper around Google Search. "
	"Useful for when you need to answer questions about current events. "
	"Always one of the first options when you need to find information on internet"
	"Input should be a search query."`
}

func (t Tool) Call(input string) (string, error) {
	result, err := t.client.Search(input)
	if err != nil {
		return "", err
	}
	if len(result) == 0 {
		return "No good Google Search Result was found", nil
	}
	return strings.Join(strings.Fields(result), " "), nil
}
