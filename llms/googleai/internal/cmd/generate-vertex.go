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

	// TODO: run through goimports

	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.ImportSpec:
			rewriteImport(x)

		case *ast.FuncDecl:
			if x.Recv != nil && len(x.Recv.List) == 1 {
				rewriteReceiverName(x)
			}
			addCastToTopK(x)
		}

		return true
	})

	fmt.Println("Modified AST:")
	//printer.Fprint(os.Stdout, fset, file)

	format.Node(os.Stdout, fset, file)
}

func rewriteImport(x *ast.ImportSpec) {
	if strings.Index(x.Path.Value, "generative-ai-go/genai") > 0 {
		x.Path.Value = `"cloud.google.com/go/vertexai/genai"`
	}
}

func rewriteReceiverName(fun *ast.FuncDecl) {
	recv := fun.Recv.List[0]
	ty := recv.Type.(*ast.StarExpr)
	tyName := ty.X.(*ast.Ident)
	tyName.Name = "Vertex"
}

func addCastToTopK(fun *ast.FuncDecl) {
	ast.Inspect(fun, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			if getIdentName(x.Fun) == "int32" && len(x.Args) == 1 {
				arg0 := x.Args[0]
				if sel, ok := arg0.(*ast.SelectorExpr); ok {
					if getIdentName(sel.X) == "opts" {
						if getIdentName(sel.Sel) == "TopK" {
							funcId := x.Fun.(*ast.Ident)
							funcId.Name = "float32"
						}
					}
				}
			}
		}
		return true
	})
}

// getIdentName returns the identifier name from ast.Ident expressions; for
// other expressions, returns an empty string.
func getIdentName(x ast.Expr) string {
	if id, ok := x.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
