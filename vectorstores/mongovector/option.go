package mongovector

// Option sets mongovector-specific options when constructing a Store.
type Option func(p *Store)

// WithIndex will set the default index to use when adding or searching
// documents with the store.
//
// Atlas Vector Search doesn't return results if you misspell the index name or
// if the specified index doesn't already exist on the cluster.
//
// The index can be update at the operation level with the NameSpace
// vectorstores option.
func WithIndex(index string) Option {
	return func(p *Store) {
		p.index = index
	}
}

// WithPath will set the path parameter used by the Atlas Search operators to
// specify the field or fields to be searched.
func WithPath(path string) Option {
	return func(p *Store) {
		p.path = path
	}
}

// WithNumCandidates sets the number of nearest neighbors to use during a
// similarity search. By default this value is 10 times the number of documents
// (or limit) passed as an argument to SimilaritySearch.
func WithNumCandidates(numCandidates int) Option {
	return func(p *Store) {
		p.numCandidates = numCandidates
	}
}
