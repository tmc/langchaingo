package zapier

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/zapier/internal"
)

type description struct {
	Params            []string
	ZapierDescription string
}

type Tool struct {
	CallbacksHandler callbacks.Handler
	client           *internal.Client
	name             string
	description      string
	actionID         string
	params           map[string]string
}

var _ tools.Tool = Tool{}

type ToolOptions struct {
	Name        string
	ActionID    string
	Params      map[string]string
	APIKey      string
	AccessToken string
	UserAgent   string
	Client      *internal.Client
}

func (tOpts ToolOptions) Validate() error {
	return nil
}

/*
New creates a new Zapier NLA Tool that is Tool Interface compliant.
*/
func New(opts ToolOptions) (*Tool, error) {
	err := opts.Validate()
	if err != nil {
		return nil, err
	}

	if opts.Client != nil {
		opts.Client, err = internal.NewClient(internal.ClientOptions{
			APIKey:      opts.APIKey,
			AccessToken: opts.AccessToken,
			UserAgent:   opts.UserAgent,
		})
		if err != nil {
			return nil, err
		}
	}

	t := &Tool{
		client:   opts.Client,
		name:     opts.Name,
		actionID: opts.ActionID,
		params:   opts.Params,
	}
	t.description = t.createDescription()
	return t, nil
}

func (t Tool) Name() string {
	return t.name
}

func (t Tool) Description() string {
	return t.description
}

// Schema returns OpenAPI schema for the tool.
func (t Tool) Schema() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"instructions": map[string]any{
				"type":        "string",
				"description": "Instructions to execute.",
			},
		},
		"required": []string{"instructions"},
	}
}

func (t Tool) Call(ctx context.Context, input any) (string, error) {
	instructions, ok := input.(map[string]any)["instructions"].(string)
	if !ok {
		return "", fmt.Errorf("invalid input: %v", input)
	}

	return t.call(ctx, instructions)
}

func (t Tool) call(ctx context.Context, instructions string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, instructions)
	}

	result, err := t.client.ExecuteAsString(ctx, t.actionID, instructions, t.params)
	if err != nil {
		if t.CallbacksHandler != nil {
			t.CallbacksHandler.HandleToolError(ctx, err)
		}
		return "", err
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}

func (t Tool) createDescription() string {
	tmpl, err := template.New("").Parse(_baseZapierDescription)
	if err != nil {
		panic(err)
	}
	var bytes bytes.Buffer

	paramNames := make([]string, 0, len(t.params))
	for k := range t.params {
		paramNames = append(paramNames, k)
	}

	desc := description{
		Params:            paramNames,
		ZapierDescription: t.name,
	}

	err = tmpl.Execute(&bytes, desc)
	if err != nil {
		panic(err)
	}

	return bytes.String()
}
