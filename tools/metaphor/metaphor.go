package metaphor

import (
	"context"
	"os"

	"github.com/metaphorsystems/metaphor-go"
	"github.com/tmc/langchaingo/tools"
)

var _ tools.Tool = &API{}

type API struct {
	client  *metaphor.Client
	options []metaphor.ClientOptions
}

type ToolInput struct {
	Operation string
	ReqBody   *metaphor.RequestBody
}

func NewClient(options ...metaphor.ClientOptions) (*API, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(apiKey, options...)
	if err != nil {
		return nil, err
	}

	return &API{
		client:  client,
		options: options,
	}, nil
}

func (tool *API) SetOptions(options ...metaphor.ClientOptions) {
	tool.options = options
}

func (tool *API) Name() string {
	return "Metaphor API Tool"
}

func (tool *API) Description() string {
	return ""
}

// trunk-ignore(golangci-lint/revive)
func (tool *API) Call(ctx context.Context, input string) (string, error) {

	return "", nil
}
