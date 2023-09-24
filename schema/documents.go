package schema

// Document is the interface for interacting with a document.
type Document struct {
	PageContent string
	Metadata    map[string]any
	Score       float32
}
