package vearch

import (
	"context"
	"net/url"
	"log" 
    "errors"
    "fmt"
	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
    client "github.com/vearch/vearch/v3/sdk/go/vearch"
    "github.com/vearch/vearch/v3/sdk/go/vearch/auth"
    "github.com/vearch/vearch/v3/sdk/go/vearch/entities/models"
    "github.com/vearch/vearch/v3/sdk/go/vearch/data"
)

type Store struct {
	DbName     string
	SpaceName  string
	ClusterUrl url.URL
	embedder   embeddings.Embedder
	
}

var _ vectorstores.VectorStore = Store{}

func setupClient(url string) *client.Client {
	host :=  url// router url
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
	options ...vectorstores.Option,
) ([]string, error) {
    
    c := setupClient(s.ClusterUrl.String())
	texts := make([]string, 0, len(docs))
	for _, doc := range docs {
		texts = append(texts, doc.PageContent)
	}
    vectors,
		err := s.embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}

    if len(vectors) != len(docs) {
		return nil, errors.New("number of vectors from embedder does not match number of documents")
	}
    documents := make([]interface{}, 0, len(docs))
    for i, doc := range docs {
        
        document := map[string]interface{}{
            "_id":          uuid.New().String(),
            "PageContent":  doc.PageContent,
            "vec":          vectors[i],
        }
        for key, value := range doc.Metadata {
			document[key] = value
		}
        documents = append(documents, document)
    }
    resp, err :=c.Data().Creator().WithDBName(s.DbName).WithSpaceName(s.SpaceName).WithDocs(documents).Do(ctx)    
    if err != nil {
        log.Fatal(err)
    }
    doc_ids := make([]string, 0, resp.Docs.Data.Total)
    if resp.Docs.Code == 0{
        for _,id :=range resp.Docs.Data.DocumentIds{
          doc_ids =append(doc_ids,id.ID)
       }
    }
    return doc_ids,err
}

func (s *Store) SimilaritySearch(
	ctx context.Context,
	query string,
	numDocuments int,
	options ...vectorstores.Option,
) ([]schema.Document, error) {
    
    opts := s.getOptions(options...)
    filters_input := s.getFilters(opts)
    c := setupClient(s.ClusterUrl.String())

	vector,
		err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}
    
    var resp *data.SearchWrapper
    var err error
    if filters_input !=nil{
        filters := &models.Filters{}
        for operator, conditions := range filters_input.(map[string]interface{}){
            filters.Operator = operator
            for _, condMap := range conditions.([]map[string]interface{}) {
                conditionInterface, ok := condMap["condition"].(map[string]interface{})
                if !ok {
                    fmt.Println("Expected condMap['condition'] to be a map[string]interface{}")
                    continue
                }
                condition := models.Condition{
                    Field:    conditionInterface["Field"].(string),
                    Operator: conditionInterface["Operator"].(string),
                    Value:    conditionInterface["Value"],
                        }
                filters.Conditions = append(filters.Conditions, condition)
            }
        }
        
        resp, err = c.Data().Searcher().WithDBName(s.DbName).WithSpaceName(s.SpaceName).WithLimit(numDocuments).WithVectors(vectors).WithFilters(filters).Do(ctx)

    }else{
        resp, err = c.Data().Searcher().WithDBName(s.DbName).WithSpaceName(s.SpaceName).WithLimit(numDocuments).WithVectors(vectors).Do(ctx)

    }
   
    if err != nil {
		panic(err)
	}
    var documents []schema.Document
	for _, item := range resp.Docs.Data.Documents {
		for _, docItem := range item.([]interface{}) { 
			docMap := docItem.(map[string]interface{}) 

			metadata := make(map[string]any)
		
			for key, value := range docMap {
				if key != "PageContent" && key != "_score" && key != "_id" {
					metadata[key] = value
				}
			}

			pageContent, _ := docMap["PageContent"].(string) 
			score, _ := docMap["_score"].(float64)           
			doc := schema.Document{
				PageContent: pageContent,
				Metadata:    metadata,
				Score:       float32(score), 
			}
			documents = append(documents, doc)
		}
	}   
    return documents,nil
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