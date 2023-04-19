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
	c.args = append(c.args, payload.Prompt)

	out, err := exec.CommandContext(ctx, c.binPath, c.args...).Output()
	if err != nil {
		return nil, err
	}

	return &completionResponsePayload{
		Response: string(out),
	}, nil
}
