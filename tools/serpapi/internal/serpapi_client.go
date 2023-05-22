package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrMissingToken = errors.New("missing the SerpAPI API key, set it in the SERPAPI_API_KEY environment variable")
	ErrNoGoodResult = errors.New("no good search results found")
)

type Client struct {
	apiKey  string
	baseURL string
}

func New(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://serpapi.com/search",
	}
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	params := make(url.Values)
	query = strings.ReplaceAll(query, " ", "+")
	params.Add("q", query)
	params.Add("google_domain", "google.com")
	params.Add("gl", "us")
	params.Add("hl", "en")
	params.Add("api_key", s.apiKey)

	reqURL := fmt.Sprintf("%s?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request in serpapi: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("doing response in serpapi: %w", err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return "", fmt.Errorf("coping data in serpapi: %w", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return "", fmt.Errorf("unmarshal data in serpapi: %w", err)
	}

	return processResponse(result)
}

func processResponse(res map[string]interface{}) (string, error) {
	if errorValue, ok := res["error"]; ok {
		return "", fmt.Errorf("Got error from SerpAPI: %v", errorValue)
	}

	answerBox, answerBoxExists := res["answer_box"].(map[string]interface{})
	sportsResults, sportsResultsExists := res["sports_results"].(map[string]interface{})
	knowledgeGraph, knowledgeGraphExists := res["knowledge_graph"].(map[string]interface{})
	organicResults, organicResultsExists := res["organic_results"].([]interface{})

	if answerBoxExists {
		if answer, ok := answerBox["answer"].(string); ok {
			return answer, nil
		}
		if snippet, ok := answerBox["snippet"].(string); ok {
			return snippet, nil
		}
		if snippetHighlightedWords, ok := answerBox["snippet_highlighted_words"].([]interface{}); ok && len(snippetHighlightedWords) > 0 {
			return snippetHighlightedWords[0].(string), nil
		}
	}
	if sportsResultsExists {
		if gameSpotlight, ok := sportsResults["game_spotlight"].(string); ok {
			return gameSpotlight, nil
		}
	}
	if knowledgeGraphExists {
		if description, ok := knowledgeGraph["description"].(string); ok {
			return description, nil
		}
	}
	if organicResultsExists && len(organicResults) > 0 {
		organicResult, ok := organicResults[0].(map[string]interface{})
		if ok {
			if snippet, ok := organicResult["snippet"].(string); ok {
				return snippet, nil
			}
		}
	}

	return "", ErrNoGoodResult
}
