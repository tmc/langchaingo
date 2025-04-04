package redisvector

import (
	"errors"
	"fmt"

	"github.com/averikitsch/langchaingo/embeddings"
)

type RedisIndexAlgorithm string

// ErrInvalidOptions is returned when the options given are invalid.
var ErrInvalidOptions = errors.New("invalid options")

// Option is a function type that can be used to modify the client.
type Option func(s *Store)

// WithEmbedder is an option for setting the embedder to use. Must be set.
func WithEmbedder(e embeddings.Embedder) Option {
	return func(s *Store) {
		s.embedder = e
	}
}

// WithConnectionURL is an option for specifying the Redis connection URL. Must be set.
// URL meets the official redis url format (https://github.com/redis/redis-specifications/blob/master/uri/redis.txt)
// Example:
//
// redis://<user>:<password>@<host>:<port>/<db_number>
// redis://<user>:<password>@<host>:<port>?addr=<host2>:<port2>&addr=<host3>:<port3>
// unix://<user>:<password>@</path/to/redis.sock>?db=<db_number>
func WithConnectionURL(connectionURL string) Option {
	return func(s *Store) {
		s.redisURL = connectionURL
	}
}

// WithIndexName is an option for specifying the index name. Must be set.
//
// `createIndexIfNotExists`:
//
//	if set false, will throw error when the index does not exist
//	if set true, will create index when the index does not exist
//
// If the `WithIndexSchema` option is set, the index will be created with this index schema,
// otherwise, the index will be created with the generated schema with document metadata in `AddDocuments`.
func WithIndexName(name string, createIndexIfNotExists bool) Option {
	return func(s *Store) {
		s.indexName = name
		s.createIndexIfNotExists = createIndexIfNotExists
	}
}

// SchemaFormat JSONSchemaFormat or YAMLSchemaFormat.
type SchemaFormat string

const (
	JSONSchemaFormat SchemaFormat = "JSON"
	YAMLSchemaFormat SchemaFormat = "YAML"
)

// WithIndexSchema is an option for specifying the index schema with file or bytes
//
//	`format`: support YAML & JSON format
//	`schemaFilePath`: schema config file path
//	`schemaBytes`: schema string
//
// throw error if schemaFilePath & schemaBytes are both empty
// the schemaBytes will overwrite the schemaFilePath if schemaFilePath & schemaBytes both set
// if index doesn't exist, the schema will be used to create index
// otherwise, it only control what fields the metadata maps to return in search result
// ref: https://python.langchain.com/docs/integrations/vectorstores/redis/#custom-metadata-indexing
func WithIndexSchema(format SchemaFormat, schemaFilePath string, schemaBytes []byte) Option {
	return func(s *Store) {
		s.schemaGenerator = &schemaGenerator{
			format:   format,
			filePath: schemaFilePath,
			buf:      schemaBytes,
		}
	}
}

func applyClientOptions(opts ...Option) (*Store, error) {
	s := &Store{}

	for _, opt := range opts {
		opt(s)
	}

	if s.embedder == nil {
		return nil, fmt.Errorf("%w: missing embedder", ErrInvalidOptions)
	}

	if s.indexName == "" {
		return nil, fmt.Errorf("%w: missing index name", ErrInvalidOptions)
	}

	if s.schemaGenerator != nil {
		schema, err := s.schemaGenerator.generate()
		if err != nil {
			return nil, err
		}
		s.indexSchema = schema
		// clear generator buf
		s.schemaGenerator = nil
	}

	return s, nil
}
