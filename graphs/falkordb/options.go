package falkordb

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring FalkorDB.
type Option func(*FalkorDB)

// WithHost sets the FalkorDB host address.
func WithHost(host string) Option {
	return func(f *FalkorDB) {
		f.host = host
	}
}

// WithPort sets the FalkorDB port.
func WithPort(port int) Option {
	return func(f *FalkorDB) {
		f.port = port
	}
}

// WithCredentials sets the username and password for authentication.
func WithCredentials(username, password string) Option {
	return func(f *FalkorDB) {
		f.username = username
		f.password = password
	}
}

// WithSSL enables or disables SSL/TLS connection.
func WithSSL(ssl bool) Option {
	return func(f *FalkorDB) {
		f.ssl = ssl
	}
}

// WithHTTPClient sets a custom HTTP client (useful for testing with httprr).
func WithHTTPClient(client *http.Client) Option {
	return func(f *FalkorDB) {
		f.httpClient = client
	}
}

// WithTimeout sets the default query timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(f *FalkorDB) {
		f.timeout = timeout
	}
}

// WithReadTimeout sets the read timeout for the connection.
func WithReadTimeout(timeout time.Duration) Option {
	return func(f *FalkorDB) {
		f.readTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout for the connection.
func WithWriteTimeout(timeout time.Duration) Option {
	return func(f *FalkorDB) {
		f.writeTimeout = timeout
	}
}
