package vectorstores

// Option is a function that configures an Options.
type Option func(*Options)

// Options is a set of options for similarity search and add documents.
type Options struct {
	NameSpace string
}

// WithNameSpace returns an Option for setting the name space.
func WithNameSpace(nameSpace string) Option {
	return func(o *Options) {
		o.NameSpace = nameSpace
	}
}
