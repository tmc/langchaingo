package metaphor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/metaphorsystems/metaphor-go"
	"github.com/tmc/langchaingo/tools"
)

var _ tools.Tool = &API{}

// API defines a tool implementation for the Metaphor API.
type API struct {
	client *metaphor.Client
}

// ToolInput defines a struct the tool expects as input.
type ToolInput struct {
	Operation  string                  `json:"operation"`
	Input      string                  `json:"input"`
	ReqOptions metaphor.RequestOptions `json:"reqOptions"`
}

// NewClient initializes a new API client.
//
// It retrieves the API key from the environment variable "METAPHOR_API_KEY"
// and creates a new client using the retrieved API key. If the API key is not
// set or an error occurs during client creation, an error is returned.
//
// Returns a pointer to the created API client and an error, if any.
func NewClient() (*API, error) {
	apiKey := os.Getenv("METAPHOR_API_KEY")

	client, err := metaphor.NewClient(apiKey)
	if err != nil {
		return nil, err
	}

	return &API{
		client: client,
	}, nil
}

// Name returns the name of the tool.
//
// No parameters.
// Returns a string.
func (tool *API) Name() string {
	return "Metaphor API Tool"
}

// Description returns the Description of the tool.
// Description contains a short instruction how to use the tool
// with the Metaphor API
//
// No parameters.
// Returns a string.
func (tool *API) Description() string {
	return `
	Metaphor API Tool is a tool to interact with the Metaphor API. Metaphor is a search engine
	trained to do link prediction.
	This means that given some text prompt, it tries to predict the link that would most likely
	follow that prompt. This tool shouls be used when you want to add spcific filters to your search qeury

	Tool expects string json of the format as input:
	{
		"operation": "YourOperation",
		"input": "YourInput",
		"reqOptions": {
			"numResults": 10,
			"includeDomains": ["example.com", "example2.com"],
			"excludeDomains": ["exclude.com"],
			"startCrawlDate": "2023-08-15T00:00:00Z",
			"endCrawlDate": "2023-08-16T00:00:00Z",
			"startPublishedDate": "2023-08-15T00:00:00Z",
			"endPublishedDate": "2023-08-16T00:00:00Z",
			"useAutoprompt": true,
			"type": "neural"
		}
	}

	Input json should be built from API reference and the following instructions:

	- operation: Api call to be performed, possible values: "Search", "FindSimilar", "GetContents"
	- input: value of the search query or link, for search and findSimilar endpoints respectively
	- reqOptions: json of options API parameters

	Note: Omit any fields in the reqOptions that you're not going to use in the call.

	API Reference:
	- Search
		POST https://api.metaphor.systems/search
		Perform a search with a Metaphor prompt-engineered query and retrieve a list of relevant results.

		Unique BODY PARAM for Search:
		query (string, required)
			The query string for the search. It's vital that the query takes the form of a declarative
			suggestion, where a high-quality search result link would follow. For example,
			'best restaurants in SF' is a bad query,
			whereas 'Here is the best restaurant in SF:' is a good query.

	- Find similar links
		POST https://api.metaphor.systems/findSimilar
		Find similar links to the link provided.

		Unique BODY PARAM for Find similar links:
		url (string, required)
			The URL for which you would like to find similar links.

	Common BODY PARAMS for both endpoints:
		numResults (integer)
			Number of search results to return. Default 10. Up to 30 for basic plans. Up to thousands for custom plans.

		includeDomains (array of strings)
			Optional list of domains to include in the search. Results will only come from these domains.

		excludeDomains (array of strings)
			Optional list of domains to exclude from the search. Results will not include any from these domains.

		startCrawlDate (date-time)
			Optional start date for the crawled data in ISO 8601 format.
			Search will only include results crawled on or after this date.

		endCrawlDate (date-time)
			Optional end date for the crawled data in ISO 8601 format.
			Search will only include results crawled on or before this date.

		startPublishedDate (date-time)
			Optional start date for the published data in ISO 8601 format.
			Search will only include results published on or after this date.

		endPublishedDate (date-time)
			Optional end date for the published data in ISO 8601 format.
			Search will only include results published on or before this date.

	- Get contents of documents
		GET https://api.metaphor.systems/contents
		Retrieve contents of documents based on a list of document IDs.

		QUERY PARAMS
		ids (array of strings, required)
			An array of document IDs obtained from either /search or /findSimilar endpoints.`
}

// Call is a function that takes a context and an input string and returns a string and an error.
//
// The function expects a JSON string as input and unmarshals it into a ToolInput struct.
// It then performs different operations based on the value of the Operation field in the ToolInput struct.
// The supported operations are "Search", "FindSimilar", and "GetContents".
//
// If the Operation is "Search", the function calls the performSearch method passing the
// context and the ToolInput struct.
// If the Operation is "FindSimilar", the function calls the findSimilar method passing the
// context and the ToolInput struct.
// If the Operation is "GetContents", the function calls the getContents method passing the
// context and the ToolInput struct.
//
// The function returns the result of the respective operation or an empty string and nil
// if the Operation is not supported.
func (tool *API) Call(ctx context.Context, input string) (string, error) {
	var toolInput ToolInput

	re := regexp.MustCompile(`(?s)\{.*\}`)
	jsonString := re.FindString(input)

	err := json.Unmarshal([]byte(jsonString), &toolInput)
	if err != nil {
		return "", err
	}

	switch toolInput.Operation {
	case "Search":
		return tool.performSearch(ctx, toolInput)
	case "FindSimilar":
		return tool.findSimilar(ctx, toolInput)
	case "GetContents":
		return tool.getContents(ctx, toolInput)
	}

	return "", nil
}

func (tool *API) performSearch(ctx context.Context, toolInput ToolInput) (string, error) {
	response, err := tool.client.Search(
		ctx,
		toolInput.Input,
		metaphor.WithRequestOptions(&toolInput.ReqOptions),
	)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoSearchResults) {
			return "Metaphor Search didn't return any results", nil
		}
		return "", err
	}
	return tool.formatResults(response), err
}

func (tool *API) findSimilar(ctx context.Context, toolInput ToolInput) (string, error) {
	response, err := tool.client.FindSimilar(
		ctx,
		toolInput.Input,
		metaphor.WithRequestOptions(&toolInput.ReqOptions),
	)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoLinksFound) {
			return "Metaphor Links Search didn't return any results", nil
		}
		return "", err
	}
	return tool.formatResults(response), err
}

func (tool *API) getContents(ctx context.Context, toolInput ToolInput) (string, error) {
	ids := strings.Split(toolInput.Input, ",")
	for i, id := range ids {
		ids[i] = strings.TrimSpace(id)
	}

	response, err := tool.client.GetContents(ctx, ids)
	if err != nil {
		if errors.Is(err, metaphor.ErrNoContentExtracted) {
			return "Metaphor Extractor didn't return any results", nil
		}
		return "", err
	}

	return tool.formatContents(response), err
}

func (tool *API) formatResults(response *metaphor.SearchResponse) string {
	formattedResults := ""

	for _, result := range response.Results {
		formattedResults += fmt.Sprintf("Title: %s\nURL: %s\nID: %s\n\n", result.Title, result.URL, result.ID)
	}

	return formattedResults
}

func (tool *API) formatContents(response *metaphor.ContentsResponse) string {
	formattedResults := ""

	for _, result := range response.Contents {
		formattedResults += fmt.Sprintf("Title: %s\nContent: %s\nURL: %s\n\n", result.Title, result.Extract, result.URL)
	}

	return formattedResults
}
