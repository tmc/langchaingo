package mongodb

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvalidOptions = errors.New("invalid options")

var serverAPI = options.ServerAPI(options.ServerAPIVersion1)

type Option func(p *Store)

func WithConnectionUri(connectionUri string) Option {
	return func(p *Store) {
		p.ConnectionUri = connectionUri
	}
}

func applyClientOptions(opts ...Option) (Store, error) {
	o := &Store{} //todo add default values
	for _, opt := range opts {
		opt(o)
	}
	if o.ConnectionUri == "" {
		return Store{}, fmt.Errorf("%w: missing mongodb connection string", ErrInvalidOptions)
	}
	o.ClientOptions = options.Client().ApplyURI(o.ConnectionUri).SetServerAPIOptions(serverAPI)
	return *o, nil
}


