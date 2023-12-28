package azureaisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type QueryType string

var (
	QueryType_Simple   QueryType = "simple"
	QueryType_Full     QueryType = "full"
	QueryType_Semantic QueryType = "semantic"
)

type QueryCaptions string

var (
	QueryType_Extractive QueryCaptions = "extractive"
	QueryType_None       QueryCaptions = "none"
)

type SpellerType string

var (
	SpellerType_Lexicon SpellerType = "lexicon"
	SpellerType_None    SpellerType = "none"
)

type SearchDocumentsRequestInput struct {
	Count                 bool                                `json:"count,omitempty"`
	Captions              QueryCaptions                       `json:"captions,omitempty"`
	Facets                []string                            `json:"facets,omitempty"`
	Filter                string                              `json:"filter,omitempty"`
	Highlight             string                              `json:"highlight,omitempty"`
	HighlightPostTag      string                              `json:"highlightPostTag,omitempty"`
	HighlightPreTag       string                              `json:"highlightPreTag,omitempty"`
	MinimumCoverage       int16                               `json:"minimumCoverage,omitempty"`
	Orderby               string                              `json:"orderby,omitempty"`
	QueryType             QueryType                           `json:"queryType,omitempty"`
	QueryLanguage         string                              `json:"queryLanguage,omitempty"`
	Speller               SpellerType                         `json:"speller,omitempty"`
	SemanticConfiguration string                              `json:"semanticConfiguration,omitempty"`
	ScoringParameters     []string                            `json:"scoringParameters,omitempty"`
	ScoringProfile        string                              `json:"scoringProfile,omitempty"`
	Search                string                              `json:"search,omitempty"`
	SearchFields          string                              `json:"searchFields,omitempty"`
	SearchMode            string                              `json:"searchMode,omitempty"`
	SessionID             string                              `json:"sessionId,omitempty"`
	ScoringStatistics     string                              `json:"scoringStatistics,omitempty"`
	Select                string                              `json:"select,omitempty"`
	Skip                  int                                 `json:"skip,omitempty"`
	Top                   int                                 `json:"top,omitempty"`
	Vectors               []SearchDocumentsRequestInputVector `json:"vectors,omitempty"`
	VectorFilterMode      string                              `json:"vectorFilterMode,omitempty"`
}

type SearchDocumentsRequestInputVector struct {
	Kind       string    `json:"kind,omitempty"`
	Value      []float32 `json:"value,omitempty"`
	Fields     string    `json:"fields,omitempty"`
	K          int       `json:"k,omitempty"`
	Exhaustive bool      `json:"exhaustive,omitempty"`
}

type SearchDocumentsRequestOuput struct {
	OdataCount   int `json:"@odata.count,omitempty"`
	SearchFacets struct {
		Category []struct {
			Count int    `json:"count,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"category,omitempty"`
	} `json:"@search.facets,omitempty"`
	SearchNextPageParameters SearchDocumentsRequestInput `json:"@search.nextPageParameters,omitempty"`
	Value                    []map[string]interface{}    `json:"value,omitempty"`
	OdataNextLink            string                      `json:"@odata.nextLink,omitempty"`
}

func (s *Store) SearchDocuments(ctx context.Context, indexName string, payload SearchDocumentsRequestInput, output *SearchDocumentsRequestOuput) error {
	URL := fmt.Sprintf("%s/indexes/%s/docs/search?api-version=2023-07-01-Preview", s.cognitiveSearchEndpoint, indexName)
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("err marshalling document for cognitive search: %v\n", err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, URL, bytes.NewBuffer(body))

	if err != nil {
		fmt.Printf("err setting request for cognitive search document: %v\n", err)
		return err
	}

	req.Header.Add("content-Type", "application/json")
	if s.cognitiveSearchAPIKey != "" {
		req.Header.Add("api-key", s.cognitiveSearchAPIKey)
	}
	return s.HTTPDefaultSend(req, "search documents on cognitive search", output)

}
