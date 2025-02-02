package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type TTSRequest struct {
	// Model specifies the TTS model to use. "tts-1" is the default model and tts-1-hd is the high definition model.
	Model string `json:"model" binding:"required" default:"tts-1"`

	// Input is the text that will be converted to speech. This field is required.
	Input string `json:"input" binding:"required"`

	// Voice determines which voice style to use for synthesis. "alloy" is the default option and other options are "ash", "coral", "echo", "fable", "onyx", "nova", "sage" and "shimmer".
	Voice string `json:"voice" binding:"required" default:"alloy"`

	// ResponseFormat specifies the output format of the audio file.
	// Defaults to "mp3" but other formats may be supported like "opus", "aac", "flac", "wav", and "pcm".
	ResponseFormat string `json:"response_format"`

	// Speed controls the speaking rate of the generated audio.
	// Acceptable range is 0.25 to 4.0, with 1.0 as the default normal speed.
	Speed float64 `json:"speed"`
}

type TTSModel string

const (
	TTS1   TTSModel = "tts-1"
	TTS1HD TTSModel = "tts-1-hd"
)

const (
	defaultTTSModel = TTS1
)

type TTSVoice string

const (
	Alloy   TTSVoice = "alloy"
	Ash     TTSVoice = "ash"
	Coral   TTSVoice = "coral"
	Echo    TTSVoice = "echo"
	Fable   TTSVoice = "fable"
	Onyx    TTSVoice = "onyx"
	Nova    TTSVoice = "nova"
	Sage    TTSVoice = "sage"
	Shimmer TTSVoice = "shimmer"
)

type TTSResponseFormat string

const (
	MP3  TTSResponseFormat = "mp3"
	WAV  TTSResponseFormat = "wav"
	OPUS TTSResponseFormat = "opus"
	AAC  TTSResponseFormat = "aac"
	FLAC TTSResponseFormat = "flac"
	PCM  TTSResponseFormat = "pcm"
)

func (c *Client) setTTSDefaults(payload *TTSRequest) {
	// Set defaults

	switch {
	// Prefer the model specified in the payload.
	case payload.Model != "":

	// If no model is set in the payload, take the one specified in the client.
	case c.Model != "":
		payload.Model = c.Model
	// Fallback: use the default model
	default:
		payload.Model = string(defaultTTSModel)
	}

	if payload.ResponseFormat == "" {
		payload.ResponseFormat = string(MP3)
	}

	if payload.Speed == 0 {
		payload.Speed = 1.0
	}

	if payload.Voice == "" {
		payload.Voice = string(Alloy)
	}
}

func (c *Client) CreateTTS(ctx context.Context, payload *TTSRequest) ([]byte, error) {
	c.setTTSDefaults(payload)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Build request
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL("/audio/speech", payload.Model), body)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	// Send request
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", response.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(response.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg) // nolint:goerr113
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message) // nolint:goerr113
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return responseData, nil

}
