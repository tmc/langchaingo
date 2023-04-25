package pineconeClient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type Client struct {
	apiKey          string
	environment     string
	vectorDimension int
	IndexName       string
	projectName     string
	pods            int
	podType         string
	context         context.Context
	metric          string
	replicas        int
}

func New(options ...ClientOption) (Client, error) {
	c := Client{
		vectorDimension: -1,
		pods:            1,
		replicas:        1,
		podType:         "s1",
		metric:          "cosine",
		context:         context.Background(),
	}

	for _, option := range options {
		option(&c)
	}

	if c.apiKey == "" {
		return c, fmt.Errorf("No value set for api key. Use WithApiKey when creating a new client")
	}

	if c.environment == "" {
		return c, fmt.Errorf("No value set for environment. Use WithEnvironment when creating a new client")
	}

	if c.vectorDimension < 0 {
		return c, fmt.Errorf("No value set for vector dimension. Use WithDimension when creating a new client")
	}

	// Get project name associated with api using the whoami command
	var err error
	c.projectName, err = c.whoami()
	if err != nil {
		return c, err
	}

	if c.IndexName == "" {
		c.IndexName = uuid.New().String()
		err := c.createIndex()
		return c, err
	}

	err = c.createIndex()
	if err != nil {
		if err != ErrIndexExists {
			return c, nil
		}

		return c, err
	}

	return c, nil
}

type whoamiResponse struct {
	ProjectName string `json:"project_name"`
	UserLabel   string `json:"user_label"`
	UserName    string `json:"user_name"`
}

func (c Client) whoami() (string, error) {
	req, err := http.NewRequestWithContext(c.context, "GET", fmt.Sprintf("https://controller.%s.pinecone.io/actions/whoami", c.environment), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Api-Key", c.apiKey)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	var response whoamiResponse

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&response)
	return response.ProjectName, err
}

func (c Client) getEndpoint() string {
	urlString := url.QueryEscape(fmt.Sprintf("%s-%s.svc.%s.pinecone.io", c.IndexName, c.projectName, c.environment))
	return "https://" + urlString
}

type ClientOption func(*Client)

// Must be set.
func WithApiKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// Must be set.
func WithEnvironment(environment string) ClientOption {
	return func(c *Client) {
		c.environment = environment
	}
}

// Must be set.
func WithDimensions(vectorDimension int) ClientOption {
	return func(c *Client) {
		c.vectorDimension = vectorDimension
	}
}

// The name of the index to. The maximum length is 45 characters. If not set the name is a UUID.
func WithIndexName(name string) ClientOption {
	return func(c *Client) {
		c.IndexName = name
	}
}

// Default is one.
func WithPods(numPods int) ClientOption {
	return func(c *Client) {
		c.pods = numPods
	}
}

// Default is "s1".
func WithPodType(podType string) ClientOption {
	return func(c *Client) {
		c.podType = podType
	}
}

// Context used for the create index and whoami calls.
func WithContext(ctx context.Context) ClientOption {
	return func(c *Client) {
		c.context = ctx
	}
}

type Metric string

const (
	Euclidean  Metric = "euclidean"
	cosine     Metric = "cosine"
	Dotproduct Metric = "dotproduct"
)

// The distance metric to be used for similarity search. You can use 'euclidean', 'cosine', or 'dotproduct'. Default is cosine.
func WithMetric(metric Metric) ClientOption {
	return func(c *Client) {
		c.metric = string(metric)
	}
}

// Default is one.
func WithReplicas(replicas int) ClientOption {
	return func(c *Client) {
		c.replicas = replicas
	}
}
