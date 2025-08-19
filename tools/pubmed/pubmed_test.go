// File: tools/pubmed/pubmed_test.go

package pubmed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/tools/pubmed/internal"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tool, err := New(5, "TestUserAgent")
	require.NoError(t, err)
	assert.NotNil(t, tool)
	assert.Equal(t, "PubMed Search", tool.Name())
}

func TestPubMedTool(t *testing.T) {
	t.Parallel()

	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request URL is correct
		assert.Contains(t, r.URL.String(), "/eutils/")

		// Serve a mock XML response
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		if r.URL.Path == "/eutils/esearch.fcgi" {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
			<eSearchResult>
				<IdList>
					<Id>12345</Id>
				</IdList>
				<WebEnv>NCID_1_1234567_130.14.22.215_9001_1234567890_123456789_0MetA0_S_MegaStore</WebEnv>
				<QueryKey>1</QueryKey>
			</eSearchResult>`))
		} else if r.URL.Path == "/eutils/efetch.fcgi" {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
			<PubmedArticleSet>
				<PubmedArticle>
					<MedlineCitation>
						<PMID>12345</PMID>
						<Article>
							<ArticleTitle>Mock PubMed Article</ArticleTitle>
							<Abstract>
								<AbstractText>This is a mock abstract of a PubMed article.</AbstractText>
							</Abstract>
							<AuthorList>
								<Author>
									<LastName>Doe</LastName>
									<ForeName>John</ForeName>
									<Initials>J</Initials>
								</Author>
							</AuthorList>
						</Article>
					</MedlineCitation>
					<PubmedData>
						<History>
							<PubMedPubDate PubStatus="pubmed">
								<Year>2023</Year>
								<Month>1</Month>
								<Day>1</Day>
							</PubMedPubDate>
						</History>
					</PubmedData>
				</PubmedArticle>
			</PubmedArticleSet>`))
		}
	}))
	defer server.Close()

	// Create a custom client that uses the test server URL
	customClient := &internal.Client{
		MaxResults: 1,
		UserAgent:  "TestAgent",
		BaseURL:    server.URL + "/eutils",
	}

	// Create the PubMed tool with the custom client
	tool := &Tool{
		client: customClient,
	}

	// Test the Call method
	result, err := tool.Call(context.Background(), "test query")

	// Assert that there's no error
	require.NoError(t, err)

	// Assert that the result contains expected information
	assert.Contains(t, result, "Title: Mock PubMed Article")
	assert.Contains(t, result, "Authors: John Doe")
	assert.Contains(t, result, "Abstract: This is a mock abstract of a PubMed article.")
	assert.Contains(t, result, "PMID: 12345")
	assert.Contains(t, result, "Published: 2023-01-01")
}

func TestPubMedToolNoResults(t *testing.T) {
	t.Parallel()

	// Create a mock HTTP server that returns no results
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
		<eSearchResult>
			<IdList></IdList>
		</eSearchResult>`))
	}))
	defer server.Close()

	customClient := &internal.Client{
		MaxResults: 1,
		UserAgent:  "TestAgent",
		BaseURL:    server.URL + "/eutils",
	}

	tool := &Tool{
		client: customClient,
	}

	result, err := tool.Call(context.Background(), "no results query")

	require.NoError(t, err)
	require.Equal(t, "No good PubMed Search Results were found", result)
}

func TestPubMedToolAPIError(t *testing.T) {
	t.Parallel()

	// Create a mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	customClient := &internal.Client{
		MaxResults: 1,
		UserAgent:  "TestAgent",
		BaseURL:    server.URL + "/eutils",
	}

	tool := &Tool{
		client: customClient,
	}

	_, err := tool.Call(context.Background(), "error query")

	require.Error(t, err)
	require.Equal(t, internal.ErrAPIResponse, err)
}
