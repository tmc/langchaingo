package redisvector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	"github.com/redis/rueidis"
	"github.com/tmc/langchaingo/schema"
)

// RedisClient interface of redis client, easy to replace third redis client package
// use rueidis temporarily, go-redis has not yet implemented RedisJSON & RediSearch.
type RedisClient interface {
	DropIndex(ctx context.Context, index string, deleteDocuments bool) error
	CheckIndexExists(ctx context.Context, index string) bool
	CreateIndexIfNotExists(ctx context.Context, index string, schema *IndexSchema) error
	AddDocWithHash(ctx context.Context, prefix string, doc schema.Document) (string, error)
	AddDocsWithHash(ctx context.Context, prefix string, docs []schema.Document) ([]string, error)
	// TODO AddDocsWithJSON
	Search(ctx context.Context, search IndexVectorSearch) (int64, []schema.Document, error)
}

type RueidisClient struct {
	client rueidis.Client
}

var _ RedisClient = RueidisClient{}

// NewRueidisClient create rueidis redist client.
func NewRueidisClient(url string) (*RueidisClient, error) {
	clientOption, err := rueidis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	client, err := rueidis.NewClient(clientOption)
	if err != nil {
		return nil, err
	}
	return &RueidisClient{client}, err
}

func (c RueidisClient) DropIndex(ctx context.Context, index string, deleteDocuments bool) error {
	if deleteDocuments {
		return c.client.Do(ctx, c.client.B().FtDropindex().Index(index).Dd().Build()).Error()
	}
	return c.client.Do(ctx, c.client.B().FtDropindex().Index(index).Build()).Error()
}

func (c RueidisClient) CheckIndexExists(ctx context.Context, index string) bool {
	if index == "" {
		return false
	}
	return c.client.Do(ctx, c.client.B().FtInfo().Index(index).Build()).Error() == nil
}

func (c RueidisClient) CreateIndexIfNotExists(ctx context.Context, index string, schema *IndexSchema) error {
	if index == "" {
		return ErrEmptyIndexName
	}

	if c.CheckIndexExists(ctx, index) {
		return nil
	}

	redisIndex := NewIndex(index, []string{getPrefix(index)}, HASHIndexType, *schema)
	createIndexCmd, err := redisIndex.AsCommand()
	if err != nil {
		return err
	}
	return c.client.Do(ctx, c.client.B().Arbitrary(createIndexCmd[0]).Keys(createIndexCmd[1]).Args(createIndexCmd[2:]...).Build()).Error()
}

func (c RueidisClient) AddDocWithHash(ctx context.Context, prefix string, doc schema.Document) (string, error) {
	docID, cmd := c.generateHSetCMD(prefix, doc)
	return docID, c.client.Do(ctx, cmd).Error()
}

func (c RueidisClient) AddDocsWithHash(ctx context.Context, prefix string, docs []schema.Document) ([]string, error) {
	cmds := make([]rueidis.Completed, 0, len(docs))
	docIDs := make([]string, 0, len(docs))
	errs := make([]error, 0, len(docs))
	for _, doc := range docs {
		docID, cmd := c.generateHSetCMD(prefix, doc)
		cmds = append(cmds, cmd)
		docIDs = append(docIDs, docID)
	}
	result := c.client.DoMulti(ctx, cmds...)
	for _, res := range result {
		if res.Error() != nil {
			errs = append(errs, res.Error())
		}
	}
	return docIDs, errors.Join(errs...)
}

func (c RueidisClient) Search(ctx context.Context, search IndexVectorSearch) (int64, []schema.Document, error) {
	cmds := search.AsCommand()
	// fmt.Println(strings.Join(cmds, " "))
	total, docs, err := c.client.Do(ctx, c.client.B().Arbitrary(cmds[0]).Keys(cmds[1]).Args(cmds[2:]...).Build()).AsFtSearch()
	if err != nil {
		return 0, nil, err
	}

	return total, convertFTSearchResIntoDocSchema(docs), nil
}

func (c RueidisClient) generateHSetCMD(prefix string, doc schema.Document) (string, rueidis.Completed) {
	kvs := make([]string, 0, len(doc.Metadata)*2)
	for k, v := range doc.Metadata {
		kvs = append(kvs, k)
		if k == defaultContentVectorFieldKey {
			// convert []float32 into string
			if _v, ok := v.([]float32); ok {
				kvs = append(kvs, VectorString32(_v))
			} else if _v, ok := v.([]float64); ok {
				kvs = append(kvs, VectorString64(_v))
			} else {
				slog.Warn("the type of content vector filed is invalid", "type", reflect.TypeOf(v))
			}
		} else {
			kvs = append(kvs, fmt.Sprintf("%v", v))
		}
	}
	docID := getDocIDWithMetaData(prefix, doc.Metadata)
	return docID, c.client.B().Arbitrary("Hmset").Keys(docID).Args(kvs...).Build()
}

// getPrefix get prefix with index name.
func getPrefix(index string) string {
	return fmt.Sprintf("doc:%s", index)
}

// get doc id with metadata
// same as langchain python version
// if metadata has ids or keys field, return ids or keys value
// if not, return uuid string.
func getDocIDWithMetaData(prefix string, meta map[string]any) string {
	// Get keys or ids from metadata
	// Other vector stores use ids
	if id, ok := meta["ids"]; ok {
		return fmt.Sprintf("%s:%v", prefix, id)
	} else if key, ok := meta["keys"]; ok {
		return fmt.Sprintf("%s:%v", prefix, key)
	}
	return fmt.Sprintf("%s:%v", prefix, uuid.New().String())
}

func convertFTSearchResIntoDocSchema(docs []rueidis.FtSearchDoc) []schema.Document {
	res := make([]schema.Document, 0, len(docs))
	for _, doc := range docs {
		_doc := schema.Document{}
		metadata := make(map[string]any, len(doc.Doc))
		//nolint: gocritic
		for k, v := range doc.Doc {
			if k == defaultContentFieldKey {
				_doc.PageContent = v
			} else if k == defaultDistanceFieldKey {
				score, _ := strconv.ParseFloat(v, 32)
				_doc.Score = float32(score)
			} else if k != defaultContentVectorFieldKey {
				metadata[k] = v
			}
		}
		if _, ok := metadata["id"]; !ok {
			metadata["id"] = doc.Key
		}
		_doc.Metadata = metadata
		res = append(res, _doc)
	}
	return res
}
