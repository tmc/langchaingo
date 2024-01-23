package opensearch

import (
	"fmt"
	"io"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func handleResponse(res *opensearchapi.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return body, nil
	}

	return nil, fmt.Errorf("status %d | message: %s", res.StatusCode, string(body))
}
