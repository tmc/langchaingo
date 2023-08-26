package duckduckgo

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/duckduckgo/internal"
)

// DefaultUserAgent defines a default value for user-agent header.
const DefaultUserAgent = "github.com/tmc/langchaingo/tools/duckduckgo"

// Tool defines a tool implementation for the DuckDuckGo Search.
type Tool struct {
	CallbacksHandler callbacks.Handler
	client           *internal.Client
}

var _ tools.Tool = Tool{}

// New initializes a new DuckDuckGo Search tool with arguments for setting a
// max results per search query and a value for the user agent header.
func New(maxResults int, userAgent string) (*Tool, error) {
	return &Tool{
		client: internal.New(maxResults, userAgent),
	}, nil
}

// Name returns a name for the tool.
func (t Tool) Name() string {
	return "DuckDuckGo Search"
}

// Description returns a description for the tool.
func (t Tool) Description() string {
	return `
	"A wrapper around DuckDuckGo Search."
	"Free search alternative to google and serpapi."
	"Input should be a search query."`
}

// Call performs the search and return the result.
func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	result, err := t.client.Search(ctx, input)
	if err != nil {
		if errors.Is(err, internal.ErrNoGoodResult) {
			return "No good DuckDuckGo Search Results was found", nil
		}
		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}
