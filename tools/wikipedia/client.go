package wikipedia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const _baseURL = "https://%s.wikipedia.org/w/api.php"

type searchResponse struct {
	Query struct {
		Search []struct {
			Ns        int       `json:"ns"`
			Title     string    `json:"title"`
			PageID    int       `json:"pageid"`
			Size      int       `json:"size"`
			WordCount int       `json:"wordcount"`
			Snippet   string    `json:"snippet"`
			Timestamp time.Time `json:"timestamp"`
		} `json:"search"`
	} `json:"query"`
}

func search(
	ctx context.Context,
	limit int,
	query,
	languageCode,
	userAgent string,
) (searchResponse, error) {
	params := make(url.Values)
	params.Add("format", "json")
	params.Add("action", "query")
	params.Add("list", "search")
	params.Add("srsearch", query)
	params.Add("srlimit", fmt.Sprintf("%v", limit))

	reqURL := fmt.Sprintf("%s?%s", fmt.Sprintf(_baseURL, languageCode), params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return searchResponse{}, fmt.Errorf("creating request in wikipedia: %w ", err)
	}
	req.Header.Add("User-Agent", userAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return searchResponse{}, fmt.Errorf("doing response in wikipedia: %w", err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return searchResponse{}, fmt.Errorf("coping data in wikipedia: %w", err)
	}

	var result searchResponse
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return searchResponse{}, fmt.Errorf("unmarshal data in wikipedia: %w", err)
	}

	return result, nil
}

type pageResult struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

func getPage(ctx context.Context, pageID int, languageCode, userAgent string) (pageResult, error) {
	params := make(url.Values)
	params.Add("format", "json")
	params.Add("action", "query")
	params.Add("prop", "extracts")
	params.Add("pageids", fmt.Sprintf("%v", (pageID)))

	reqURL := fmt.Sprintf("%s?%s", fmt.Sprintf(_baseURL, languageCode), params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return pageResult{}, fmt.Errorf("creating request in wikipedia: %w ", err)
	}
	req.Header.Add("User-Agent", userAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return pageResult{}, fmt.Errorf("doing response in wikipedia: %w", err)
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return pageResult{}, fmt.Errorf("coping data in wikipedia: %w", err)
	}

	var result pageResult
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return pageResult{}, fmt.Errorf("unmarshal data in wikipedia: %w", err)
	}

	return result, nil
}
