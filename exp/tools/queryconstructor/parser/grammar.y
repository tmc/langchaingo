
%{
package queryconstructor_parser

%}

// the union will take those values from grammar file (.y) and put them in yySymType struct
%union {
    function StructuredFilter
    expr interface{}
    argString string
    argBoolean bool
    argFloat float64
    argInt int
    args []interface{}
    functionName string
}

// create yacc types and map them to yacc "union" types
%type <functionName>  FunctionName
%type <args> Args
%type <expr> Expr
%type <function> FunctionCall

// define what are the symbol used for syntax

%token LPAREN RPAREN COMMA
%token <argString>  ArgString
%token <argBoolean>  ArgBoolean
%token <argFloat>  ArgFloat
%token <argInt>  ArgInt
%token <functionName>  TokenFunctionName

%left '(' LPAREN
%left ',' COMMA
%left ')' RPAREN

// the request returned by LLM start with a function call
%start query
 
%%

query:
    FunctionCall {
        yylex.(*Lexer).function = $1
    }


FunctionCall
: FunctionName LPAREN Args RPAREN {
    $$ = setFunction($1, $3)
}
;

Args : Expr { $$ = []interface{}{$1} }
| Args COMMA Expr { $$ = append($1, $3) }
;

Expr : 
    ArgString  { $$ = $1 }
    | ArgBoolean { $$ = $1 }
    | ArgFloat { $$ = $1 }
    | ArgInt { $$ = $1 }
    | FunctionCall { $$ = $1 }
;

FunctionName: TokenFunctionName
;

%%