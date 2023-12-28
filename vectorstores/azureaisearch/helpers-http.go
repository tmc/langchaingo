package azureaisearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (s *Store) HTTPDefaultSend(req *http.Request, serviceName string, output any) error {
	response, err := s.client.Do(req)
	if err != nil {
		fmt.Printf("err sending request for "+serviceName+": %v\n", err)
		return err
	}

	return HTTPReadBody(response, serviceName, output)
}

func HTTPReadBody(response *http.Response, serviceName string, output any) error {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)

	if err != nil {
		fmt.Printf("err can't read response for "+serviceName+": %v\n", err)
		return err
	}

	if output != nil {
		// fmt.Printf("body: %v\n", string(body))
		if err := json.Unmarshal(body, output); err != nil {
			fmt.Printf("err unmarshal body for "+serviceName+": %v\n", err)
			fmt.Printf("body: %v\n", string(body))
			return err
		}
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		if output != nil {
			return json.Unmarshal(body, output)
		}
		return nil
	} else {
		fmt.Printf("response: %v\n", string(body))
	}

	return errors.New("Error returned from " + serviceName + " " + response.Status + " | " + strconv.Itoa(response.StatusCode))
}
