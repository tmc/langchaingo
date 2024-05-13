package serpapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/serpapi/internal"
)

var ErrMissingToken = errors.New("missing the serpapi API key, set it in the SERPAPI_API_KEY environment variable")

type Tool struct {
	CallbacksHandler callbacks.Handler
	client           *internal.Client
}

func (t Tool) Schema() tools.Schema {
	return tools.Schema{
		Type: "object",
		Properties: map[string]any{
			"query": map[string]any{
				"title": "query",
				"type":  "string",
			},
		},
		Required: []string{"query"},
	}
}

var _ tools.Tool = Tool{}

// New creates a new serpapi tool to search on internet.
func New(opts ...Option) (*Tool, error) {
	options := &options{
		apiKey: os.Getenv("SERPAPI_API_KEY"),
	}

	for _, opt := range opts {
		opt(options)
	}

	if options.apiKey == "" {
		return nil, ErrMissingToken
	}

	return &Tool{
		client: internal.New(options.apiKey),
	}, nil
}

func (t Tool) Name() string {
	return "GoogleSearch"
}

func (t Tool) Description() string {
	return `
	"A wrapper around Google Search. "
	"Useful for when you need to answer questions about current events. "
	"Always one of the first options when you need to find information on internet"
	"Input should be a search query."`
}

func (t Tool) Call(ctx context.Context, input map[string]any) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	query, ok := input["query"].(string)
	if !ok {
		return "", fmt.Errorf("could not obtain `query` in %v", input)
	}

	result, err := t.client.Search(ctx, query)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No good Google Search Results was found", nil
		}

		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}

		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return strings.Join(strings.Fields(result), " "), nil
}
