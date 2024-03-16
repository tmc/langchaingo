package cloudflareclient

import (
	"fmt"
	"net/http"
)

type Client struct {
	httpClient         *http.Client
	accountID          string
	token              string
	baseURL            string
	modelName          string
	embeddingModelName string
	endpointURL        string
	bearerToken        string
}

func NewClient(client *http.Client, accountID, baseURL, token, modelName, embeddingModelName string) *Client {
	if client == nil {
		client = &http.Client{}
	}

	return &Client{
		httpClient:         client,
		accountID:          accountID,
		baseURL:            baseURL,
		token:              token,
		modelName:          modelName,
		embeddingModelName: embeddingModelName,
		endpointURL:        fmt.Sprintf("%s/%s/ai/run/%s", baseURL, accountID, modelName),
		bearerToken:        "Bearer " + token,
	}
}
