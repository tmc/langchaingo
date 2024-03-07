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
)

const _url = "https://serpapi.com/search"

var (
	ErrNoGoodResult = errors.New("no good search results found")
	ErrAPIError     = errors.New("error from SerpAPI")
)

type Client struct {
	apiKey string
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

func (s *Client) Search(ctx context.Context, query string) (string, error) {
	params := make(url.Values)
	params.Add("q", query)
	params.Add("google_domain", "google.com")
	params.Add("gl", "us")
	params.Add("hl", "en")
	params.Add("api_key", s.apiKey)

	reqURL := fmt.Sprintf("%s?%s", _url, params.Encode())
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
		return "", fmt.Errorf("%w: %v", ErrAPIError, errorValue)
	}
	if res := getAnswerBox(res); res != "" {
		return res, nil
	}
	if res := getSportResult(res); res != "" {
		return res, nil
	}
	if res := getKnowledgeGraph(res); res != "" {
		return res, nil
	}
	if res := getOrganicResult(res); res != "" {
		return res, nil
	}

	return "", ErrNoGoodResult
}

func getAnswerBox(res map[string]interface{}) string {
	answerBox, answerBoxExists := res["answer_box"].(map[string]interface{})
	if answerBoxExists {
		if answer, ok := answerBox["answer"].(string); ok {
			return answer
		}
		if snippet, ok := answerBox["snippet"].(string); ok {
			return snippet
		}
		snippetHighlightedWords, ok := answerBox["snippet_highlighted_words"].([]interface{})
		if ok && len(snippetHighlightedWords) > 0 {
			return fmt.Sprintf("%v", snippetHighlightedWords[0])
		}
	}

	return ""
}

func getSportResult(res map[string]interface{}) string {
	sportsResults, sportsResultsExists := res["sports_results"].(map[string]interface{})
	if sportsResultsExists {
		if gameSpotlight, ok := sportsResults["game_spotlight"].(string); ok {
			return gameSpotlight
		}
	}

	return ""
}

func getKnowledgeGraph(res map[string]interface{}) string {
	knowledgeGraph, knowledgeGraphExists := res["knowledge_graph"].(map[string]interface{})
	if knowledgeGraphExists {
		if description, ok := knowledgeGraph["description"].(string); ok {
			return description
		}
	}

	return ""
}

func getOrganicResult(res map[string]interface{}) string {
	organicResults, organicResultsExists := res["organic_results"].([]interface{})

	if organicResultsExists && len(organicResults) > 0 {
		organicResult, ok := organicResults[0].(map[string]interface{})
		if ok {
			if snippet, ok := organicResult["snippet"].(string); ok {
				return snippet
			}
		}
	}

	return ""
}
