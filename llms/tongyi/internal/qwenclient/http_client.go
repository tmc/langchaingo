package qwenclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"time"
)

type HTTPOption func(c *HTTPCli)

// unit test: mockgen -destination=http_client_mock.go -package=qwen_client . IHttpClient.
type IHttpClient interface {
	PostSSE(ctx context.Context, urll string, reqbody interface{}, options ...HTTPOption) (chan string, error)
	Post(ctx context.Context, urll string, reqbody interface{}, resp interface{}, options ...HTTPOption) error
}

type HTTPCli struct {
	client http.Client
	req    *http.Request

	sseStream chan string
}

func NewHTTPClient() *HTTPCli {
	return &HTTPCli{
		client:    http.Client{},
		sseStream: nil,
	}
}

type HeaderMap map[string]string

func WithHeader(header HeaderMap) HTTPOption {
	return func(c *HTTPCli) {
		for k, v := range header {
			c.req.Header.Set(k, v)
		}
	}
}

func WithTimeout(timeout time.Duration) HTTPOption {
	return func(c *HTTPCli) {
		c.client.Timeout = timeout
	}
}

func withStream() HTTPOption {
	return func(c *HTTPCli) {
		c.req.Header.Set("Accept", "text/event-stream")
	}
}

// nolint:lll
func (c *HTTPCli) PostSSE(ctx context.Context, urll string, reqbody interface{}, options ...HTTPOption) (chan string, error) {
	if reqbody == nil {
		err := &EmptyRequestBodyError{}
		return nil, err
	}

	chanBuffer := 500
	sseStream := make(chan string, chanBuffer)
	c.sseStream = sseStream

	options = append(options, withStream(), WithHeader(HeaderMap{"content-type": "application/json"}))

	errChan := make(chan error)

	go func() {
		resp, err := c.httpInner(ctx, "POST", urll, reqbody, options...)
		if err != nil {
			errChan <- err
		} else {
			errChan <- nil
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			c.sseStream <- line
		}

		close(c.sseStream)
	}()

	err := <-errChan
	if err != nil {
		return nil, err
	}

	return c.sseStream, nil
}

// nolint:lll
func (c *HTTPCli) Post(ctx context.Context, urll string, reqbody interface{}, respbody interface{}, options ...HTTPOption) error {
	options = append(options, WithHeader(HeaderMap{"content-type": "application/json"}))

	if reqbody == nil {
		err := &EmptyRequestBodyError{}
		return err
	}

	resp, err := c.httpInner(ctx, "POST", urll, reqbody, options...)
	if err != nil {
		return err
	}
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(result, &respbody)
	if err != nil {
		return &WrapMessageError{Message: "Unmarshal Json failed", Cause: err}
	}

	return nil
}

func (c *HTTPCli) EncodeJSONBody(body interface{}) ([]byte, error) {
	var err error

	var bodyJSON []byte
	if body != nil {
		switch body := body.(type) {
		case []byte:
			bodyJSON = body
		default:
			bodyJSON, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}
	}
	return bodyJSON, nil
}

// nolint:lll
func (c *HTTPCli) httpInner(ctx context.Context, method, url string, body interface{}, options ...HTTPOption) (*http.Response, error) {
	var err error

	bodyJSON, err := c.EncodeJSONBody(body)
	if err != nil {
		return nil, err
	}

	c.req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		option(c)
	}

	resp, err := c.client.Do(c.req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		result, errIo := io.ReadAll(resp.Body)
		if errIo != nil {
			err = errIo
			return resp, err
		}

		err = &DashscopeError{Message: "request Failed: " + string(result), Code: resp.StatusCode}
		return resp, err
	}

	// Note: close rsp at outer function
	return resp, nil
}

func NetworkStatus() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "8.8.8.8")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return false, ErrNetwork
	}

	return true, nil
}
