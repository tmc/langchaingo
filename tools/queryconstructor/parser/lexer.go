package queryconstructorparser

import (
	"errors"
	"go/scanner"
	"go/token"
	"strconv"
	"strings"
)

type StructuredFilter struct {
	Args         []interface{}
	FunctionName string
}

func setFunction(functionName string, args []interface{}) StructuredFilter {
	return StructuredFilter{
		FunctionName: functionName,
		Args:         args,
	}
}

// lexer is used to define what kind of token are passed.
type Lexer struct {
	scan     *scanner.Scanner
	function StructuredFilter
	err      error
}

// pseudo constructor for lexer.
func NewLexer(query string) Lexer {
	scan := scanner.Scanner{}

	fset := token.NewFileSet()

	scan.Init(fset.AddFile("query", -1, len(query)), []byte(query), func(pos token.Position, msg string) {}, scanner.ScanComments)

	return Lexer{
		scan: &scan,
	}
}

// Lex is used to understand what kind of token is parsed

//nolint:cyclop
func (l *Lexer) Lex(lval *yySymType) int {
	_, tok, lit := l.scan.Scan()

	switch {
	case tok == token.LPAREN:
		return LPAREN
	case tok == token.RPAREN:
		return RPAREN
	case tok == token.COMMA:
		return COMMA
	case token.IDENT == tok && strings.ToLower(lit) == "true" || strings.ToLower(lit) == "false":
		boolLit, _ := strconv.ParseBool(lit)
		lval.argBoolean = boolLit
		return ArgBoolean
	case token.IDENT == tok:
		lval.functionName = lit
		return TokenFunctionName
	case token.STRING == tok:
		trimmedArgString := strings.Trim(lit, `"`)
		lval.argString = trimmedArgString
		return ArgString
	case token.FLOAT == tok:
		float64Lit, _ := strconv.ParseFloat(lit, 64)
		lval.argFloat = float64Lit
		return ArgFloat
	case token.INT == tok:
		intLit, _ := strconv.Atoi(lit)
		lval.argInt = intLit
		return ArgInt
	case token.SEMICOLON == tok || token.EOF == tok:
		fallthrough
	default:
		return 0
	}
}

// handle error from goyacc parser.
func (l *Lexer) Error(e string) {
	l.err = errors.New(e)
}
