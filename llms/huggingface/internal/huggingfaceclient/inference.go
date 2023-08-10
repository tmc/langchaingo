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
	Temperature       float64 `json:"temperature,omitempty"`
	TopP              float64 `json:"top_p,omitempty"`
	TopK              int     `json:"top_k,omitempty"`
	MinLength         int     `json:"min_length,omitempty"`
	MaxLength         int     `json:"max_length,omitempty"`
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	Seed              int     `json:"seed,omitempty"`
}

type (
	inferenceResponsePayload []inferenceResponse
	inferenceResponse        struct {
		Text string `json:"generated_text"`
	}
)

const huggingfaceAPIBaseURL = "https://api-inference.huggingface.co"

func (c *Client) runInference(ctx context.Context, payload *inferencePayload) (inferenceResponsePayload, error) {
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

	var response inferenceResponsePayload
	err = json.NewDecoder(r.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
