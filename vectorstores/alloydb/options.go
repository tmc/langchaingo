package alloydb

import "context"

const defaultPort = 5432 // default PostgreSQL port

// Option is a function type that can be used to modify the Engine.
type Option func(p *engineConfig)

type engineConfig struct {
	projectID       string
	region          string
	cluster         string
	instance        string
	host            string
	port            int
	database        string
	user            string
	password        string
	ipType          string
	iAmAccountEmail string
	emailRetreiver  EmailRetreiver
}

// WithProjectID sets the ProjectId field
func WithProjectID(projectID string) Option {
	return func(p *engineConfig) {
		p.projectID = projectID
	}
}

// WithRegion sets the Region field
func WithRegion(region string) Option {
	return func(p *engineConfig) {
		p.region = region
	}
}

// WithCluster sets the Cluster field
func WithCluster(cluster string) Option {
	return func(p *engineConfig) {
		p.cluster = cluster
	}
}

// WithInstance sets the Instance field
func WithInstance(instance string) Option {
	return func(p *engineConfig) {
		p.instance = instance
	}
}

// WithHost sets the Host field
func WithHost(host string) Option {
	return func(p *engineConfig) {
		p.host = host
	}
}

// WithPort sets the Port field
func WithPort(port int) Option {
	return func(p *engineConfig) {
		p.port = port
	}
}

// WithDatabase sets the Database field
func WithDatabase(database string) Option {
	return func(p *engineConfig) {
		p.database = database
	}
}

// WithUser sets the User field
func WithUser(user string) Option {
	return func(p *engineConfig) {
		p.user = user
	}
}

// WithPassword sets the Password field
func WithPassword(password string) Option {
	return func(p *engineConfig) {
		p.password = password
	}
}

// WithIpType sets the IpType field
func WithIpType(ipType string) Option {
	return func(p *engineConfig) {
		p.ipType = ipType
	}
}

// WithIAMAccountEmail sets the IAMAccountEmail field
func WithIAMAccountEmail(email string) Option {
	return func(p *engineConfig) {
		p.iAmAccountEmail = email
	}
}

// WithServiceAccountRetriever sets the ServiceAccountRetriever field
func WithServiceAccountRetriever(emailRetriever func(context.Context) (string, error)) Option {
	return func(p *engineConfig) {
		p.emailRetreiver = emailRetriever
	}
}
