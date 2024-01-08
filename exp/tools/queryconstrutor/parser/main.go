package queryconstructor_parser

func Parse(request string) *Function {
	lexer := NewLexer(request)
	yyParse(&lexer)

	return &lexer.function
}
