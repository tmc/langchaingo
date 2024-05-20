package httputil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/fatih/color"
)

var DebugHTTPColorJSON = &http.Client{ //nolint:gochecknoglobals
	Transport: &logJSONTransport{http.DefaultTransport},
}

type logJSONTransport struct {
	Transport http.RoundTripper
}

func (t *logJSONTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	var requestBodyJSON bytes.Buffer
	if err := json.Indent(&requestBodyJSON, requestBody, "", " "); err != nil {
		return nil, err
	}
	color.Blue(requestBodyJSON.String()) //nolint:forbidigo

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	var responseBodyJSON bytes.Buffer
	if err := json.Indent(&responseBodyJSON, responseBody, "", " "); err != nil {
		return nil, err
	}
	color.Green(responseBodyJSON.String()) //nolint:forbidigo

	return resp, nil
}
