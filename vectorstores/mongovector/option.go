package mongovector

type Option func(p *Store)

func WithIndex(index string) Option {
	return func(p *Store) {
		p.index = index
	}
}

func WithPath(path string) Option {
	return func(p *Store) {
		p.path = path
	}
}
