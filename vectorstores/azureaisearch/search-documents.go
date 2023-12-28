package azureaisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type QueryType string

const (
	QueryTypeSimple   QueryType = "simple"
	QueryTypeFull     QueryType = "full"
	QueryTypeSemantic QueryType = "semantic"
)

type QueryCaptions string

const (
	QueryTypeExtractive QueryCaptions = "extractive"
	QueryTypeNone       QueryCaptions = "none"
)

type SpellerType string

const (
	SpellerTypeLexicon SpellerType = "lexicon"
	SpellerTypeNone    SpellerType = "none"
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

func (s *Store) SearchDocuments(
	ctx context.Context,
	indexName string,
	payload SearchDocumentsRequestInput,
	output *SearchDocumentsRequestOuput,
) error {
	URL := fmt.Sprintf("%s/indexes/%s/docs/search?api-version=2023-07-01-Preview", s.azureAISearchEndpoint, indexName)
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("err marshalling document for azure ai search: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("err setting request for azure ai search document: %w", err)
	}

	req.Header.Add("content-Type", "application/json")
	if s.azureAISearchAPIKey != "" {
		req.Header.Add("api-key", s.azureAISearchAPIKey)
	}
	return s.HTTPDefaultSend(req, "search documents on azure ai search", output)
}
