// extract the errors in the package to the top level:

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
	"time"

	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	// The Anthropic API enforces strict rate limits, especially for users on the initial (free or low-tier) plans.
	// As a result, HTTP 429 (Too Many Requests) errors are frequently encountered.
	maxSendingRetries = 10
	retryInterval     = 10 * time.Second
)

var (
	ErrInvalidEventType           = fmt.Errorf("invalid event type field type")
	ErrInvalidMessageField        = fmt.Errorf("invalid message field type")
	ErrInvalidUsageField          = fmt.Errorf("invalid usage field type")
	ErrInvalidIndexField          = fmt.Errorf("invalid index field type")
	ErrInvalidContentBlockField   = fmt.Errorf("invalid content block field type")
	ErrInvalidDeltaField          = fmt.Errorf("invalid delta field type")
	ErrInvalidDeltaTypeField      = fmt.Errorf("invalid delta type field type")
	ErrInvalidDeltaTextField      = fmt.Errorf("invalid delta text field type")
	ErrInvalidDeltaToolCallField  = fmt.Errorf("invalid delta tool call field type")
	ErrInvalidDeltaThinkingField  = fmt.Errorf("invalid delta thinking field type")
	ErrInvalidDeltaSignatureField = fmt.Errorf("invalid delta signature field type")
	ErrContentIndexOutOfRange     = fmt.Errorf("content index out of range")
	ErrFailedCastToTextContent    = fmt.Errorf("failed to cast content to TextContent")
	ErrInvalidFieldType           = fmt.Errorf("invalid field type")
)

// For correct using thinking events, use this guide:
// https://docs.anthropic.com/en/docs/build-with-claude/extended-thinking
// In this implementation, we don't send the thinking block to the server.
const (
	EventTypeText     = "text"
	EventTypeToolUse  = "tool_use"
	EventTypeThinking = "thinking"
)

const (
	EventDeltaTypeText           = "text_delta"
	EventDeltaTypeToolUse        = "input_json_delta"
	EventDeltaTypeThinking       = "thinking_delta"
	EventDeltaTypeSignatureDelta = "signature_delta"
)

type ChatMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type ThinkingPayload struct {
	Type   string `json:"type"`
	Budget int    `json:"budget_tokens,omitempty"`
}

type messagePayload struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	System      any           `json:"system,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	StopWords   []string      `json:"stop_sequences,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	Tools       []Tool        `json:"tools,omitempty"`
	ToolChoice  any           `json:"tool_choice,omitempty"`

	Thinking *ThinkingPayload `json:"thinking,omitempty"`

	StreamingFunc streaming.Callback `json:"-"`
}

// Tool used for the request message payload.
type Tool struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	InputSchema  any           `json:"input_schema,omitempty"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
}

// CacheControl represents Anthropic's prompt caching configuration.
type CacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"` // "5m" or "1h"
}

// Content can be TextContent or ToolUseContent depending on the type.
type Content interface {
	GetType() string
}

type TextContent struct {
	Type         string        `json:"type"`
	Text         string        `json:"text"`
	CacheControl *CacheControl `json:"cache_control,omitempty"`
	Signature    string        `json:"signature,omitempty"`
}

func (tc TextContent) GetType() string {
	return tc.Type
}

type PartialJSONContent struct {
	Type        string `json:"type"`
	PartialJSON string `json:"partial_json"`
}

func (tc PartialJSONContent) GetType() string {
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

type ThinkingContent struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
}

func (tc ThinkingContent) GetType() string {
	return tc.Type
}

type ToolUseContent struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Input        map[string]interface{} `json:"input"`
	CacheControl *CacheControl          `json:"cache_control,omitempty"`
	Signature    string                 `json:"signature,omitempty"`

	rawStreamInput string
}

func (tuc *ToolUseContent) AppendStreamChunk(chunk string) {
	tuc.rawStreamInput += chunk
}

func (tuc *ToolUseContent) GetStreamInput() string {
	return tuc.rawStreamInput
}

func (tuc *ToolUseContent) DecodeStream() error {
	if tuc.rawStreamInput == "" {
		return nil
	}

	err := json.Unmarshal([]byte(tuc.rawStreamInput), &tuc.Input)
	if err != nil {
		return err
	}

	return nil
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
		CacheCreation            struct {
			Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens,omitempty"`
			Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens,omitempty"`
		} `json:"cache_creation,omitempty"`
		ServiceTier string `json:"service_tier,omitempty"`
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
		content, err := parseContentBlock(raw)
		if err != nil {
			return err
		}
		m.Content = append(m.Content, content)
	}

	return nil
}

func parseContentBlock(raw []byte) (Content, error) {
	var typeStruct struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &typeStruct); err != nil {
		return nil, err
	}

	switch typeStruct.Type {
	case EventTypeText:
		tc := &TextContent{}
		if err := json.Unmarshal(raw, tc); err != nil {
			return nil, err
		}
		return tc, nil
	case EventTypeToolUse:
		tuc := &ToolUseContent{}
		if err := json.Unmarshal(raw, tuc); err != nil {
			return nil, err
		}
		return tuc, nil
	case EventTypeThinking:
		thc := &ThinkingContent{}
		if err := json.Unmarshal(raw, thc); err != nil {
			return nil, err
		}
		return thc, nil
	default:
		return nil, fmt.Errorf("unknown content type: %s\n%v", typeStruct.Type, string(raw)) //nolint:err113
	}
}

func (c *Client) setMessageDefaults(payload *messagePayload) {
	// Set defaults
	if payload.MaxTokens == nil || *payload.MaxTokens == 0 {
		payload.MaxTokens = getIntPointer(2048)
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

	var resp *http.Response
	for range maxSendingRetries {
		resp, err = c.doWithHeaders(ctx, "/messages", payloadBytes, betaHeaders)
		if err != nil {
			return nil, fmt.Errorf("send request: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close() // avoid memory leaks

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryInterval):
				continue
			}
		}

		break
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeError(resp)
	}

	if payload.Stream {
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

func parseStreamingMessageResponse(
	ctx context.Context,
	r *http.Response,
	payload *messagePayload,
) (*MessageResponsePayload, error) {
	scanner := bufio.NewScanner(r.Body)
	eventChan := make(chan MessageEvent)

	go func() {
		defer close(eventChan)
		defer streaming.CallWithDone(ctx, payload.StreamingFunc) //nolint:errcheck

		var response MessageResponsePayload
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				// it's happening when the server answer is not a streaming response
				// we need to parse the response as a normal response and return it
				if err := parseMessageResponse(ctx, line, payload, eventChan); err != nil {
					eventChan <- MessageEvent{
						Response: nil,
						Err:      fmt.Errorf("failed to parse stream message response: %w", err),
					}
					return
				}
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

func parseMessageResponse(ctx context.Context, line string,
	payload *messagePayload, eventChan chan<- MessageEvent,
) error {
	var response MessageResponsePayload
	if err := json.Unmarshal([]byte(line), &response); err != nil {
		// skip line if it's not a valid json as a message response
		// it's happening when custom server is used with the same API as anthropic
		return nil //nolint:nilerr
	}

	for _, content := range response.Content {
		switch cv := content.(type) {
		case *ThinkingContent:
			reasoning := &reasoning.ContentReasoning{Content: cv.Thinking}
			if err := streaming.CallWithReasoning(ctx, payload.StreamingFunc, reasoning); err != nil {
				return fmt.Errorf("streaming func returned an error: %w", err)
			}
		case *TextContent:
			if err := streaming.CallWithText(ctx, payload.StreamingFunc, cv.Text); err != nil {
				return fmt.Errorf("streaming func returned an error: %w", err)
			}
		case *ToolUseContent:
			toolArgs, err := json.Marshal(cv.Input)
			if err != nil {
				return fmt.Errorf("failed to marshal tool use input: %w", err)
			}
			toolCall := streaming.NewToolCall(cv.ID, cv.Name, string(toolArgs))
			if err := streaming.CallWithToolCall(ctx, payload.StreamingFunc, toolCall); err != nil {
				return fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
	}

	eventChan <- MessageEvent{Response: &response, Err: nil}
	return nil
}

func parseStreamEvent(data string) (map[string]interface{}, error) {
	var event map[string]interface{}
	return event, json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&event)
}

func processStreamEvent(ctx context.Context, event map[string]interface{}, payload *messagePayload,
	response MessageResponsePayload, eventChan chan<- MessageEvent,
) (MessageResponsePayload, error) {
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
		return handleContentBlockStopEvent(response)
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

func handleMessageStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) { //nolint:lll
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

func handleContentBlockStartEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) { //nolint:lll
	indexValue, ok := event["index"].(float64)
	if !ok {
		return response, ErrInvalidIndexField
	}
	index := int(indexValue)

	cb, ok := event["content_block"].(map[string]any)
	if !ok {
		return response, ErrInvalidContentBlockField
	}
	if getString(cb, "type") == "" {
		return response, fmt.Errorf("%w: content block type is empty", ErrInvalidDeltaField)
	}

	if len(response.Content) <= index {
		switch eventType := getString(cb, "type"); eventType {
		case EventTypeText:
			response.Content = append(response.Content, &TextContent{
				Type: eventType,
				Text: getString(cb, "text"),
			})
		case EventTypeToolUse:
			response.Content = append(response.Content, &ToolUseContent{
				Type:  eventType,
				ID:    getString(cb, "id"),
				Name:  getString(cb, "name"),
				Input: getMap(cb, "input"),
			})
		case EventTypeThinking:
			response.Content = append(response.Content, &ThinkingContent{
				Type:      eventType,
				Thinking:  getString(cb, "thinking"),
				Signature: getString(cb, "signature"),
			})
		default:
			return response, fmt.Errorf("unknown content block type: %s", eventType)
		}
	}

	return response, nil
}

func handleContentBlockDeltaEvent(ctx context.Context, event map[string]interface{},
	response MessageResponsePayload, payload *messagePayload,
) (MessageResponsePayload, error) {
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

	var err error

	switch deltaType {
	case EventDeltaTypeText:
		response, err = handleTextDelta(ctx, delta, response, index, payload)
	case EventDeltaTypeToolUse:
		response, err = handleInputJSONDelta(ctx, delta, response, index, payload)
	case EventDeltaTypeThinking:
		response, err = handleThinkingDelta(ctx, delta, response, index, payload)
	case EventDeltaTypeSignatureDelta:
		response, err = handleSignatureDelta(ctx, delta, response, index, payload)
	default:
		return response, fmt.Errorf("unknown delta type: %s", deltaType)
	}

	if err != nil {
		return response, err
	}

	return response, err
}

func handleTextDelta(ctx context.Context, delta map[string]interface{},
	response MessageResponsePayload, index int, payload *messagePayload,
) (MessageResponsePayload, error) {
	text, ok := delta["text"].(string)
	if !ok {
		return response, ErrInvalidDeltaTextField
	}

	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}

	textContent, ok := response.Content[index].(*TextContent)
	if !ok {
		return response, ErrFailedCastToTextContent
	}

	textContent.Text += text

	return response, streaming.CallWithText(ctx, payload.StreamingFunc, text)
}

func handleInputJSONDelta(ctx context.Context, delta map[string]interface{},
	response MessageResponsePayload, index int, payload *messagePayload,
) (MessageResponsePayload, error) {
	partial, ok := delta["partial_json"].(string)
	if !ok {
		return response, ErrInvalidDeltaToolCallField
	}

	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}

	tuc, ok := response.Content[index].(*ToolUseContent)
	if !ok {
		asJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return response, fmt.Errorf("failed to marshal response: %w", err)
		}
		msg := fmt.Sprintf("failed to cast index %d to ToolUseContent: \n%s", index, string(asJSON))
		return response, errors.New(msg) //nolint:err113
	}

	tuc.AppendStreamChunk(partial)
	toolCall := streaming.NewToolCall(tuc.ID, tuc.Name, partial)
	return response, streaming.CallWithToolCall(ctx, payload.StreamingFunc, toolCall)
}

func handleThinkingDelta(ctx context.Context, delta map[string]interface{},
	response MessageResponsePayload, index int, payload *messagePayload,
) (MessageResponsePayload, error) {
	thinking, ok := delta["thinking"].(string)
	if !ok {
		return response, ErrInvalidDeltaThinkingField
	}

	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}

	thinkingContent, ok := response.Content[index].(*ThinkingContent)
	if !ok {
		return response, ErrFailedCastToTextContent
	}

	thinkingContent.Thinking += thinking
	if signature, ok := delta["signature"].(string); ok && signature != "" {
		thinkingContent.Signature = signature
	}

	// Include signature in reasoning chunk if available
	var signature []byte
	if len(thinkingContent.Signature) > 0 {
		signature = []byte(thinkingContent.Signature)
	}

	reasoningChunk := &reasoning.ContentReasoning{
		Content:   thinking,
		Signature: signature, // Include accumulated signature
	}
	return response, streaming.CallWithReasoning(ctx, payload.StreamingFunc, reasoningChunk)
}

func handleSignatureDelta(_ context.Context, delta map[string]interface{},
	response MessageResponsePayload, index int, _ *messagePayload,
) (MessageResponsePayload, error) {
	signature, ok := delta["signature"].(string)
	if !ok {
		return response, ErrInvalidDeltaSignatureField
	}

	if len(response.Content) <= index {
		return response, ErrContentIndexOutOfRange
	}

	thinkingContent, ok := response.Content[index].(*ThinkingContent)
	if !ok {
		return response, ErrFailedCastToTextContent
	}

	thinkingContent.Signature += signature

	return response, nil // no need to inform about this delta event
}

func handleContentBlockStopEvent(response MessageResponsePayload) (MessageResponsePayload, error) {
	for _, content := range response.Content {
		if content == nil {
			continue
		}
		tuc, ok := content.(*ToolUseContent)
		if !ok {
			continue
		}

		err := tuc.DecodeStream()
		if err != nil {
			return response, fmt.Errorf("error decoding stream tool data: %w", err)
		}
	}

	return response, nil
}

func handleMessageDeltaEvent(event map[string]interface{}, response MessageResponsePayload) (MessageResponsePayload, error) { //nolint:lll
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

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	value, ok := m[key].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	return value
}

func getIntPointer(i int) *int {
	return &i
}
