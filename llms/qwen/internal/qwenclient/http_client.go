package qwen_client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

type HttpOption func(c *HttpCli)

// unit test: mockgen -destination=http_client_mock.go -package=qwen_client . IHttpClient
type IHttpClient interface {
	PostSSE(ctx context.Context, urll string, reqbody interface{}, options ...HttpOption) (chan string, error)
	Post(ctx context.Context, urll string, reqbody interface{}, resp interface{}, options ...HttpOption) error
}

type HttpCli struct {
	client http.Client
	req    *http.Request

	sseStream chan string
}

func NewHttpClient() *HttpCli {
	return &HttpCli{
		client:    http.Client{},
		sseStream: nil,
	}
}

type HeaderMap map[string]string

func WithHeader(header HeaderMap) HttpOption {
	return func(c *HttpCli) {
		for k, v := range header {
			c.req.Header.Set(k, v)
		}
	}
}

func WithTimeout(timeout time.Duration) HttpOption {
	return func(c *HttpCli) {
		c.client.Timeout = timeout
	}
}

func withStream() HttpOption {
	return func(c *HttpCli) {
		c.req.Header.Set("Accept", "text/event-stream")
	}
}

func (c *HttpCli) PostSSE(ctx context.Context, urll string, body interface{}, options ...HttpOption) (chan string, error) {
	sseStream := make(chan string, 500)
	c.sseStream = sseStream

	options = append(options, withStream())

	resp, err := c.httpInner(ctx, "POST", urll, body, options...)
	if err != nil {
		return nil, err
	}
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			c.sseStream <- line
		}

		close(c.sseStream)
		resp.Body.Close()
	}()


	return c.sseStream, nil
}

func (c *HttpCli) Post(ctx context.Context, urll string, reqbody interface{}, respbody interface{}, options ...HttpOption) error {

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
		return fmt.Errorf("json.Unmarshal Err: %v", err.Error())
	}

	return nil
}

func (c *HttpCli) httpInner(ctx context.Context, method, url string, body interface{}, options ...HttpOption) (*http.Response, error) {
	var err error

	if (method == "POST" || method == "PUT") && body == nil {
		err := errors.New("POST or PUT Body Cant Be Empty")
		return nil, err
	}
	var bodyJson []byte
	if body != nil {
		options = append(options, WithHeader(HeaderMap{"content-type": "application/json"}))

		switch body := body.(type) {
		case []byte:
			bodyJson = body
		default:
			bodyJson, err = json.Marshal(body)
			if err != nil {
				return nil, err
			}
		}
	}

	c.req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyJson))
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
		result, err_io := io.ReadAll(resp.Body)
		if err_io != nil {
			err = err_io
			return resp, err
		}

		err = fmt.Errorf("request Failed: code: %v, error_msg: %v", resp.StatusCode, string(result))
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
		return false, fmt.Errorf("network error")
	}

	return true, nil
}
