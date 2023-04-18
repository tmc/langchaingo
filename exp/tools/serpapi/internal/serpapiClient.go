package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var ErrMissingToken = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")
var NoGoodResult = "No good search result found"

type SerpapiClient struct {
	apiKey  string
	baseURL string
}

func New() (*SerpapiClient, error) {
	apiKey := os.Getenv("SERPAPI_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingToken
	}

	return &SerpapiClient{
			apiKey:  apiKey,
			baseURL: "https://serpapi.com/search",
		},
		nil
}

func (s *SerpapiClient) Search(query string) (string, error) {
	var params url.Values
	query = strings.ReplaceAll(query, " ", "+")
	params.Set("q", query)
	params.Add("google_domain", "google.com")
	params.Add("gl", "us")
	params.Add("hl", "en")
	params.Add("api_key", s.apiKey)

	reqURL := fmt.Sprintf("%s?%s", s.baseURL, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return "", err
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

	return NoGoodResult, nil
}
