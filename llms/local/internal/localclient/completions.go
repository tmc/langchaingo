package localclient

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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

func (c *Client) createStreamCompletion(ctx context.Context, payload *completionPayload) (chan string, chan bool, error) {
	ch := make(chan string)
	bch := make(chan bool)
	c.Args = append(c.Args, payload.Prompt)

	cmd := exec.CommandContext(ctx, c.BinPath, c.Args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ch, bch, errors.New(fmt.Sprintf("Error creating stdout: %s", err))
	}

	if err := cmd.Start(); err != nil {
		return ch, bch, errors.New(fmt.Sprintf("Error starting command: %s", err))
	}

	reader := bufio.NewReader(stdout)
	buf := make([]byte, 1) // Read one byte at a time

	for {
		_, err := reader.Read(buf)
		if err != nil {
			break // Exit loop on error or EOF
		}
		ch <- string(buf)
	}

	//receiver is polling for messages on channel so cant just return
	if err := cmd.Wait(); err != nil {
		bch <- true
		return ch, bch, errors.New(fmt.Sprintf("Error waiting for command:", err))
	}

	//exit as normal
	bch <- true
	return ch, bch, nil
}
