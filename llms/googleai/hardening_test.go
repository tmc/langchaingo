package googleai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

type hardeningRoundTripFunc func(*http.Request) (*http.Response, error)

func (f hardeningRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAPIKeyTransport(t *testing.T) {
	tests := []struct {
		name     string
		rawQuery string
		wantKey  string
	}{
		{"missing", "alt=json", "explicit"},
		{"replace", "key=old&alt=json", "explicit"},
		{"deduplicate", "key=one&key=two", "explicit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalURL := &url.URL{Scheme: "https", Host: "example.invalid", Path: "/v1/test", RawQuery: tt.rawQuery}
			req := &http.Request{Method: http.MethodGet, URL: originalURL, Header: http.Header{"X-Goog-Api-Key": {"caller-header"}}}
			client := withAPIKey(&http.Client{Transport: hardeningRoundTripFunc(func(got *http.Request) (*http.Response, error) {
				if keys := got.URL.Query()["key"]; len(keys) != 1 || keys[0] != tt.wantKey {
					t.Fatalf("keys=%q", keys)
				}
				if got.Header.Get("X-Goog-Api-Key") != "caller-header" {
					t.Fatalf("X-Goog-Api-Key=%q", got.Header.Get("X-Goog-Api-Key"))
				}
				return emptyResponse(), nil
			})}, tt.wantKey)
			if _, err := client.Do(req); err != nil {
				t.Fatal(err)
			}
			if originalURL.RawQuery != tt.rawQuery {
				t.Fatalf("original query mutated: %q", originalURL.RawQuery)
			}
			if req.URL != originalURL {
				t.Fatal("original request URL replaced")
			}
		})
	}
}

func TestAuthOptionPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		want    authKind
		wantKey string
	}{
		{"api key", []Option{WithAPIKey("one")}, authAPIKey, "one"},
		{"last API key", []Option{WithAPIKey("one"), WithAPIKey("two")}, authAPIKey, "two"},
		{"credentials after key", []Option{WithAPIKey("one"), WithCredentialsJSON([]byte(`{}`))}, authCredentials, ""},
		{"key after credentials", []Option{WithCredentialsJSON([]byte(`{}`)), WithAPIKey("two")}, authAPIKey, "two"},
		{"empty credentials ignored", []Option{WithAPIKey("one"), WithCredentialsJSON(nil), WithCredentialsFile("")}, authAPIKey, "one"},
		{"explicit empty key", []Option{WithAPIKey("")}, authAPIKey, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := DefaultOptions()
			for _, opt := range tt.opts {
				opt(&o)
			}
			if o.auth != tt.want || o.apiKey != tt.wantKey {
				t.Fatalf("auth=%v key=%q", o.auth, o.apiKey)
			}
		})
	}
}

func TestEnvironmentAPIKeyPrecedence(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "environment")
	tests := []struct {
		name    string
		opts    []Option
		want    authKind
		wantKey string
	}{
		{"fallback", nil, authAPIKey, "environment"},
		{"explicit key", []Option{WithAPIKey("explicit")}, authAPIKey, "explicit"},
		{"explicit empty key", []Option{WithAPIKey("")}, authAPIKey, ""},
		{"credentials", []Option{WithCredentialsJSON([]byte(`{}`))}, authCredentials, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := DefaultOptions()
			for _, opt := range tt.opts {
				opt(&o)
			}
			o.EnsureAuthPresent()
			if o.auth != tt.want || o.apiKey != tt.wantKey {
				t.Fatalf("auth=%v key=%q", o.auth, o.apiKey)
			}
		})
	}
}

func TestEmbeddingModelDefaults(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want string
	}{
		{"implicit", nil, "gemini-embedding-001"},
		{"explicit", []Option{WithDefaultEmbeddingModel("custom")}, "custom"},
		{"explicit empty", []Option{WithDefaultEmbeddingModel("")}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := DefaultOptions()
			for _, opt := range tt.opts {
				opt(&o)
			}
			if got := o.DefaultEmbeddingModel; got != tt.want {
				t.Fatalf("DefaultEmbeddingModel = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGoogleAIEmbeddingRequestModelAndAuthentication(t *testing.T) {
	var requestURL *url.URL
	transport := hardeningRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		u := *req.URL
		requestURL = &u
		body := `{"embeddings":[{"values":[1]}]}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	client, err := New(context.Background(), WithRest(), WithAPIKey("wire-key"), WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	got, err := client.CreateEmbedding(context.Background(), []string{"hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || len(got[0]) != 1 || got[0][0] != 1 {
		t.Fatalf("embedding=%v", got)
	}
	if !strings.Contains(requestURL.Path, "/models/gemini-embedding-001:batchEmbedContents") {
		t.Fatalf("path=%q", requestURL.Path)
	}
	if keys := requestURL.Query()["key"]; len(keys) != 1 || keys[0] != "wire-key" {
		t.Fatalf("keys=%q", keys)
	}
}

func TestCustomEndpointKeepsClientAndAPIKey(t *testing.T) {
	caller := &http.Client{}
	var got *http.Request
	caller.Transport = hardeningRoundTripFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Clone(req.Context())
		body := `{"embeddings":[{"values":[1]}]}`
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	endpoint := func(o *Options) {
		o.ClientOptions = append(o.ClientOptions, option.WithEndpoint("custom.invalid"))
	}
	client, err := New(context.Background(), WithRest(), endpoint, WithHTTPClient(caller), WithAPIKey("explicit"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.CreateEmbedding(context.Background(), []string{"hello"}); err != nil {
		t.Fatal(err)
	}
	if got.URL.Host != "custom.invalid" {
		t.Fatalf("host=%q", got.URL.Host)
	}
	if keys := got.URL.Query()["key"]; len(keys) != 1 || keys[0] != "explicit" {
		t.Fatalf("keys=%q", keys)
	}
	if got.Header.Get("X-Goog-Api-Key") != "" {
		t.Fatal("unexpected API key header")
	}
	if _, ok := caller.Transport.(apiKeyTransport); ok {
		t.Fatal("caller client was mutated")
	}
}

func emptyResponse() *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
}

func TestGenerateContentEmptyStream(t *testing.T) {
	t.Parallel()
	// Gemini can end a stream without sending any response, for example
	// when thinking uses the whole output token budget.
	single := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "Say OK"),
	}
	chat := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "Be brief."),
		llms.TextParts(llms.ChatMessageTypeHuman, "Say OK"),
	}
	stream := llms.WithStreamingFunc(func(context.Context, []byte) error { return nil })
	tests := []struct {
		name     string
		messages []llms.MessageContent
		opts     []llms.CallOption
	}{
		{"chat", chat, nil},
		{"single streaming", single, []llms.CallOption{stream}},
		{"chat streaming", chat, []llms.CallOption{stream}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transport := hardeningRoundTripFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("[]")), Request: req}, nil
			})
			client, err := New(context.Background(), WithRest(), WithAPIKey("wire-key"), WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			opts := append([]llms.CallOption{llms.WithMaxTokens(10)}, tt.opts...)
			_, err = client.GenerateContent(context.Background(), tt.messages, opts...)
			if !errors.Is(err, ErrNoContentInResponse) {
				t.Fatalf("err=%v, want %v", err, ErrNoContentInResponse)
			}
		})
	}
}
