package fstring

import "errors"

var (
	ErrEmptyExpression       = errors.New("empty expression not allowed")
	ErrArgsNotDefined        = errors.New("args not defined")
	ErrLeftBracketNotClosed  = errors.New("single '{' is not allowed")
	ErrRightBracketNotClosed = errors.New("single '}' is not allowed")
)

// Format interpolates the given template with the given values by using
// f-string.
func Format(template string, values map[string]any) (string, error) {
	p := newParser(template, values)
	if err := p.parse(); err != nil {
		return "", err
	}
	return string(p.result), nil
}
