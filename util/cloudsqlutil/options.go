package cloudsqlutil

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultSchemaName = "public"
	defaultUserAgent  = "langchaingo-cloud-sql-pg/0.0.0"
)

// Option is a function type that can be used to modify the Engine.
type Option func(p *engineConfig)

type engineConfig struct {
	projectID       string
	region          string
	instance        string
	connPool        *pgxpool.Pool
	database        string
	user            string
	password        string
	ipType          string
	iamAccountEmail string
	emailRetriever  EmailRetriever
	userAgents      string
}

// VectorstoreTableOptions is used with the InitVectorstoreTable to use the required and default fields.
type VectorstoreTableOptions struct {
	TableName          string
	VectorSize         int
	SchemaName         string
	ContentColumnName  string
	EmbeddingColumn    string
	MetadataJSONColumn string
	IDColumn           Column
	MetadataColumns    []Column
	OverwriteExisting  bool
	StoreMetadata      bool
}

// WithCloudSQLInstance sets the project, region, and instance fields.
func WithCloudSQLInstance(projectID, region, instance string) Option {
	return func(p *engineConfig) {
		p.projectID = projectID
		p.region = region
		p.instance = instance
	}
}

// WithPool sets the Port field.
func WithPool(pool *pgxpool.Pool) Option {
	return func(p *engineConfig) {
		p.connPool = pool
	}
}

// WithDatabase sets the Database field.
func WithDatabase(database string) Option {
	return func(p *engineConfig) {
		p.database = database
	}
}

// WithUser sets the User field.
func WithUser(user string) Option {
	return func(p *engineConfig) {
		p.user = user
	}
}

// WithPassword sets the Password field.
func WithPassword(password string) Option {
	return func(p *engineConfig) {
		p.password = password
	}
}

// WithIPType sets the IpType field.
func WithIPType(ipType string) Option {
	return func(p *engineConfig) {
		p.ipType = ipType
	}
}

// WithIAMAccountEmail sets the IAMAccountEmail field.
func WithIAMAccountEmail(email string) Option {
	return func(p *engineConfig) {
		p.iamAccountEmail = email
	}
}

func applyClientOptions(opts ...Option) (engineConfig, error) {
	cfg := &engineConfig{
		emailRetriever: getServiceAccountEmail,
		ipType:         "PUBLIC",
		userAgents:     defaultUserAgent,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.connPool == nil && cfg.projectID == "" && cfg.region == "" && cfg.instance == "" {
		return engineConfig{}, errors.New("missing connection: provide a connection pool or connection fields")
	}

	return *cfg, nil
}

// Option function type.
type OptionInitChatHistoryTable func(*InitChatHistoryTableOptions)

// Option type for defining options.
type InitChatHistoryTableOptions struct {
	schemaName string
}

// WithSchemaName sets a custom schema name.
func WithSchemaName(schemaName string) OptionInitChatHistoryTable {
	return func(i *InitChatHistoryTableOptions) {
		i.schemaName = schemaName
	}
}

// applyChatMessageHistoryOptions applies the given options to the
// ChatMessageHistory.
func applyChatMessageHistoryOptions(opts ...OptionInitChatHistoryTable) InitChatHistoryTableOptions {
	cfg := &InitChatHistoryTableOptions{
		schemaName: defaultSchemaName,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return *cfg
}
