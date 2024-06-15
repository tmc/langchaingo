package mongodb

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvalidOptions = errors.New("invalid options")

var serverAPI = options.ServerAPI(options.ServerAPIVersion1)

type Option func(p *Store)

func WithConnectionString(connectionUri string) Option {
	return func(p *Store) {
		p.connectionUri = connectionUri
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{} //todo add default values
	for _, opt := range opts {
		opt(o)
	}
	if o.connectionUri == "" {
		return Store{}, fmt.Errorf("%w: missing mongodb connection string", ErrInvalidOptions)
	}
	o.clientOptions = options.Client().ApplyURI(o.connectionUri).SetServerAPIOptions(serverAPI)
	return *o, nil
}


