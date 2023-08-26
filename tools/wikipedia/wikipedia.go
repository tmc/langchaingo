package wikipedia

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/tools"
)

const (
	_defaultTopK         = 2
	_defaultDocMaxChars  = 2000
	_defaultLanguageCode = "en"
)

// ErrUnexpectedAPIResult is returned if the result form the wikipedia api is unexpected.
var ErrUnexpectedAPIResult = errors.New("unexpected result from wikipedia api")

// Tool is an implementation of the tool interface that finds information using the wikipedia api.
type Tool struct {
	CallbacksHandler callbacks.Handler
	// The number of wikipedia pages to include in the result.
	TopK int
	// The number of characters to take from each page.
	DocMaxChars int
	// The language code to use.
	LanguageCode string
	// The user agent sent in the heder. See https://www.mediawiki.org/wiki/API:Etiquette.
	UserAgent string
}

var _ tools.Tool = Tool{}

// New creates a new wikipedia tool to find wikipedia pages using the wikipedia api. TopK is set
// to 2, DocMaxChars is set to 2000 and the language code is set to "en".
func New(userAgent string) Tool {
	return Tool{
		TopK:         _defaultTopK,
		DocMaxChars:  _defaultDocMaxChars,
		LanguageCode: _defaultLanguageCode,
		UserAgent:    userAgent,
	}
}

func (t Tool) Name() string {
	return "Wikipedia"
}

func (t Tool) Description() string {
	return `
	A wrapper around Wikipedia. 
	Useful for when you need to answer general questions about 
	people, places, companies, facts, historical events, or other subjects. 
	Input should be a search query.`
}

// Call uses the wikipedia api to find the top search results for the input and returns
// the first part of the documents combined.
func (t Tool) Call(ctx context.Context, input string) (string, error) {
	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolStart(ctx, input)
	}

	searchResult, err := search(ctx, t.TopK, input, t.LanguageCode, t.UserAgent)
	if err != nil {
		return "", err
	}

	if len(searchResult.Query.Search) == 0 {
		return "no wikipedia pages found", nil
	}

	result := ""

	for _, search := range searchResult.Query.Search {
		getPageResult, err := getPage(ctx, search.PageID, t.LanguageCode, t.UserAgent)
		if err != nil {
			return "", err
		}

		page, ok := getPageResult.Query.Pages[fmt.Sprintf("%v", search.PageID)]
		if !ok {
			return "", ErrUnexpectedAPIResult
		}
		if len(page.Extract) >= t.DocMaxChars {
			result += page.Extract[0:t.DocMaxChars]
			continue
		}
		result += page.Extract
	}

	if t.CallbacksHandler != nil {
		t.CallbacksHandler.HandleToolEnd(ctx, result)
	}

	return result, nil
}
