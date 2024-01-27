package opensearch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func handleResponse(output any, res *opensearchapi.Response) error {
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if output == nil {
			return nil
		}

		return json.Unmarshal(body, &output)
	}

	return fmt.Errorf("status %d | message: %s", res.StatusCode, string(body))
}
