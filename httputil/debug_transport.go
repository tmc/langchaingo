package httputil

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// DebugHTTPClient is an http.Client that logs the request and response with full contents.
var DebugHTTPClient = &http.Client{
	Transport: &logTransport{http.DefaultTransport},
}

type logTransport struct {
	Transport http.RoundTripper
}

// RoundTrip logs the request and response with full contents using httputil.DumpRequest and httputil.DumpResponse
func (t *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", dump)
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s\n", dump)
	return resp, nil
}
