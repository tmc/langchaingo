package openaiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/reasoning"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

const (
	defaultChatModel = "gpt-4.1-mini"
)

var ErrContentExclusive = errors.New("only one of Content / MultiContent allowed in message")

type StreamOptions struct {
	// If set, an additional chunk will be streamed before the data: [DONE] message.
	// The usage field on this chunk shows the token usage statistics for the entire request,
	// and the choices field will always be an empty array.
	// All other chunks will also include a usage field, but with a null value.
	IncludeUsage bool `json:"include_usage,omitempty"`
}

// ReasoningOptions is enabling reasoning if the model supports it.
// There should have to use one of the fields: effort or max_tokens.
type ReasoningOptions struct {
	Effort    llms.ReasoningEffort `json:"effort,omitempty"`
	MaxTokens int                  `json:"max_tokens,omitempty"`
}

// ChatRequest is a request to complete a chat completion..
type ChatRequest struct {
	Model               string         `json:"model"`
	Messages            []*ChatMessage `json:"messages"`
	Temperature         *float64       `json:"temperature,omitempty"`
	TopK                *int           `json:"top_k,omitempty"`
	TopP                *float64       `json:"top_p,omitempty"`
	MinP                *float64       `json:"min_p,omitempty"`
	MaxTokens           *int           `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int           `json:"max_completion_tokens,omitempty"`
	N                   *int           `json:"n,omitempty"`
	StopWords           []string       `json:"stop,omitempty"`
	Stream              bool           `json:"stream,omitempty"`
	FrequencyPenalty    *float64       `json:"frequency_penalty,omitempty"`
	PresencePenalty     *float64       `json:"presence_penalty,omitempty"`
	RepetitionPenalty   *float64       `json:"repetition_penalty,omitempty"`
	Seed                *int           `json:"seed,omitempty"`

	// ReasoningEffort enables reasoning mode for models that support it.
	// Set this field when you want to use the legacy reasoning configuration.
	// Do not use ReasoningEffort together with Reasoning; only one should be set at a time.
	ReasoningEffort *llms.ReasoningEffort `json:"reasoning_effort,omitempty"`

	// Reasoning provides advanced reasoning configuration for models that support it.
	// Use either the Effort or MaxTokens field to control reasoning behavior.
	// This field should be set when using the modern reasoning format.
	// Do not set both Reasoning and ReasoningEffort at the same time, as they are mutually exclusive.
	Reasoning *ReasoningOptions `json:"reasoning,omitempty"`

	// ResponseFormat is the format of the response.
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`

	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int `json:"top_logprobs,omitempty"`

	Tools []Tool `json:"tools,omitempty"`
	// This can be either a string or a ToolChoice object.
	// If it is a string, it should be one of 'none', or 'auto', otherwise it should be a ToolChoice object specifying a specific tool to use.
	ToolChoice any `json:"tool_choice,omitempty"`

	// Options for streaming response. Only set this when you set stream: true.
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc streaming.Callback `json:"-"`

	// Deprecated: use Tools instead.
	Functions []FunctionDefinition `json:"functions,omitempty"`
	// Deprecated: use ToolChoice instead.
	FunctionCallBehavior FunctionCallBehavior `json:"function_call,omitempty"`

	// Metadata allows you to specify additional information that will be passed to the model.
	Metadata map[string]any `json:"metadata,omitempty"`

	// WebSearchOptions configures web search behavior for search-enabled models
	// like gpt-4o-search-preview and gpt-4o-mini-search-preview.
	WebSearchOptions *WebSearchOptions `json:"web_search_options,omitempty"`

	// ExtraBody allows passing additional fields that will be merged into the request body.
	// These fields take precedence over the standard fields.
	ExtraBody map[string]any `json:"-"`
}

// ToolType is the type of a tool.
type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

// WebSearchOptions configures web search behavior for OpenAI models.
// This is used with search-enabled models like gpt-4o-search-preview.
type WebSearchOptions struct {
	// SearchContextSize controls how much context is gathered from web search.
	// Valid values: "low", "medium", "high". Higher values provide more context
	// but increase latency and cost.
	SearchContextSize string `json:"search_context_size,omitempty"`

	// UserLocation provides approximate user location for localized search results.
	UserLocation *UserLocation `json:"user_location,omitempty"`
}

// UserLocation represents the user's approximate location for web search.
type UserLocation struct {
	// Type must be "approximate" for user-provided location.
	Type string `json:"type"`

	// Approximate contains the approximate location details.
	Approximate *ApproximateLocation `json:"approximate,omitempty"`
}

// ApproximateLocation contains approximate location information.
type ApproximateLocation struct {
	// Country is the two-letter ISO country code (e.g., "US", "GB").
	Country string `json:"country,omitempty"`

	// City is the city name (e.g., "San Francisco", "London").
	City string `json:"city,omitempty"`

	// Region is the region or state (e.g., "California", "London").
	Region string `json:"region,omitempty"`
}

// Tool is a tool to use in a chat request.
type Tool struct {
	Type     ToolType           `json:"type"`
	Function FunctionDefinition `json:"function,omitempty"`
}

// ToolChoice is a choice of a tool to use.
type ToolChoice struct {
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

// ToolFunction is a function to be called in a tool choice.
type ToolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCall is a call to a tool.
type ToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

type ResponseFormatJSONSchemaProperty struct {
	Type                 string                                       `json:"type"`
	Description          string                                       `json:"description,omitempty"`
	Enum                 []interface{}                                `json:"enum,omitempty"`
	Items                *ResponseFormatJSONSchemaProperty            `json:"items,omitempty"`
	Properties           map[string]*ResponseFormatJSONSchemaProperty `json:"properties,omitempty"`
	AdditionalProperties bool                                         `json:"additionalProperties"`
	Required             []string                                     `json:"required,omitempty"`
	Ref                  string                                       `json:"$ref,omitempty"`
}

type ResponseFormatJSONSchema struct {
	Name   string                            `json:"name"`
	Strict bool                              `json:"strict"`
	Schema *ResponseFormatJSONSchemaProperty `json:"schema"`
}

// ResponseFormat is the format of the response.
type ResponseFormat struct {
	Type       string                    `json:"type"`
	JSONSchema *ResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

// ChatMessage is a message in a chat request.
type ChatMessage struct { //nolint:musttag
	// The role of the author of this message. One of system, user, assistant, function, or tool.
	Role string

	// The content of the message.
	// This field is mutually exclusive with MultiContent.
	Content string

	// MultiContent is a list of content parts to use in the message.
	MultiContent []llms.ContentPart

	// The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,
	// with a maximum length of 64 characters.
	Name string

	// ToolCalls is a list of tools that were called in the message.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// FunctionCall represents a function call that was made in the message.
	// Deprecated: use ToolCalls instead.
	FunctionCall *FunctionCall

	// ToolCallID is the ID of the tool call this message is for.
	// Only present in tool messages.
	ToolCallID string `json:"tool_call_id,omitempty"`

	// This field is primarily used by reasoning-capable models. It contains
	// the assistant's step-by-step reasoning or thought process, provided before the final answer.
	ReasoningContent string `json:"reasoning_content,omitempty"`

	// This field serves as a fallback for ReasoningContent. If ReasoningContent is empty,
	// Reasoning may contain the assistant's reasoning or explanation.
	Reasoning string `json:"reasoning,omitempty"`
}

func (m ChatMessage) MarshalJSON() ([]byte, error) {
	if m.Content != "" && m.MultiContent != nil {
		return nil, ErrContentExclusive
	}
	if text, ok := isSingleTextContent(m.MultiContent); ok {
		m.Content = text
		m.MultiContent = nil
	}
	if len(m.MultiContent) > 0 {
		msg := struct {
			Role         string             `json:"role"`
			Content      string             `json:"-"`
			MultiContent []llms.ContentPart `json:"content,omitempty"`
			Name         string             `json:"name,omitempty"`
			ToolCalls    []ToolCall         `json:"tool_calls,omitempty"`

			// Deprecated: use ToolCalls instead.
			FunctionCall *FunctionCall `json:"function_call,omitempty"`

			// ToolCallID is the ID of the tool call this message is for.
			// Only present in tool messages.
			ToolCallID string `json:"tool_call_id,omitempty"`

			// Reasoning content result fields
			ReasoningContent string `json:"reasoning_content,omitempty"`
			Reasoning        string `json:"reasoning,omitempty"`
		}(m)
		if msg.ReasoningContent == "" && msg.Reasoning != "" {
			msg.ReasoningContent = msg.Reasoning
		}
		return json.Marshal(msg)
	}
	msg := struct {
		Role         string             `json:"role"`
		Content      string             `json:"content"`
		MultiContent []llms.ContentPart `json:"-"`
		Name         string             `json:"name,omitempty"`
		ToolCalls    []ToolCall         `json:"tool_calls,omitempty"`
		// Deprecated: use ToolCalls instead.
		FunctionCall *FunctionCall `json:"function_call,omitempty"`

		// ToolCallID is the ID of the tool call this message is for.
		// Only present in tool messages.
		ToolCallID string `json:"tool_call_id,omitempty"`

		// Reasoning content result fields
		ReasoningContent string `json:"reasoning_content,omitempty"`
		Reasoning        string `json:"reasoning,omitempty"`
	}(m)
	if msg.ReasoningContent == "" && msg.Reasoning != "" {
		msg.ReasoningContent = msg.Reasoning
	}
	return json.Marshal(msg)
}

func isSingleTextContent(parts []llms.ContentPart) (string, bool) {
	if len(parts) != 1 {
		return "", false
	}
	tc, isText := parts[0].(llms.TextContent)
	return tc.Text, isText
}

func (m *ChatMessage) UnmarshalJSON(data []byte) error {
	msg := struct {
		Role         string             `json:"role"`
		Content      string             `json:"content"`
		MultiContent []llms.ContentPart `json:"-"` // not expected in response
		Name         string             `json:"name,omitempty"`
		ToolCalls    []ToolCall         `json:"tool_calls,omitempty"`
		// Deprecated: use ToolCalls instead.
		FunctionCall *FunctionCall `json:"function_call,omitempty"`

		// ToolCallID is the ID of the tool call this message is for.
		// Only present in tool messages.
		ToolCallID string `json:"tool_call_id,omitempty"`

		// Reasoning content result fields
		ReasoningContent string `json:"reasoning_content,omitempty"`
		Reasoning        string `json:"reasoning,omitempty"`
	}{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}
	if msg.ReasoningContent == "" && msg.Reasoning != "" {
		msg.ReasoningContent = msg.Reasoning
	}
	*m = ChatMessage(msg)
	return nil
}

type TopLogProbs struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

// LogProb represents the probability information for a token.
type LogProb struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"` // Omitting the field if it is null
	// TopLogProbs is a list of the most likely tokens and their log probability, at this token position.
	// In rare cases, there may be fewer than the number of requested top_logprobs returned.
	TopLogProbs []TopLogProbs `json:"top_logprobs"`
}

// LogProbs is the top-level structure containing the log probability information.
type LogProbs struct {
	// Content is a list of message content tokens with log probability information.
	Content []LogProb `json:"content"`
}

type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonFunctionCall  FinishReason = "function_call"
	FinishReasonToolCalls     FinishReason = "tool_calls"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonNull          FinishReason = "null"
)

func (r FinishReason) MarshalJSON() ([]byte, error) {
	if r == FinishReasonNull || r == "" {
		return []byte("null"), nil
	}
	return []byte(`"` + string(r) + `"`), nil // best effort to not break future API changes
}

// ChatCompletionChoice is a choice in a chat response.
type ChatCompletionChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	FinishReason FinishReason `json:"finish_reason"`
	LogProbs     *LogProbs    `json:"logprobs,omitempty"`
}

// ChatUsage is the usage of a chat completion request.
type ChatUsage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens     int `json:"cached_tokens"`
		CacheWriteTokens int `json:"cache_write_tokens,omitempty"`
		AudioTokens      int `json:"audio_tokens,omitempty"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
	CostDetails struct {
		UpstreamInferencePromptCost      *float64 `json:"upstream_inference_prompt_cost,omitempty"`
		UpstreamInferenceCompletionsCost *float64 `json:"upstream_inference_completions_cost,omitempty"`
	} `json:"cost_details,omitempty"`
}

// ChatCompletionResponse is a response to a chat request.
type ChatCompletionResponse struct {
	ID                string                  `json:"id,omitempty"`
	Created           int64                   `json:"created,omitempty"`
	Choices           []*ChatCompletionChoice `json:"choices,omitempty"`
	Model             string                  `json:"model,omitempty"`
	Object            string                  `json:"object,omitempty"`
	Usage             ChatUsage               `json:"usage,omitempty"`
	SystemFingerprint string                  `json:"system_fingerprint"`
}

type Usage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
		AudioTokens  int `json:"audio_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionTokensDetails struct {
		ReasoningTokens          int `json:"reasoning_tokens"`
		AudioTokens              int `json:"audio_tokens"`
		AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
		RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	} `json:"completion_tokens_details"`
}

// StreamedToolCall is a call to a tool.
type StreamedToolCall struct {
	Index    *int         `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     ToolType     `json:"type"`
	Function ToolFunction `json:"function,omitempty"`
}

type StreamedChatResponseChunkDelta struct {
	Role         string        `json:"role,omitempty"`
	Content      string        `json:"content,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"`
	// ToolCalls is a list of tools that were called in the message.
	ToolCalls []*StreamedToolCall `json:"tool_calls,omitempty"`
	// This field is only used with the deepseek-reasoner model and represents the reasoning contents of the assistant message before the final answer.
	ReasoningContent string `json:"reasoning_content,omitempty"`
	// Fallback field for reasoning content (it depends on the model and the provider)
	Reasoning string `json:"reasoning,omitempty"`
}

// StreamedChatResponseChunk is a chunk from the stream.
type StreamedChatResponseChunk struct {
	Index        int                             `json:"index"`
	FinishReason FinishReason                    `json:"finish_reason,omitempty"`
	Delta        *StreamedChatResponseChunkDelta `json:"delta,omitempty"`
}

// StreamedChatResponsePayload is a SSE paylaod from the stream.
type StreamedChatResponsePayload struct {
	ID                string                      `json:"id,omitempty"`
	Created           float64                     `json:"created,omitempty"`
	Model             string                      `json:"model,omitempty"`
	Object            string                      `json:"object,omitempty"`
	Choices           []StreamedChatResponseChunk `json:"choices,omitempty"`
	SystemFingerprint string                      `json:"system_fingerprint"`
	// An optional field that will only be present when you set stream_options: {"include_usage": true} in your request.
	// When present, it contains a null value except for the last chunk which contains the token usage statistics
	// for the entire request.
	Usage *Usage `json:"usage,omitempty"`
	Error error  `json:"-"` // use for error handling only
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description,omitempty"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters"`
	// Strict is a flag to enable structured output mode.
	Strict bool `json:"strict,omitempty"`
}

// FunctionCallBehavior is the behavior to use when calling functions.
type FunctionCallBehavior string

const (
	// FunctionCallBehaviorUnspecified is the empty string.
	FunctionCallBehaviorUnspecified FunctionCallBehavior = ""
	// FunctionCallBehaviorNone will not call any functions.
	FunctionCallBehaviorNone FunctionCallBehavior = "none"
	// FunctionCallBehaviorAuto will call functions automatically.
	FunctionCallBehaviorAuto FunctionCallBehavior = "auto"
)

// FunctionCall is a call to a function.
type FunctionCall struct {
	// Name is the name of the function to call.
	Name string `json:"name"`
	// Arguments is the set of arguments to pass to the function.
	Arguments string `json:"arguments"`
}

func (c *Client) createChat(ctx context.Context, payload *ChatRequest) (*ChatCompletionResponse, error) {
	if payload.StreamingFunc != nil {
		payload.Stream = true
		if payload.StreamOptions == nil {
			payload.StreamOptions = &StreamOptions{IncludeUsage: true}
		}
	}

	// Filter out internal metadata that shouldn't be sent to the API
	originalMetadata := payload.Metadata
	if payload.Metadata != nil {
		filteredMetadata := make(map[string]any)
		for k, v := range payload.Metadata {
			// Skip internal openai: prefixed metadata fields
			if !strings.HasPrefix(k, "openai:") {
				filteredMetadata[k] = v
			}
		}
		if len(filteredMetadata) > 0 {
			payload.Metadata = filteredMetadata
		} else {
			payload.Metadata = nil
		}
	}

	payloadBytes, err := json.Marshal(payload)

	// Restore original metadata
	payload.Metadata = originalMetadata

	if err != nil {
		return nil, err
	}

	// If ExtraBody is provided, merge it with the standard payload
	if len(payload.ExtraBody) > 0 {
		var baseMap map[string]any
		if err := json.Unmarshal(payloadBytes, &baseMap); err != nil {
			return nil, err
		}

		// Merge ExtraBody with priority (ExtraBody overwrites existing fields)
		maps.Copy(baseMap, payload.ExtraBody)

		if payloadBytes, err = json.Marshal(baseMap); err != nil {
			return nil, err
		}
	}

	// Build request
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL("/chat/completions", payload.Model), body)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	// Send request
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, sanitizeHTTPError(err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg)
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message)
	}
	if payload.Stream {
		return parseStreamingChatResponse(ctx, r, payload)
	}

	return parseChatResponse(r.Body)
}

func parseChatResponse(body io.Reader) (*ChatCompletionResponse, error) {
	var response ChatCompletionResponse
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Try to restore reasoning content (some model providers don't return reasoning content)
	for _, choice := range response.Choices {
		if choice.Message.ReasoningContent == "" {
			choice.Message.ReasoningContent = choice.Message.Reasoning
		}
		if choice.Message.ReasoningContent == "" {
			choice.Message.ReasoningContent, choice.Message.Content = reasoning.SplitContent(choice.Message.Content)
		}
	}

	return &response, nil
}

func parseStreamingChatResponse(
	ctx context.Context,
	r *http.Response,
	payload *ChatRequest,
) (*ChatCompletionResponse, error) {
	// Parse response
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan StreamedChatResponsePayload)

	go func() {
		defer close(responseChan)
		for scanner.Scan() {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if line == "" {
				continue
			}

			// Skip SSE comment lines (any line starting with ':')
			// According to SSE spec: https://www.w3.org/TR/eventsource/
			// "Lines that start with a U+003A COLON character (:) are comments"
			if strings.HasPrefix(line, ":") {
				continue
			}

			// Only process lines that start with "data:"
			if !strings.HasPrefix(line, "data:") {
				// Skip any other non-data lines (like event:, id:, retry:, etc.)
				continue
			}

			data := strings.TrimPrefix(line, "data:") // here use `data:` instead of `data: ` for compatibility
			data = strings.TrimSpace(data)
			if data == "[DONE]" {
				return
			}
			if !isValidJSON(data) {
				continue
			}

			var streamPayload StreamedChatResponsePayload
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				// Skip non-JSON data values that some providers might send
				// This could happen if the data field contains non-JSON content
				continue
			}

			// Non-blocking send with context check
			select {
			case <-ctx.Done():
				return
			case responseChan <- streamPayload:
			}
		}
		if err := scanner.Err(); err != nil {
			select {
			case <-ctx.Done():
				return
			case responseChan <- StreamedChatResponsePayload{Error: fmt.Errorf("error reading streaming response: %w", err)}:
			}
			return
		}
	}()

	// Combine response
	return combineStreamingChatResponse(ctx, payload, responseChan)
}

func isValidJSON(data string) bool {
	var dummy any
	data = strings.Trim(data, " \n\r\t")
	if !strings.HasPrefix(data, "{") || !strings.HasSuffix(data, "}") {
		return false
	}
	return json.Unmarshal([]byte(data), &dummy) == nil
}

//nolint:gocognit,cyclop
func combineStreamingChatResponse(
	ctx context.Context,
	payload *ChatRequest,
	responseChan chan StreamedChatResponsePayload,
) (*ChatCompletionResponse, error) {
	defer streaming.CallWithDone(ctx, payload.StreamingFunc) //nolint:errcheck

	var (
		response          ChatCompletionResponse
		splitters         []reasoning.ChunkContentSplitter
		toolCallNameCache = make(map[string]string) // Cache tool call names by ID for streaming
	)

	for streamResponse := range responseChan {
		if streamResponse.Error != nil {
			return nil, streamResponse.Error
		}

		updateChatUsage(&response.Usage, streamResponse.Usage)

		if len(streamResponse.Choices) == 0 {
			continue
		}

		for _, choice := range streamResponse.Choices {
			// Grow response.Choices slice to the length of the streamResponse.Choices
			for idx := range choice.Index + 1 {
				if len(response.Choices) <= idx {
					response.Choices = append(response.Choices, &ChatCompletionChoice{})
					splitters = append(splitters, reasoning.NewChunkContentSplitter())
				}
			}
			// Get current updatable values
			splitter := splitters[choice.Index]
			responseChoice := response.Choices[choice.Index]

			if choice.FinishReason != "" { // Update to last non-empty finish reason
				responseChoice.FinishReason = choice.FinishReason
			}
			if choice.Delta == nil { // Unexpected case, skip
				continue
			}

			content, reasoningContent := getChunkContent(choice, splitter)
			responseChoice.Message.Content += content
			responseChoice.Message.ReasoningContent += reasoningContent

			reasoning := &reasoning.ContentReasoning{Content: reasoningContent}
			if err := streaming.CallWithReasoning(ctx, payload.StreamingFunc, reasoning); err != nil {
				return nil, fmt.Errorf("streaming reasoning func returned an error: %w", err)
			}
			if err := streaming.CallWithText(ctx, payload.StreamingFunc, content); err != nil {
				return nil, fmt.Errorf("streaming text func returned an error: %w", err)
			}

			if choice.Delta.FunctionCall != nil {
				functionCall := choice.Delta.FunctionCall
				updateFunctionCall(&responseChoice.Message, functionCall)

				toolCall := streaming.NewToolCall("", functionCall.Name, functionCall.Arguments)
				if err := streaming.CallWithToolCall(ctx, payload.StreamingFunc, toolCall); err != nil {
					return nil, fmt.Errorf("streaming tool call func returned an error: %w", err)
				}
			}

			for _, toolCall := range choice.Delta.ToolCalls {
				updateToolCall(&responseChoice.Message, toolCall, toolCallNameCache)

				toolCall := streaming.NewToolCall(toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)
				if err := streaming.CallWithToolCall(ctx, payload.StreamingFunc, toolCall); err != nil {
					return nil, fmt.Errorf("streaming tool call func returned an error: %w", err)
				}
			}
		}
	}

	removeEmptyToolCalls(&response)

	return &response, nil
}

func getChunkContent(choice StreamedChatResponseChunk, splitter reasoning.ChunkContentSplitter) (string, string) {
	content := choice.Delta.Content
	reasoningContent := choice.Delta.ReasoningContent

	// Fallback to legacy reasoning field if reasoningContent is empty
	if reasoningContent == "" {
		reasoningContent = choice.Delta.Reasoning
	}

	// If reasoning content is received separately from the main content, just return it
	if reasoningContent != "" {
		return content, reasoningContent
	}

	// Try to split the content into content and reasoning content
	return splitter.Split(content)
}

func updateChatUsage(chatUsage *ChatUsage, streamUsage *Usage) {
	if streamUsage == nil {
		return
	}

	chatUsage.CompletionTokens = streamUsage.CompletionTokens
	chatUsage.PromptTokens = streamUsage.PromptTokens
	chatUsage.TotalTokens = streamUsage.TotalTokens
	chatUsage.PromptTokensDetails.AudioTokens = streamUsage.PromptTokensDetails.AudioTokens
	chatUsage.PromptTokensDetails.CachedTokens = streamUsage.PromptTokensDetails.CachedTokens
	chatUsage.CompletionTokensDetails.AudioTokens = streamUsage.CompletionTokensDetails.AudioTokens
	chatUsage.CompletionTokensDetails.AcceptedPredictionTokens = streamUsage.CompletionTokensDetails.AcceptedPredictionTokens
	chatUsage.CompletionTokensDetails.RejectedPredictionTokens = streamUsage.CompletionTokensDetails.RejectedPredictionTokens
	chatUsage.CompletionTokensDetails.ReasoningTokens = streamUsage.CompletionTokensDetails.ReasoningTokens
}

func updateFunctionCall(message *ChatMessage, functionCall *FunctionCall) {
	if message.FunctionCall == nil {
		message.FunctionCall = functionCall
	} else {
		message.FunctionCall.Arguments += functionCall.Arguments
	}
}

func updateToolCall(message *ChatMessage, delta *StreamedToolCall, nameCache map[string]string) {
	if delta == nil || nameCache == nil {
		return
	}

	// If index is not set, update the last tool call by rules
	if delta.Index == nil {
		// It's the first delta chunk, have to append a new tool call
		if delta.ID != "" && delta.Type != "" && delta.Function.Name != "" {
			message.ToolCalls = append(message.ToolCalls, ToolCall{})
		}
		// Get the index of the last tool call
		lastIdx := len(message.ToolCalls) - 1
		delta.Index = &lastIdx
	}

	// Grow the tool calls slice to the length of the index
	for idx := range *delta.Index + 1 {
		if len(message.ToolCalls) <= idx {
			message.ToolCalls = append(message.ToolCalls, ToolCall{})
		}
	}

	// Get current tool call which is being updated
	toolCall := &message.ToolCalls[*delta.Index]

	// If it is the first delta chunk, set the tool call fields to the current tool call
	if delta.ID != "" && delta.Type != "" && delta.Function.Name != "" {
		toolCall.ID = delta.ID
		toolCall.Type = delta.Type
		toolCall.Function.Name = delta.Function.Name
		toolCall.Function.Arguments = delta.Function.Arguments

		// Cache the tool call name by ID for subsequent chunks
		nameCache[delta.ID] = delta.Function.Name
	}

	// For next delta chunks, append arguments to the current tool call
	if delta.ID == "" {
		// Standard case: no ID in subsequent chunks (most providers)
		toolCall.Function.Arguments += delta.Function.Arguments

		// Complete the tool call fields with stored values from the current tool call
		delta.Function.Name = toolCall.Function.Name
		delta.ID = toolCall.ID
		delta.Type = toolCall.Type
	} else if delta.Function.Name == "" {
		// If ID is present but name is missing (some providers don't send name in subsequent chunks),
		// restore the name from cache
		if cachedName, ok := nameCache[delta.ID]; ok {
			delta.Function.Name = cachedName
			toolCall.Function.Arguments += delta.Function.Arguments
			delta.Type = toolCall.Type
		}
	}
}

// some providers starts streaming tool calls since the first index number istead of zero
func removeEmptyToolCalls(response *ChatCompletionResponse) {
	for _, choice := range response.Choices {
		if len(choice.Message.ToolCalls) == 0 {
			continue
		}
		toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
		for _, toolCall := range choice.Message.ToolCalls {
			if toolCall.ID == "" || toolCall.Function.Name == "" {
				continue
			}
			toolCalls = append(toolCalls, toolCall)
		}
		choice.Message.ToolCalls = toolCalls
	}
}
