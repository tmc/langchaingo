package localclient

import (
	"context"
	"os/exec"
)

type completionPayload struct {
	Prompt string `json:"prompt"`
}

type completionResponsePayload struct {
	Response string
}

func (c *Client) createCompletion(ctx context.Context, payload *completionPayload) (*completionResponsePayload, error) {
	// Append the prompt to the args
	c.Args = append(c.Args, payload.Prompt)

	// #nosec G204
	out, err := exec.CommandContext(ctx, c.BinPath, c.Args...).Output()
	if err != nil {
		return nil, err
	}

	return &completionResponsePayload{
		Response: string(out),
	}, nil
}
