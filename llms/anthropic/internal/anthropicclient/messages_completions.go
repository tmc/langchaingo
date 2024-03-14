package anthropicclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesCompletionPayload struct {
	Model        string         `json:"model"`
	Messages     []*ChatMessage `json:"messages"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	Temperature  float64        `json:"temperature"`
	MaxTokens    int            `json:"max_tokens,omitempty"`
	TopP         float64        `json:"top_p,omitempty"`
	StopWords    []string       `json:"stop_sequences,omitempty"`
	// NOTE: Currently not supported
	//Stream       bool          `json:"stream,omitempty"`

	// NOTE: Currently not supported
	//StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
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
	// TODO: support streaming
	// if payload.StreamingFunc != nil {
	// 	payload.Stream = true
	// }
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
	// TODO: Handle streaming
	// if payload.StreamingFunc != nil {
	// 	// Read chunks
	// 	return parseStreamingMessagesCompletionResponse(ctx, r, payload)
	// }

	// Parse response
	var response MessagesCompletionResponsePayload
	if err := json.NewDecoder(r.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return &response, nil
}

// TODO: Support streaming
// func parseStreamingMessagesCompletionResponse(ctx context.Context, r *http.Response, payload *messagesCompletionPayload) (*MessagesCompletionResponsePayload, error) { // nolint:lll
// 	scanner := bufio.NewScanner(r.Body)
// 	responseChan := make(chan *MessagesCompletionResponsePayload)
// 	go func() {
// 		defer close(responseChan)
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			if line == "" {
// 				continue
// 			}
// 			if !strings.HasPrefix(line, "data:") {
// 				continue
// 			}
// 			data := strings.TrimPrefix(line, "data: ")
// 			streamPayload := &MessagesCompletionResponsePayload{}
// 			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
// 			if err != nil {
// 				log.Fatalf("failed to decode stream payload: %v", err)
// 			}
// 			responseChan <- streamPayload
// 		}
// 		if err := scanner.Err(); err != nil {
// 			log.Println("issue scanning response:", err)
// 		}
// 	}()
// 	// Parse response
// 	response := MessagesCompletionResponsePayload{}
//
// 	var lastResponse *MessagesCompletionResponsePayload
// 	for streamResponse := range responseChan {
// 		response.Completion += streamResponse.Completion
// 		if payload.StreamingFunc != nil {
// 			err := payload.StreamingFunc(ctx, []byte(streamResponse.Completion))
// 			if err != nil {
// 				return nil, fmt.Errorf("streaming func returned an error: %w", err)
// 			}
// 		}
// 		lastResponse = streamResponse
// 	}
// 	response.Model = lastResponse.Model
// 	response.LogID = lastResponse.LogID
// 	response.Stop = lastResponse.Stop
// 	response.StopReason = lastResponse.StopReason
// 	return &response, nil
// }
