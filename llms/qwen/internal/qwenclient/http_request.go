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

type HeaderSetter map[string]string

func PostSSE(ctx context.Context, urll string, body interface{}, sseChan chan string, newHeaders ...HeaderSetter) error {
	var headerList []HeaderSetter

	if len(newHeaders) > 0 {
		headerList = append(headerList, newHeaders...)
	}

	return httpInnerSSE(ctx, "POST", urll, body, sseChan, headerList...)
}

func Post[T any](ctx context.Context, urll string, body interface{}, timeOut time.Duration, newHeaders ...HeaderSetter) (T, error) {
	var headerList []HeaderSetter
	var result T

	if len(newHeaders) > 0 {
		headerList = append(headerList, newHeaders...)
	}

	byteRsp, err := httpInner(ctx, "POST", urll, body, timeOut, headerList...)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(byteRsp, &result)
	if err != nil {
		return result, fmt.Errorf("json.Unmarshal Err: %v", err.Error())
	}

	return result, nil
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

func httpInnerSSE(ctx context.Context, method, url string, body interface{}, sseStream chan string, headerSetters ...HeaderSetter) (err error) {
	// fmt.Println("httpInnerSSE: len headers: ", len(headerSetters), headerSetters[0])
	defer close(sseStream)
	if headerSetters == nil {
		headerSetters = make([]HeaderSetter, 0)
	}

	if (method == "POST" || method == "PUT") && body == nil {
		// logger.E("POST or PUT Method Can't Have a Empty Body")
		err = errors.New("POST or PUT Body Cant Be Empty")
		return
	}
	var bodyJson []byte
	if body != nil {
		headerSetters = append(headerSetters, HeaderSetter{"content-type": "application/json"})
		switch body := body.(type) {
		case []byte:
			bodyJson = body
		default:
			bodyJson, err = json.Marshal(body)
			if err != nil {
				panic(err)
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return fmt.Errorf("httpNewRequest Err: %v", err.Error())
	}

	for _, HeaderMap := range headerSetters {
		for k, v := range HeaderMap {
			req.Header.Set(k, v)
		}
	}

	httpCli := http.Client{}

	rsp, err := httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("Http-Req-Err: method:[%v], url:[%v], body:[%v] err:[%v]", method, url, string(bodyJson), err.Error())
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 300 || rsp.StatusCode < 200 {
		result, err_io := io.ReadAll(rsp.Body)
		if err_io != nil {
			err = err_io
			return
		}

		err = fmt.Errorf("request Failed: code: %v, error_msg: %v", rsp.StatusCode, string(result))
		return
	}

	scanner := bufio.NewScanner(rsp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		sseStream <- line
	}

	return
}

func httpInner(ctx context.Context, method, url string, body interface{}, timeOut time.Duration, headerSetters ...HeaderSetter) (result []byte, err error) {

	if (method == "POST" || method == "PUT") && body == nil {
		err = errors.New("POST or PUT Body Cant Be Empty")
		return
	}
	var bodyJson []byte
	if body != nil {
		headerSetters = append(headerSetters, HeaderSetter{"content-type": "application/json"})
		switch body := body.(type) {
		case []byte:
			bodyJson = body
		default:
			bodyJson, err = json.Marshal(body)
			if err != nil {
				// logger.E("http url[%v] Body【%+v】 json-Marshal Err: %v", url, body, err.Error())
				return
			}
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyJson))
	if err != nil {
		return
	}

	for _, HeaderMap := range headerSetters {
		for k, v := range HeaderMap {
			req.Header.Set(k, v)
		}
	}

	httpCli := http.Client{}
	if timeOut != 0 {
		httpCli.Timeout = timeOut * time.Second
	}

	rsp, err := httpCli.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	result, err = io.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	if rsp.StatusCode >= 300 || rsp.StatusCode < 200 {
		err = errors.New("request Failed: " + string(result))
		return
	}

	return
}
