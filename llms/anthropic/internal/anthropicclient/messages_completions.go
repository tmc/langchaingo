package anthropicclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesCompletionPayload struct {
	Model        string         `json:"model"`
	Messages     []*ChatMessage `json:"messages"`
	SystemPrompt string         `json:"system,omitempty"`
	Temperature  float64        `json:"temperature"`
	MaxTokens    int            `json:"max_tokens,omitempty"`
	TopP         float64        `json:"top_p,omitempty"`
	StopWords    []string       `json:"stop_sequences,omitempty"`

	Stream        bool                                          `json:"stream,omitempty"`
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type MessagesCompletionResponsePayload struct {
	Content []struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text,omitempty"`
	} `json:"content,omitempty"`
	Id           string `json:"id,omitempty"`
	Type         string `json:"type,omitempty"`
	Role         string `json:"role,omitempty"`
	Model        string `json:"model,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	Usage        struct {
		InputTokens  int `json:"input_tokens,omitempty"`
		OutputTokens int `json:"output_tokens,omitempty"`
	} `json:"usage,omitempty"`
}

type StreamPayload struct {
	Type         string       `json:"type"`
	Message      *Message     `json:"message,omitempty"`
	ContentBlock ContentBlock `json:"content_block,omitempty"`
	Index        int          `json:"index,omitempty"`
	Delta        Delta        `json:"delta,omitempty"`
	Usage        Usage        `json:"usage"`
}

type Message struct {
	ID         string   `json:"id"`
	Type       string   `json:"type"`
	Role       string   `json:"role"`
	Content    []string `json:"content"`
	Model      string   `json:"model"`
	StopReason string   `json:"stop_reason"`
	Usage      Usage    `json:"usage"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`

	StopSequence string `json:"stop_sequence,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func (c *Client) setMessagesCompletionDefaults(payload *messagesCompletionPayload) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 256
	}

	if len(payload.StopWords) == 0 {
		payload.StopWords = nil
	}

	switch {
	// Prefer the model specified in the payload.
	case payload.Model != "":
	// If no model is set in the payload, take the one specified in the client.
	case c.Model != "":
		payload.Model = c.Model
	// Fallback: use the default model
	default:
		payload.Model = defaultCompletionModel
	}
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
}

// nolint:lll
func (c *Client) createMessagesCompletion(ctx context.Context, payload *messagesCompletionPayload) (*MessagesCompletionResponsePayload, error) {
	c.setMessagesCompletionDefaults(payload)

	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}

	url := fmt.Sprintf("%s/messages", c.baseURL)
	// Build request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	// Send request
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg) // nolint:goerr113
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	}
	if payload.StreamingFunc != nil {
		// Read chunks
		return parseStreamingMessagesCompletionResponse(ctx, r, payload)
	}

	// Parse response
	var response MessagesCompletionResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

func parseStreamingMessagesCompletionResponse(ctx context.Context, r *http.Response, payload *messagesCompletionPayload) (*MessagesCompletionResponsePayload, error) {
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan *StreamPayload)

	go func() {
		defer close(responseChan)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			streamPayload := &StreamPayload{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				log.Fatalf("failed to decode stream payload: %v", err)
			}
			responseChan <- streamPayload
		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()

	// Parse response
	response := MessagesCompletionResponsePayload{}
	var content string
	for streamPayload := range responseChan {
		switch streamPayload.Type {
		case "message_start":
			response.Id = streamPayload.Message.ID
			response.Type = streamPayload.Message.Type
			response.Role = streamPayload.Message.Role
			response.Model = streamPayload.Message.Model
			response.StopReason = streamPayload.Message.StopReason
			response.Usage.InputTokens = streamPayload.Message.Usage.InputTokens
		case "content_block_delta":
			content += streamPayload.Delta.Text

			payload.StreamingFunc(ctx, []byte(streamPayload.Delta.Text))
		case "message_delta":
			if streamPayload.Delta.StopReason != "" {
				response.StopReason = streamPayload.Delta.StopReason
			}
			if streamPayload.Usage.OutputTokens != 0 {
				response.Usage.OutputTokens = streamPayload.Usage.OutputTokens
			}
		}

	}

	response.Content = []struct {
		Type string `json:"type,omitempty"`
		Text string `json:"text,omitempty"`
	}{
		{
			Type: "text",
			Text: content,
		},
	}

	return &response, nil
}
