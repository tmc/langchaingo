package openaiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type TranscribeAudioResponse struct {
	Text string `json:"text"`
}

func (c *Client) uploadAudioAndGetTranscription(ctx context.Context, audioFilePath, language string, temperature float64) ([]byte, error) {

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(audioFilePath)

	if errFile1 != nil {
		return nil, errFile1
	}

	defer file.Close()

	part1, errFile1 := writer.CreateFormFile("file", filepath.Base(audioFilePath))
	if errFile1 != nil {
		return nil, errFile1
	}
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		return nil, errFile1
	}

	_ = writer.WriteField("model", c.Model)
	_ = writer.WriteField("response_format", "json")
	_ = writer.WriteField("temperature", fmt.Sprintf("%f", temperature))
	_ = writer.WriteField("language", language)
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/audio/transcriptions", payload)

	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var transcriptionResponse TranscribeAudioResponse
	err = json.Unmarshal(body, &transcriptionResponse)
	if err != nil {
		return nil, err
	}
	return []byte(transcriptionResponse.Text), nil
}
