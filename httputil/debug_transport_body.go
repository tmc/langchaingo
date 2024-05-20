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
	var requestBodyBuffer bytes.Buffer
	if req.Body != nil {
		requestBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		requestBodyBuffer.Write(requestBody)
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}
	var requestBodyJSON bytes.Buffer
	if err := json.Indent(&requestBodyJSON, requestBodyBuffer.Bytes(), "", "  "); err != nil {
		return nil, err
	}
	color.Blue(requestBodyJSON.String()) //nolint:forbidigo

	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	var responseBodyBuffer bytes.Buffer
	if resp.Body != nil {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		responseBodyBuffer.Write(responseBody)
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}
	var responseBodyJSON bytes.Buffer
	if err := json.Indent(&responseBodyJSON, responseBodyBuffer.Bytes(), "", "  "); err != nil {
		return nil, err
	}
	color.Green(responseBodyJSON.String()) //nolint:forbidigo

	return resp, nil
}
