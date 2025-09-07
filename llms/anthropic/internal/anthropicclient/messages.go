// extract the errors in the package to the top level:

package anthropicclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	ErrInvalidEventType             = fmt.Errorf("invalid event type field type")
	ErrInvalidMessageField          = fmt.Errorf("invalid message field type")
	ErrInvalidUsageField            = fmt.Errorf("invalid usage field type")
	ErrInvalidIndexField            = fmt.Errorf("invalid index field type")
	ErrInvalidDeltaField            = fmt.Errorf("invalid delta field type")
	ErrInvalidDeltaTypeField        = fmt.Errorf("invalid delta type field type")
	ErrInvalidDeltaTextField        = fmt.Errorf("invalid delta text field type")
	ErrInvalidDeltaPartialJSONField = fmt.Errorf("invalid delta partial_json field type")
	ErrContentIndexOutOfRange       = fmt.Errorf("content index out of range")
	ErrFailedCastToTextContent      = fmt.Errorf("failed to cast content to TextContent")
	ErrFailedCastToToolUseContent   = fmt.Errorf("failed to cast content to ToolUseContent")
	ErrInvalidFieldType             = fmt.Errorf("invalid field type")
)

type ChatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type messagePayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	System      string        `json:"system,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	StopWords   []string      `json:"stop_sequences,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature float64       `json:"temperature"`
	Tools       []Tool        `json:"tools,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`

	// Extended thinking parameters (Claude 3.7+)
	Thinking *ThinkingConfig `json:"thinking,omitempty"`

	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
}

// ThinkingConfig represents the thinking configuration for Claude 3.7+
type ThinkingConfig struct {
	Type         string `json:"type"` // "enabled" or "disabled"
	BudgetTokens int    `json:"budget_tokens,omitempty"`
}

// Tool used for the request message payload.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema any    `json:"input_schema,omitempty"`
}

// CacheControl represents Anthropic's prompt caching configuration.
type CacheControl struct {
	Type string `json:"type"`
}

// Content can be TextContent or ToolUseContent depending on the type.
type Content interface {
	GetType() string
}

type TextContent struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

func (tc TextContent) GetType() string {
	return tc.Type
}

type ImageContent struct {
	Type         string        `json:"type"`
	Source       ImageSource   `json:"source"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

func (ic ImageContent) GetType() string {
	return ic.Type
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ToolUseContent struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Input        map[string]interface{} `json:"input"`
	CacheControl *CacheControl          `json:"cache_control,omitempty"`

	inputData string `json:"-"` // Used to gather input data when streaming
}

func (tuc ToolUseContent) GetType() string {
	return tuc.Type
}

type ToolResultContent struct {
	Type         string        `json:"type"`
	ToolUseID    string        `json:"tool_use_id"`
	Content      string        `json:"content"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

func (trc ToolResultContent) GetType() string {
	return trc.Type
}

// ThinkingContent represents Claude's thinking/reasoning content
type ThinkingContent struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

func (tc ThinkingContent) GetType() string {
	return tc.Type
}

type MessageResponsePayload struct {
	Content      []Content `json:"content"`
	ID           string    `json:"id"`
	Model        string    `json:"model"`
	Role         string    `json:"role"`
	StopReason   string    `json:"stop_reason"`
	StopSequence string    `json:"stop_sequence"`
	Type         string    `json:"type"`
	Usage        struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	} `json:"usage"`
}

func (m *MessageResponsePayload) UnmarshalJSON(data []byte) error {
	type Alias MessageResponsePayload
	aux := &struct {
		Content []json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	for _, raw := range aux.Content {
		var typeStruct struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &typeStruct); err != nil {
			return err
		}

		switch typeStruct.Type {
		case "text":
			tc := &TextContent{}
			if err := json.Unmarshal(raw, tc); err != nil {
				return err
			}
			m.Content = append(m.Content, tc)
		case "tool_use":
			tuc := &ToolUseContent{}
			if err := json.Unmarshal(raw, tuc); err != nil {
				return err
			}
			m.Content = append(m.Content, tuc)
		case "thinking":
			tc := &ThinkingContent{}
			if err := json.Unmarshal(raw, tc); err != nil {
				return err
			}
			m.Content = append(m.Content, tc)
		default:
			return fmt.Errorf("unknown content type: %s\n%v", typeStruct.Type, string(raw))
		}
	}

	return nil
}

func (c *Client) setMessageDefaults(payload *messagePayload) {
	// Set defaults
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 2048
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

func (c *Client) createMessage(ctx context.Context, payload *messagePayload, betaHeaders []string) (*MessageResponsePayload, error) {
	c.setMessageDefaults(payload)

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	resp, err := c.doWithHeaders(ctx, "/messages", payloadBytes, betaHeaders)
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
		return response, ErrInvalidEventType
	}
	switch eventType {
	case "message_start":
		return handleMessageStartEvent(event, response)
	case "content_block_start":
		return handleContentBlockStartEvent(event, response)
	case "content_block_delta":
		return handleContentBlockDeltaEvent(ctx, event, response, payload)
	case "content_block_stop":
		return handleContentBlockStop(event, response)
	case "message_delta":
		return handleMessageDeltaEvent(event, response)
	case "message_stop":
		eventChan <- MessageEvent{Response: &response, Err: nil}
	case "ping":
		// Nothing to do here
	case "error":
		eventChan <- MessageEvent{Response: nil, Err: fmt.Errorf("received error event: %v", event)}
	default:
		log.Printf("unknown event type: %s - %v", eventType, event)
	}
	return response, nil
}

func handleMessageStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	message, ok := event["message"].(map[string]interface{})
	if !ok {
		return response, ErrInvalidMessageField
	}

	usage, ok := message["usage"].(map[string]interface{})
	if !ok {
		return response, ErrInvalidUsageField
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

	// Capture cache token information if present
	if cacheCreationTokens, err := getFloat64(usage, "cache_creation_input_tokens"); err == nil {
		response.Usage.CacheCreationInputTokens = int(cacheCreationTokens)
	}
	if cacheReadTokens, err := getFloat64(usage, "cache_read_input_tokens"); err == nil {
		response.Usage.CacheReadInputTokens = int(cacheReadTokens)
	}

	return response, nil
}

func handleContentBlockStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, ErrInvalidIndexField
	}
	index := int(indexValue)

	var eventType string
	var contentBlock map[string]any
	if cb, ok := event["content_block"].(map[string]any); ok {
		contentBlock = cb
		typ, _ := cb["type"].(string)
		eventType = typ
	}
	if eventType == "" {
		return response, fmt.Errorf("%w: content block type is empty", ErrInvalidDeltaField)
	}

	if len(response.Content) <= index {
		switch eventType {
		case "text":
			response.Content = append(response.Content, &TextContent{
				Type: eventType,
			})
		case "tool_use":
			input, ok := event["input"].(map[string]interface{})
			if !ok {
				// If the input is not provided, it may be coming in a future event.
				input = make(map[string]interface{})
			}

			response.Content = append(response.Content, &ToolUseContent{
				Type:  eventType,
				ID:    getString(contentBlock, "id"),
				Name:  getString(contentBlock, "name"),
				Input: input,
			})
		case "thinking":
			response.Content = append(response.Content, &ThinkingContent{
				Type: eventType,
			})
		default:
			return response, fmt.Errorf("%w: unknown content block type: %s", ErrInvalidDeltaField, eventType)
		}
	}
	return response, nil
}

// handleContentBlockDeltaEvent processes delta events for content blocks, handling both text and JSON deltas.
func handleContentBlockDeltaEvent(ctx context.Context, event map[string]interface{}, response MessageResponsePayload, payload *messagePayload) (MessageResponsePayload, error) {
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, ErrInvalidIndexField
	}
	index := int(indexValue)

	delta, ok := event["delta"].(map[string]interface{})
	if !ok {
		return response, ErrInvalidDeltaField
	}
	deltaType, ok := delta["type"].(string)
	if !ok {
		return response, ErrInvalidDeltaTypeField
	}

	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}

	switch deltaType {
	case "text_delta":
		return handleTextDelta(ctx, delta, response, payload, index)
	case "input_json_delta":
		return handleJSONDelta(delta, response, index)
	case "thinking_delta":
		return handleThinkingDelta(delta, response, index)
	}

	return response, nil
}

// handleTextDelta processes text delta events for content blocks.
func handleTextDelta(ctx context.Context, delta map[string]interface{}, response MessageResponsePayload, payload *messagePayload, index int) (MessageResponsePayload, error) {
	text, ok := delta["text"].(string)
	if !ok {
		return response, ErrInvalidDeltaTextField
	}
	textContent, ok := response.Content[index].(*TextContent)
	if !ok {
		return response, ErrFailedCastToTextContent
	}
	textContent.Text += text

	// Streaming functions only work with text deltas.
	if payload.StreamingFunc != nil {
		err := payload.StreamingFunc(ctx, []byte(text))
		if err != nil {
			return response, fmt.Errorf("streaming func returned an error: %w", err)
		}
	}

	return response, nil
}

// handleJSONDelta processes JSON delta events for content blocks.
func handleJSONDelta(delta map[string]interface{}, response MessageResponsePayload, index int) (MessageResponsePayload, error) {
	partialJSON, ok := delta["partial_json"].(string)
	if !ok {
		return response, ErrInvalidDeltaPartialJSONField
	}
	toolUseContent, ok := response.Content[index].(*ToolUseContent)
	if !ok {
		return response, ErrFailedCastToToolUseContent
	}
	toolUseContent.inputData += partialJSON

	return response, nil
}

// handleThinkingDelta processes thinking delta events for content blocks.
func handleThinkingDelta(delta map[string]interface{}, response MessageResponsePayload, index int) (MessageResponsePayload, error) {
	thinking, ok := delta["thinking"].(string)
	if !ok {
		return response, ErrInvalidDeltaTextField
	}
	thinkingContent, ok := response.Content[index].(*ThinkingContent)
	if !ok {
		return response, fmt.Errorf("failed to cast to ThinkingContent at index %d", index)
	}
	thinkingContent.Thinking += thinking
	return response, nil
}

func handleContentBlockStop(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, ErrInvalidIndexField
	}

	index := int(indexValue)
	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}
	if toolUseContent, ok := response.Content[index].(*ToolUseContent); ok {
		toolUseContent.inputData = strings.TrimSpace(toolUseContent.inputData)
		if toolUseContent.inputData != "" {
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(toolUseContent.inputData), &input); err != nil {
				return response, fmt.Errorf("failed to unmarshal input data: %w", err)
			}
			toolUseContent.Input = input
		}
	}

	return response, nil
}

func handleMessageDeltaEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) {
	delta, ok := event["delta"].(map[string]interface{})
	if !ok {
		return response, ErrInvalidDeltaField
	}
	if stopReason, ok := delta["stop_reason"].(string); ok {
		response.StopReason = stopReason
	}

	usage, ok := event["usage"].(map[string]interface{})
	if !ok {
		return response, ErrInvalidUsageField
	}
	if outputTokens, ok := usage["output_tokens"].(float64); ok {
		response.Usage.OutputTokens = int(outputTokens)
	}
	// Also capture cache tokens in the final message_delta event
	if inputTokens, err := getFloat64(usage, "input_tokens"); err == nil {
		response.Usage.InputTokens = int(inputTokens)
	}
	if cacheCreationTokens, err := getFloat64(usage, "cache_creation_input_tokens"); err == nil {
		response.Usage.CacheCreationInputTokens = int(cacheCreationTokens)
	}
	if cacheReadTokens, err := getFloat64(usage, "cache_read_input_tokens"); err == nil {
		response.Usage.CacheReadInputTokens = int(cacheReadTokens)
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
		return 0, ErrInvalidFieldType
	}
	return value, nil
}
