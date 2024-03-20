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

type messagePayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	System      string        `json:"system,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	StopWords   []string      `json:"stop_sequences,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature"`
	TopP        float64       `json:"top_p,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

type MessageResponsePayload struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	ID           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Type         string `json:"type"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	}
}

func (c *Client) setMessageDefaults(payload *messagePayload) {
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
		payload.Model = defaultModel
	}
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
}

func (c *Client) createMessage(ctx context.Context, payload *messagePayload) (*MessageResponsePayload, error) {
	c.setMessageDefaults(payload)

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
		return parseStreamingMessageResponse(ctx, r, payload)
	}

	// Parse response
	var response MessageResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

func parseStreamingMessageResponse(ctx context.Context, r *http.Response, payload *messagePayload) (*MessageResponsePayload, error) {
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan *MessageResponsePayload)
	go func() {
		defer close(responseChan)
		var response MessageResponsePayload
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			var event map[string]interface{}
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&event)
			if err != nil {
				log.Fatalf("failed to decode stream event: %v", err)
			}
			eventType, ok := event["type"].(string)
			if !ok {
				log.Println("missing event type")
				continue
			}
			switch eventType {
			case "message_start":
				message := event["message"].(map[string]interface{})
				response = MessageResponsePayload{
					ID:    message["id"].(string),
					Model: message["model"].(string),
					Role:  message["role"].(string),
					Type:  message["type"].(string),
				}
				response.Usage.InputTokens = int(message["usage"].(map[string]interface{})["input_tokens"].(float64))
			case "content_block_start":
				index := int(event["index"].(float64))
				if len(response.Content) <= index {
					response.Content = append(response.Content, struct {
						Text string `json:"text"`
						Type string `json:"type"`
					}{})
				}
			case "content_block_delta":
				index := int(event["index"].(float64))
				delta := event["delta"].(map[string]interface{})
				if delta["type"] == "text_delta" {
					response.Content[index].Text += delta["text"].(string)
				}
				if payload.StreamingFunc != nil {
					err := payload.StreamingFunc(ctx, []byte(delta["text"].(string)))
					if err != nil {
						return
					}
				}
			case "content_block_stop":
				// Nothing to do here
			case "message_delta":
				delta := event["delta"].(map[string]interface{})
				if stopReason, ok := delta["stop_reason"]; ok {
					response.StopReason = stopReason.(string)
				}
				usage := event["usage"].(map[string]interface{})
				if outputTokens, ok := usage["output_tokens"]; ok {
					response.Usage.OutputTokens = int(outputTokens.(float64))
				}
			case "message_stop":
				responseChan <- &response
			case "ping":
				// Nothing to do here
			default:
				log.Printf("unknown event type: %s", eventType)
			}
		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()
	var lastResponse *MessageResponsePayload
	for response := range responseChan {
		lastResponse = response
	}
	return lastResponse, nil
}
