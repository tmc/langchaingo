package huggingfaceclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
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
	Stream     bool       `json:"stream,omitempty"`
}

type parameters struct {
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	MinLength         int      `json:"min_length,omitempty"`
	MaxLength         int      `json:"max_length,omitempty"`
	RepetitionPenalty float64  `json:"repetition_penalty,omitempty"`
	Seed              int      `json:"seed,omitempty"`
	Stop              []string `json:"stop,omitempty"`
}

type StreamResponse struct {
	Details struct {
		FinishReason    string `json:"finish_reason"`    // Possible values: length, eos_token, stop_sequence
		GeneratedTokens int    `json:"generated_tokens"` // Number of generated tokens
		InputLength     int    `json:"input_length"`     // Length of the input
		Seed            int    `json:"seed"`             // Seed used for generation
	} `json:"details"` // Details about the generation
	GeneratedText string  `json:"generated_text"` // Generated text
	Index         int     `json:"index"`          // Index of the response
	Token         Token   `json:"token"`          // Token information
	TopTokens     []Token `json:"top_tokens"`     // List of top tokens
}

type Token struct {
	ID      int     `json:"id"`      // Token ID
	Logprob float64 `json:"logprob"` // Log probability of the token
	Special bool    `json:"special"` // Whether the token is special
	Text    string  `json:"text"`    // Text representation of the token
}

type (
	inferenceResponsePayload []inferenceResponse
	inferenceResponse        struct {
		Text string `json:"generated_text"`
	}
)

func (c *Client) runInference(ctx context.Context, payload *inferencePayload, options ...*llms.CallOptions) (inferenceResponsePayload, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/models/%s", c.url, payload.Model), body) //nolint:lll
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

	r, err := http.DefaultClient.Do(req)
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
	if payload.Stream && len(options) == 1 && options[0].StreamingFunc != nil {
		return parseStreamingMessageResponse(ctx, r, options[0].StreamingFunc)
	}
	var response inferenceResponsePayload
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func parseStreamingMessageResponse(ctx context.Context, r *http.Response, streamingFunc func(ctx context.Context, chunk []byte) error) (inferenceResponsePayload, error) {
	scanner := bufio.NewScanner(r.Body)
	var lastResponse string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var resp StreamResponse
		err := json.Unmarshal([]byte(data), &resp)
		if err != nil {
			return nil, err
		}
		err = streamingFunc(ctx, []byte(resp.Token.Text))
		if err != nil {
			return nil, err
		}
		if resp.GeneratedText != "" {
			lastResponse = resp.GeneratedText
		}
	}
	return []inferenceResponse{{Text: lastResponse}}, nil
}
