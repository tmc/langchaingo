package alloydb

const defaultPort = 5432 // default PostgreSQL port

// Option is a function type that can be used to modify the client.
type Option func(p *PostgresEngineConfig)

// WithProjectId sets the ProjectId field
func WithProjectId(projectId string) Option {
	return func(p *PostgresEngineConfig) {
		p.projectId = projectId
	}
}

// WithRegion sets the Region field
func WithRegion(region string) Option {
	return func(p *PostgresEngineConfig) {
		p.region = region
	}
}

// WithCluster sets the Cluster field
func WithCluster(cluster string) Option {
	return func(p *PostgresEngineConfig) {
		p.cluster = cluster
	}
}

// WithInstance sets the Instance field
func WithInstance(instance string) Option {
	return func(p *PostgresEngineConfig) {
		p.instance = instance
	}
}

// WithHost sets the Host field
func WithHost(host string) Option {
	return func(p *PostgresEngineConfig) {
		p.host = host
	}
}

// WithPort sets the Port field
func WithPort(port int) Option {
	return func(p *PostgresEngineConfig) {
		p.port = port
	}
}

// WithDatabase sets the Database field
func WithDatabase(database string) Option {
	return func(p *PostgresEngineConfig) {
		p.database = database
	}
}

// WithUser sets the User field
func WithUser(user string) Option {
	return func(p *PostgresEngineConfig) {
		p.user = user
	}
}

// WithPassword sets the Password field
func WithPassword(password string) Option {
	return func(p *PostgresEngineConfig) {
		p.password = password
	}
}

// WithIpType sets the IpType field
func WithIpType(ipType string) Option {
	return func(p *PostgresEngineConfig) {
		p.ipType = ipType
	}
}

// WithIAMAccountEmail sets the IAMAccountEmail field
func WithIAMAccountEmail(email string) Option {
	return func(p *PostgresEngineConfig) {
		p.iAmAccountEmail = email
	}
}

// WithServiceAccountRetriever sets the ServiceAccountRetriever field
func WithServiceAccountRetriever(retrievedServiceAccount serviceAccountRetriever) Option {
	return func(p *PostgresEngineConfig) {
		p.serviceAccountRetriever = retrievedServiceAccount
	}
}

// applyClientOptions initializes a new PostgresEngineConfig with functional options
func applyClientOptions(options ...Option) *PostgresEngineConfig {
	newPostgresEngineConfig := &PostgresEngineConfig{
		port: defaultPort,
	}
	for _, o := range options {
		o(newPostgresEngineConfig)
	}
	// TODO :: add parameters that will cause failure if missing.

	return newPostgresEngineConfig
}
