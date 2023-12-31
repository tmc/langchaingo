package azureaisearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrSendingRequest basic error when the request failed.
var ErrSendingRequest = errors.New(
	"error sedding request",
)

func (s *Store) httpDefaultSend(req *http.Request, serviceName string, output any) error {
	response, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("err sending request for %s: %w", serviceName, err)
	}

	return httpReadBody(response, serviceName, output)
}

func httpReadBody(response *http.Response, serviceName string, output any) error {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("err can't read response for %s: %w", serviceName, err)
	}

	if output != nil {
		if err := json.Unmarshal(body, output); err != nil {
			return fmt.Errorf("err unmarshal body for %s: %w", serviceName, err)
		}
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if output != nil {
			return json.Unmarshal(body, output)
		}
		return nil
	}

	return fmt.Errorf("error returned from %s | Status : %s |  Status Code: %d | body: %s %w",
		serviceName,
		response.Status,
		response.StatusCode,
		string(body),
		ErrSendingRequest,
	)
}
