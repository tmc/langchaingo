package textsplitter

const (
	_defaultChunkSize    = 4000
	_defaultChunkOverlap = 200
)

// TextSplitter is the standard interface for splitting texts.
type TextSplitter interface {
	SplitText(string) ([]string, error)
}
