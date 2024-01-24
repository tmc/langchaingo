package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", os.Stdin, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	//file.Imports = append(file.Imports, &ast.ImportSpec{
	//Path: &ast.BasicLit{
	//Kind: token.STRING,
	////Value: `"github.com/tmc/langchaingo/llms/googleai"`,
	//Value: "sdfsdfsdf",
	//},
	//})

	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.ImportSpec:
			if strings.Index(x.Path.Value, "generative-ai-go/genai") > 0 {
				x.Path.Value = `"cloud.google.com/go/vertexai/genai"`
			}
			//id, ok := x.Fun.(*ast.Ident)
			//if ok {
			//if id.Name == "pred" {
			//c.Replace(&ast.UnaryExpr{
			//Op: token.NOT,
			//X:  x,
			//})
			//}
			//}

		case *ast.FuncDecl:
			if x.Recv != nil && len(x.Recv.List) == 1 {
				recv := x.Recv.List[0]
				ty := recv.Type.(*ast.StarExpr)
				tyName := ty.X.(*ast.Ident)
				tyName.Name = "Vertex"
			}
		}

		return true
	})

	fmt.Println("Modified AST:")
	//printer.Fprint(os.Stdout, fset, file)

	format.Node(os.Stdout, fset, file)
}
