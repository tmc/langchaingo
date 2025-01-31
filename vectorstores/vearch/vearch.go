package vearch

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	client "github.com/vearch/vearch/v3/sdk/go"
	"github.com/vearch/vearch/v3/sdk/go/auth"
	"github.com/vearch/vearch/v3/sdk/go/data"
	"github.com/vearch/vearch/v3/sdk/go/entities/models"
)

type Store struct {
	DBName     string
	SpaceName  string
	ClusterURL url.URL
	embedder   embeddings.Embedder
}

func setupClient(url string) *client.Client {
	host := url // router url
	user := "root"
	secret := "secret"
	authConfig := auth.BasicAuth{UserName: user, Secret: secret}
	c, err := client.NewClient(client.Config{Host: host, AuthConfig: authConfig})
	if err != nil {
		panic(err)
	}
	return c
}

func New(opts ...Option) (Store, error) {
	s, err := applyClientOptions(opts...)
	if err != nil {
		return Store{}, err
	}
	return s, nil
}

func (s Store) AddDocuments(
	ctx context.Context,
	docs []schema.Document,
) ([]string, error) {
	c := setupClient(s.ClusterURL.String())
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
	vectors, err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(docs) {
		return nil, errors.New("number of vectors from embedder does not match number of documents")
	}
	documents := make([]interface{}, 0, len(docs))
	for i, doc := range docs {
		document := map[string]interface{}{
			"_id":         uuid.New().String(),
			"PageContent": doc.PageContent,
			"vec":         vectors[i],
		}
		for key, value := range doc.Metadata {
			document[key] = value
		}
		documents = append(documents, document)
	}
	resp, err := c.Data().Creator().WithDBName(s.DBName).WithSpaceName(s.SpaceName).WithDocs(documents).Do(ctx)
	if err != nil {
		log.Fatal(err)
	}
	docIDs := make([]string, 0, resp.Docs.Data.Total)
	if resp.Docs.Code == 0 {
		for _, id := range resp.Docs.Data.DocumentIds {
			docIDs = append(docIDs, id.ID)
		}
	}
	return docIDs, err
}

func (s *Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
	opts := s.getOptions(options...)
	filtersInput := s.getFilters(opts)
	c := setupClient(s.ClusterURL.String())

	vector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	vectors := []models.Vector{
		{
			Field:   "vec",
			Feature: vector,
		},
	}

	resp, err := s.executeSearch(ctx, c, vectors, filtersInput, numDocuments)
	if err != nil {
		return nil, err
	}

	return s.parseSearchResponse(resp), nil
}

func (s *Store) executeSearch(
	ctx context.Context,
	c *client.Client,
	vectors []models.Vector,
	filtersInput any,
	numDocuments int,
) (*data.SearchWrapper, error) {
	if filtersInput != nil {
		filters, err := s.buildFilters(filtersInput)
		if err != nil {
			return nil, err
		}
		return s.searchWithFilters(ctx, c, vectors, filters, numDocuments)
	}
	return s.searchWithoutFilters(ctx, c, vectors, numDocuments)
}

func (s *Store) searchWithFilters(
	ctx context.Context,
	c *client.Client,
	vectors []models.Vector,
	filters *models.Filters,
	numDocuments int,
) (*data.SearchWrapper, error) {
	return c.Data().Searcher().
		WithDBName(s.DBName).
		WithSpaceName(s.SpaceName).
		WithLimit(numDocuments).
		WithVectors(vectors).
		WithFilters(filters).
		Do(ctx)
}

func (s *Store) searchWithoutFilters(
	ctx context.Context,
	c *client.Client,
	vectors []models.Vector,
	numDocuments int,
) (*data.SearchWrapper, error) {
	return c.Data().Searcher().
		WithDBName(s.DBName).
		WithSpaceName(s.SpaceName).
		WithLimit(numDocuments).
		WithVectors(vectors).
		Do(ctx)
}

func (s *Store) buildFilters(filtersInput any) (*models.Filters, error) {
	filtersMap, ok := filtersInput.(map[string]interface{})
	if !ok {
		return nil, errors.New("filtersInput must be a map[string]interface{}")
	}
	filters := &models.Filters{}
	for operator, conditions := range filtersMap {
		filters.Operator = operator
		conditionsList, ok := conditions.([]interface{})
		if !ok {
			return nil, errors.New("conditions must be a []interface{}")
		}
		for _, cond := range conditionsList {
			condMap, ok := cond.(map[string]interface{})
			if !ok {
				return nil, errors.New("each condition must be a map[string]interface{}")
			}
			conditionInterface, ok := condMap["condition"].(map[string]interface{})
			if !ok {
				return nil, errors.New("each condition must be a map[string]interface{}")
			}
			field, ok := conditionInterface["Field"].(string)
			if !ok {
				return nil, errors.New("field must be a string")
			}
			operator, ok := conditionInterface["Operator"].(string)
			if !ok {
				return nil, errors.New("operator must be a string")
			}
			value := conditionInterface["Value"]
			condition := models.Condition{
				Field:    field,
				Operator: operator,
				Value:    value,
			}
			filters.Conditions = append(filters.Conditions, condition)
		}
	}
	return filters, nil
}

func (s *Store) parseSearchResponse(resp *data.SearchWrapper) []schema.Document {
	var documents []schema.Document
	for _, item := range resp.Docs.Data.Documents {
		docList, ok := item.([]interface{})
		if !ok {
			continue
		}
		for _, docItem := range docList {
			docMap, ok := docItem.(map[string]interface{})
			if !ok {
				continue
			}
			documents = append(documents, s.createDocumentFromMap(docMap))
		}
	}
	return documents
}

func (s *Store) createDocumentFromMap(docMap map[string]interface{}) schema.Document {
	metadata := make(map[string]any)
	for key, value := range docMap {
		if key != "PageContent" && key != "_score" && key != "_id" {
			metadata[key] = value
		}
	}
	pageContent, _ := docMap["PageContent"].(string)
	score, _ := docMap["_score"].(float64)
	return schema.Document{
		PageContent: pageContent,
		Metadata:    metadata,
		Score:       float32(score),
	}
}

func (s Store) getOptions(options ...vectorstores.Option) vectorstores.Options {
	opts := vectorstores.Options{}
	for _, opt := range options {
		opt(&opts)
	}
	return opts
}

func (s Store) getFilters(opts vectorstores.Options) any {
	if opts.Filters != nil {
		return opts.Filters
	}
	return nil
}
