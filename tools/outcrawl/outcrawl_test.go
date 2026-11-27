package outcrawl

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newServer(t *testing.T, status int, body string, check func(t *testing.T, r *http.Request, payload map[string]any)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer test-key")
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		if check != nil {
			check(t, r, payload)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSearchTool(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"hits":[
		{"url":"https://example.com/a","title":"First","snippet":"alpha\n  beta","rank":1},
		{"url":"https://example.com/b","title":"Second","snippet":"","rank":2}]}`,
		func(t *testing.T, r *http.Request, payload map[string]any) {
			if r.URL.Path != "/v1/search" {
				t.Errorf("path = %s, want /v1/search", r.URL.Path)
			}
			if payload["query"] != "go agents" {
				t.Errorf("query = %v", payload["query"])
			}
			if payload["limit"] != float64(2) {
				t.Errorf("limit = %v, want 2", payload["limit"])
			}
		})

	tool, err := NewSearch(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"), WithLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	got, err := tool.Call(context.Background(), "go agents")
	if err != nil {
		t.Fatal(err)
	}
	want := "1. First\nhttps://example.com/a\nalpha beta\n\n2. Second\nhttps://example.com/b\n"
	if got != want {
		t.Errorf("result = %q, want %q", got, want)
	}
}

func TestSearchToolOmitsLimitByDefault(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"hits":[]}`,
		func(t *testing.T, r *http.Request, payload map[string]any) {
			if _, ok := payload["limit"]; ok {
				t.Errorf("limit should be omitted, got %v", payload["limit"])
			}
		})
	tool, err := NewSearch(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := tool.Call(context.Background(), "nothing")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "No good Outcrawl search results") {
		t.Errorf("result = %q", got)
	}
}

func TestSearchToolAPIError(t *testing.T) {
	srv := newServer(t, http.StatusUnauthorized, `{"code":"unauthorized","message":"authentication failed"}`, nil)
	tool, err := NewSearch(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tool.Call(context.Background(), "q")
	if !errors.Is(err, ErrAPIError) {
		t.Fatalf("err = %v, want ErrAPIError", err)
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("err = %v, want API message", err)
	}
}

func TestScrapeTool(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"url":"https://example.com/","statusCode":200,"markdown":"# Example\n\nHello"}`,
		func(t *testing.T, r *http.Request, payload map[string]any) {
			if r.URL.Path != "/v1/scrape" {
				t.Errorf("path = %s, want /v1/scrape", r.URL.Path)
			}
			if payload["url"] != "https://example.com" {
				t.Errorf("url = %v", payload["url"])
			}
			formats, _ := payload["formats"].([]any)
			if len(formats) != 1 || formats[0] != "markdown" {
				t.Errorf("formats = %v", payload["formats"])
			}
			if payload["onlyMainContent"] != true {
				t.Errorf("onlyMainContent = %v, want true", payload["onlyMainContent"])
			}
		})
	tool, err := NewScrape(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1/"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := tool.Call(context.Background(), " https://example.com ")
	if err != nil {
		t.Fatal(err)
	}
	if got != "# Example\n\nHello" {
		t.Errorf("result = %q", got)
	}
}

func TestScrapeToolFullPage(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"markdown":"x"}`,
		func(t *testing.T, r *http.Request, payload map[string]any) {
			if payload["onlyMainContent"] != false {
				t.Errorf("onlyMainContent = %v, want false", payload["onlyMainContent"])
			}
		})
	tool, err := NewScrape(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"), WithOnlyMainContent(false))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tool.Call(context.Background(), "https://example.com"); err != nil {
		t.Fatal(err)
	}
}

func TestScrapeToolEmptyMarkdown(t *testing.T) {
	srv := newServer(t, http.StatusOK, `{"markdown":"  "}`, nil)
	tool, err := NewScrape(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tool.Call(context.Background(), "https://example.com")
	if !errors.Is(err, ErrNoResult) {
		t.Fatalf("err = %v, want ErrNoResult", err)
	}
}

func TestScrapeToolHTTPError(t *testing.T) {
	srv := newServer(t, http.StatusBadGateway, `not json`, nil)
	tool, err := NewScrape(WithAPIKey("test-key"), WithBaseURL(srv.URL+"/v1"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = tool.Call(context.Background(), "https://example.com")
	if !errors.Is(err, ErrAPIError) {
		t.Fatalf("err = %v, want ErrAPIError", err)
	}
}

func TestMissingAPIKey(t *testing.T) {
	t.Setenv("OUTCRAWL_API_KEY", "")
	if _, err := NewSearch(); !errors.Is(err, ErrMissingToken) {
		t.Errorf("NewSearch err = %v, want ErrMissingToken", err)
	}
	if _, err := NewScrape(); !errors.Is(err, ErrMissingToken) {
		t.Errorf("NewScrape err = %v, want ErrMissingToken", err)
	}
}

func TestAPIKeyFromEnv(t *testing.T) {
	t.Setenv("OUTCRAWL_API_KEY", "test-key")
	srv := newServer(t, http.StatusOK, `{"hits":[]}`, nil)
	tool, err := NewSearch(WithBaseURL(srv.URL + "/v1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tool.Call(context.Background(), "q"); err != nil {
		t.Fatal(err)
	}
}
