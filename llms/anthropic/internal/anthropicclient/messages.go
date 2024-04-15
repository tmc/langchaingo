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
	} `json:"usage"`
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

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := c.do(ctx, "/messages", payloadBytes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	if payload.StreamingFunc != nil {
		return parseStreamingMessageResponse(ctx, resp, payload)
	}

	var response MessageResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

type MessageEvent struct {
	Response *MessageResponsePayload
	Err      error
}

func parseStreamingMessageResponse(ctx context.Context, r *http.Response, payload *messagePayload) (*MessageResponsePayload, error) {
	scanner := bufio.NewScanner(r.Body)
	eventChan := make(chan MessageEvent)

	go func() {
		defer close(eventChan)
		var response MessageResponsePayload
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			event, err := parseStreamEvent(data)
			if err != nil {
				eventChan <- MessageEvent{Response: nil, Err: fmt.Errorf("failed to parse stream event: %w", err)}
				return
			}
			response, err = processStreamEvent(ctx, event, payload, response, eventChan)
			if err != nil {
				eventChan <- MessageEvent{Response: nil, Err: fmt.Errorf("failed to process stream event: %w", err)}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			eventChan <- MessageEvent{Response: nil, Err: fmt.Errorf("issue scanning response: %w", err)}
		}
	}()

	var lastResponse *MessageResponsePayload
	for event := range eventChan {
		if event.Err != nil {
			return nil, event.Err
		}
		lastResponse = event.Response
	}
	return lastResponse, nil
}

func parseStreamEvent(data string) (map[string]interface{}, error) {
	var event map[string]interface{}
	err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&event)
	return event, err
}

func processStreamEvent(ctx context.Context, event map[string]interface{}, payload *messagePayload, response MessageResponsePayload, eventChan chan<- MessageEvent) (MessageResponsePayload, error) {
	eventType, ok := event["type"].(string)
	if !ok {
		return response, errors.New("invalid event type field type")
	}
	switch eventType {
	case "message_start":
		return handleMessageStartEvent(event, response)
	case "content_block_start":
		return handleContentBlockStartEvent(event, response)
	case "content_block_delta":
		return handleContentBlockDeltaEvent(ctx, event, response, payload)
	case "content_block_stop":
		// Nothing to do here
	case "message_delta":
		return handleMessageDeltaEvent(event, response)
	case "message_stop":
		eventChan <- MessageEvent{Response: &response, Err: nil}
	case "ping":
		// Nothing to do here
	default:
		log.Printf("unknown event type: %s", eventType)
	}
	return response, nil
}

func handleMessageStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	message, ok := event["message"].(map[string]interface{})
	if !ok {
		return response, errors.New("invalid message field type")
	}

	usage, ok := message["usage"].(map[string]interface{})
	if !ok {
		return response, errors.New("invalid usage field type")
	}

	inputTokens, err := getFloat64(usage, "input_tokens")
	if err != nil {
		return response, err
	}

	response.ID = getString(message, "id")
	response.Model = getString(message, "model")
	response.Role = getString(message, "role")
	response.Type = getString(message, "type")
	response.Usage.InputTokens = int(inputTokens)

	return response, nil
}

func handleContentBlockStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, errors.New("invalid index field type")
	}
	index := int(indexValue)

	if len(response.Content) <= index {
		response.Content = append(response.Content, struct {
			Text string `json:"text"`
			Type string `json:"type"`
		}{})
	}
	return response, nil
}

func handleContentBlockDeltaEvent(ctx context.Context, event map[string]interface{}, response MessageResponsePayload, payload *messagePayload) (MessageResponsePayload, error) {
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, errors.New("invalid index field type")
	}
	index := int(indexValue)

	delta, ok := event["delta"].(map[string]interface{})
	if !ok {
		return response, errors.New("invalid delta field type")
	}
	deltaType, ok := delta["type"].(string)
	if !ok {
		return response, errors.New("invalid delta type field type")
	}

	if deltaType == "text_delta" {
		text, ok := delta["text"].(string)
		if !ok {
			return response, errors.New("invalid delta text field type")
		}
		if len(response.Content) > index {
			response.Content[index].Text += text
		} else {
			return response, errors.New("content index out of range")
		}
	}

	if payload.StreamingFunc != nil {
		text, ok := delta["text"].(string)
		if !ok {
			return response, errors.New("invalid delta text field type")
		}
		err := payload.StreamingFunc(ctx, []byte(text))
		if err != nil {
			return response, fmt.Errorf("streaming func returned an error: %w", err)
		}
	}
	return response, nil
}

func handleMessageDeltaEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	delta, ok := event["delta"].(map[string]interface{})
	if !ok {
		return response, errors.New("invalid delta field type")
	}
	if stopReason, ok := delta["stop_reason"].(string); ok {
		response.StopReason = stopReason
	}

	usage, ok := event["usage"].(map[string]interface{})
	if !ok {
		return response, errors.New("invalid usage field type")
	}
	if outputTokens, ok := usage["output_tokens"].(float64); ok {
		response.Usage.OutputTokens = int(outputTokens)
	}
	return response, nil
}

func getString(m map[string]interface{}, key string) string {
	value, ok := m[key].(string)
	if !ok {
		return ""
	}
	return value
}

func getFloat64(m map[string]interface{}, key string) (float64, error) {
	value, ok := m[key].(float64)
	if !ok {
		return 0, errors.New("invalid field type")
	}
	return value, nil
}
