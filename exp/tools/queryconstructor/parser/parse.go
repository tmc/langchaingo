package queryconstructor_parser

func Parse(text string) (*StructuredFilter, error) {
	lexer := NewLexer(text)
	yyParse(&lexer)

	if lexer.err != nil {
		return nil, lexer.err
	}

	return &lexer.function, nil
}
