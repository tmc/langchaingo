package langsmith

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

var ErrInvalidUUID = errors.New("invalid UUID")

var ErrMissingAPIKey = errors.New("the LangSmith API key was not defined, define LANGCHAIN_API_KEY environment variable or configure your client with WithAPIKey()")

type LangSmitAPIError struct {
	StatusCode int
	URL        string
	Body       []byte
}

func NewLangSmitAPIErrorFromHTTP(req *http.Request, resp *http.Response) *LangSmitAPIError {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte(fmt.Sprintf("<We were unable to read HTTP body, got error %q>", err))
	}

	return &LangSmitAPIError{
		StatusCode: resp.StatusCode,
		URL:        req.URL.String(),
		Body:       body,
	}
}

func (e *LangSmitAPIError) Error() string {
	return fmt.Sprintf("LangSmith API error on %q received status %d: %s", e.URL, e.StatusCode, string(e.Body))
}
