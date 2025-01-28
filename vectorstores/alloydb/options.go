package alloydb

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Option is a function type that can be used to modify the Engine.
type Option func(p *engineConfig)

type engineConfig struct {
	projectID       string
	region          string
	cluster         string
	instance        string
	connPool        *pgxpool.Pool
	database        string
	user            string
	password        string
	ipType          string
	iAmAccountEmail string
	emailRetreiver  EmailRetreiver
}

// WithAlloyDBInstance sets the project, region, cluster, and instance fields.
func WithAlloyDBInstance(projectID, region, cluster, instance string) Option {
	return func(p *engineConfig) {
		p.projectID = projectID
		p.region = region
		p.cluster = cluster
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
		p.iAmAccountEmail = email
	}
}

// withServiceAccountRetriever sets the ServiceAccountRetriever field.
func withServiceAccountRetriever(emailRetriever func(context.Context) (string, error)) Option {
	return func(p *engineConfig) {
		p.emailRetreiver = emailRetriever
	}
}

func applyClientOptions(opts ...Option) (engineConfig, error) {
	cfg := &engineConfig{
		emailRetreiver: getServiceAccountEmail,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.connPool == nil && cfg.projectID == "" && cfg.region == "" && cfg.cluster == "" && cfg.instance == "" {
		return engineConfig{}, errors.New("missing connection: provide a connection pool or connection fields")
	}

	return *cfg, nil
}
