package pineconeClient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrIndexExists = errors.New("Index of given name already exists.")

type createIndexPayload struct {
	Metric    string `json:"metric"`
	Pods      int    `json:"pods"`
	Replicas  int    `json:"replicas"`
	PodType   string `json:"pod_type"`
	Name      string `json:"name"`
	Dimension int    `json:"dimension"`
}

func (c Client) createIndex() error {
	payload := createIndexPayload{
		Metric:    c.metric,
		Pods:      c.pods,
		Replicas:  c.replicas,
		PodType:   c.podType,
		Name:      c.IndexName,
		Dimension: c.vectorDimension,
	}

	body, statusCode, err := doRequest(c.context, payload, fmt.Sprintf("https://controller.%s.pinecone.io/databases", c.environment), c.apiKey)
	if err != nil {
		return err
	}
	if statusCode == http.StatusCreated {
		return nil
	}
	if statusCode == http.StatusConflict {
		return ErrIndexExists
	}

	errBuf := new(bytes.Buffer)
	_, err = io.Copy(errBuf, body)
	if err != nil {
		return fmt.Errorf("Error creating index. %s", err)
	}

	return fmt.Errorf("Error creating index. Status code: %v. %s", statusCode, errBuf.String())
}
