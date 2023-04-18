package pineconeClient

import (
	"errors"
	"fmt"
	"io/ioutil"
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

	if statusCode == 201 {
		return nil
	}

	if statusCode == 409 {
		return ErrIndexExists
	}

	errMessage, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("Error creating index. %s", err)
	}

	return fmt.Errorf("Error creating index. Status code: %v. %s", statusCode, errMessage)
}
