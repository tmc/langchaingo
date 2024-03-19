package openaiclient

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) uploadAudioAndGetTranscription(ctx context.Context, audioFilePath, model, responseFormat, language string, temperature string) ([]byte, error) {

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
	_ = writer.WriteField("temperature", temperature)
	_ = writer.WriteField("language", language)
	err := writer.Close()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL("/audio/transcriptions", c.Model), payload)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
