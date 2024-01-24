package queryconstructorparser

// Parse is a helper function to use lexer.
func Parse(text string) (*StructuredFilter, error) {
	lexer := NewLexer(text)
	yyParse(&lexer)

	if lexer.err != nil {
		return nil, lexer.err
	}

	return &lexer.function, nil
}
