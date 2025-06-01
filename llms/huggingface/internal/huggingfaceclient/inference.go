package huggingfaceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrUnexpectedStatusCode = errors.New("unexpected status code")

// InferenceTask is the type of inference task to run.
type InferenceTask string

const (
	InferenceTaskTextGeneration      InferenceTask = "text-generation"
	InferenceTaskText2TextGeneration InferenceTask = "text2text-generation"
)

type inferencePayload struct {
	Model      string     `json:"-"`
	Inputs     string     `json:"inputs"`
	Parameters parameters `json:"parameters,omitempty"`
}

type parameters struct {
	Temperature       float64 `json:"temperature"`
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	MinLength         int     `json:"min_length,omitempty"`
	MaxLength         int     `json:"max_length,omitempty"`
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	Seed              int     `json:"seed,omitempty"`
}

// Chat completions API structures for router-based requests
type chatCompletionsPayload struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Stream      bool          `json:"stream"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionsResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Index int `json:"index"`
	} `json:"choices"`
}

type (
	inferenceResponsePayload []inferenceResponse
	inferenceResponse        struct {
		Text string `json:"generated_text"`
	}
)

func (c *Client) runInference(ctx context.Context, payload *inferencePayload) (inferenceResponsePayload, error) {
	if c.provider != "" {
		// Use chat completions API for router-based requests
		return c.runChatCompletions(ctx, payload)
	}

	// Standard inference API
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	requestURL := fmt.Sprintf("%s/models/%s", c.url, payload.Model)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	// debug print the http request with httputil:

	// reqDump, err := httputil.DumpRequestOut(req, true)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Fprintf(os.Stderr, "%s", reqDump)

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if len(b) > 0 {
			err = fmt.Errorf("%w: %d, body: %s", ErrUnexpectedStatusCode, r.StatusCode, string(b))
		} else {
			err = fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, r.StatusCode)
		}
		return nil, err
	}

	// debug print the http response with httputil:
	// resDump, err := httputil.DumpResponse(r, true)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Fprintf(os.Stderr, "%s", resDump)

	var response inferenceResponsePayload
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// runChatCompletions runs inference using the chat completions API format for router-based requests
func (c *Client) runChatCompletions(ctx context.Context, payload *inferencePayload) (inferenceResponsePayload, error) {
	// Convert parameters to chat completions format
	chatPayload := chatCompletionsPayload{
		Model: payload.Model,
		Messages: []chatMessage{
			{
				Role:    "user",
				Content: payload.Inputs,
			},
		},
		Stream: false,
	}

	// Map parameters
	if payload.Parameters.Temperature > 0 {
		chatPayload.Temperature = &payload.Parameters.Temperature
	}
	if payload.Parameters.TopP > 0 {
		chatPayload.TopP = &payload.Parameters.TopP
	}
	if payload.Parameters.MaxLength > 0 {
		chatPayload.MaxTokens = &payload.Parameters.MaxLength
	}

	payloadBytes, err := json.Marshal(chatPayload)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)

	requestURL := fmt.Sprintf("%s/%s/v1/chat/completions", c.url, c.provider)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if len(b) > 0 {
			err = fmt.Errorf("%w: %d, body: %s", ErrUnexpectedStatusCode, r.StatusCode, string(b))
		} else {
			err = fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, r.StatusCode)
		}
		return nil, err
	}

	var chatResponse chatCompletionsResponse
	err = json.NewDecoder(r.Body).Decode(&chatResponse)
	if err != nil {
		return nil, err
	}

	// Convert chat completions response to inference response format
	if len(chatResponse.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}

	// Convert to the expected response format
	response := make(inferenceResponsePayload, 1)
	response[0] = inferenceResponse{
		Text: chatResponse.Choices[0].Message.Content,
	}

	return response, nil
}
