package textsplitter

// TextSplitter is the standard interface for splitting texts.
type TextSplitter interface {
	SplitText(string) ([]string, error)
}
