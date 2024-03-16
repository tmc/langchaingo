package cloudflareclient

import "net/http"

type Client struct {
	httpClient         *http.Client
	AccountID          string
	Token              string
	ModelName          string
	EmbeddingModelName string
	Url                string
}

func NewClient(client *http.Client, accountID, url, token, modelName, embeddingModelName string) *Client {
	if client == nil {
		client = &http.Client{}
	}

	return &Client{
		httpClient:         client,
		AccountID:          accountID,
		Url:                url,
		Token:              token,
		ModelName:          modelName,
		EmbeddingModelName: embeddingModelName,
	}
}
