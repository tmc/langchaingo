package llms

import "errors"

var (
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
)
