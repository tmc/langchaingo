// Package main provides a tool to analyze value vs pointer receivers in the codebase
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ReceiverInfo struct {
	Type       string
	Method     string
	IsPointer  bool
	File       string
	Line       int
	StructSize string // estimated
}

type StructInfo struct {
	Name   string
	File   string
	Fields []string
	Size   int // estimated number of fields
}

func main() {
	receivers := []ReceiverInfo{}
	structs := map[string]StructInfo{}

	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, .git, and test files for now
		if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
			return nil
		}

		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			analyzeFile(path, &receivers, structs)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	generateReport(receivers, structs)
}

func analyzeFile(filename string, receivers *[]ReceiverInfo, structs map[string]StructInfo) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Recv != nil && len(x.Recv.List) > 0 {
				recv := x.Recv.List[0]
				var typeName string
				var isPointer bool

				switch t := recv.Type.(type) {
				case *ast.Ident:
					typeName = t.Name
					isPointer = false
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						typeName = ident.Name
						isPointer = true
					}
				}

				if typeName != "" {
					pos := fset.Position(x.Pos())
					*receivers = append(*receivers, ReceiverInfo{
						Type:      typeName,
						Method:    x.Name.Name,
						IsPointer: isPointer,
						File:      filename,
						Line:      pos.Line,
					})
				}
			}
		case *ast.TypeSpec:
			if structType, ok := x.Type.(*ast.StructType); ok {
				fields := []string{}
				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						fields = append(fields, name.Name)
					}
				}
				structs[x.Name.Name] = StructInfo{
					Name:   x.Name.Name,
					File:   filename,
					Fields: fields,
					Size:   len(fields),
				}
			}
		}
		return true
	})
}

func generateReport(receivers []ReceiverInfo, structs map[string]StructInfo) {
	fmt.Println("# Receiver Analysis Report")
	fmt.Println()

	// Group by type
	typeGroups := make(map[string][]ReceiverInfo)
	for _, r := range receivers {
		typeGroups[r.Type] = append(typeGroups[r.Type], r)
	}

	// Analyze patterns
	fmt.Println("## Summary")
	fmt.Printf("- Total types analyzed: %d\n", len(typeGroups))
	fmt.Printf("- Total methods analyzed: %d\n", len(receivers))
	fmt.Println()

	valueReceivers := 0
	pointerReceivers := 0
	mixedTypes := 0

	for _, methods := range typeGroups {
		hasValue := false
		hasPointer := false
		
		for _, method := range methods {
			if method.IsPointer {
				hasPointer = true
			} else {
				hasValue = true
			}
		}
		
		if hasValue && !hasPointer {
			valueReceivers++
		} else if hasPointer && !hasValue {
			pointerReceivers++
		} else {
			mixedTypes++
		}
	}

	fmt.Printf("- Types using only value receivers: %d\n", valueReceivers)
	fmt.Printf("- Types using only pointer receivers: %d\n", pointerReceivers)
	fmt.Printf("- Types with mixed receivers: %d\n", mixedTypes)
	fmt.Println()

	// Detailed analysis
	fmt.Println("## Detailed Analysis")
	fmt.Println()

	// Sort types by name for consistent output
	var sortedTypes []string
	for typeName := range typeGroups {
		sortedTypes = append(sortedTypes, typeName)
	}
	sort.Strings(sortedTypes)

	problemTypes := []string{}

	for _, typeName := range sortedTypes {
		methods := typeGroups[typeName]
		_ = typeName // Use the variable to avoid unused warning
		
		hasValue := false
		hasPointer := false
		for _, method := range methods {
			if method.IsPointer {
				hasPointer = true
			} else {
				hasValue = true
			}
		}

		structInfo := structs[typeName]
		
		fmt.Printf("### %s\n", typeName)
		if structInfo.Name != "" {
			fmt.Printf("- **Struct size**: %d fields\n", structInfo.Size)
			fmt.Printf("- **File**: %s\n", structInfo.File)
		}
		
		if hasValue && hasPointer {
			fmt.Printf("- **⚠️  MIXED RECEIVERS** - This is problematic!\n")
			problemTypes = append(problemTypes, typeName)
		} else if hasValue {
			if structInfo.Size > 3 {
				fmt.Printf("- **⚠️  Value receivers on large struct** - Consider pointer receivers\n")
				problemTypes = append(problemTypes, typeName)
			} else {
				fmt.Printf("- **✅ Value receivers** - Appropriate for small struct\n")
			}
		} else {
			fmt.Printf("- **✅ Pointer receivers** - Good for larger structs or when mutation is needed\n")
		}

		fmt.Printf("- **Methods**: %d\n", len(methods))
		for _, method := range methods {
			receiverType := "value"
			if method.IsPointer {
				receiverType = "pointer"
			}
			fmt.Printf("  - `%s()` - %s receiver\n", method.Method, receiverType)
		}
		fmt.Println()
	}

	// Recommendations
	fmt.Println("## Recommendations")
	fmt.Println()
	
	if len(problemTypes) > 0 {
		fmt.Println("### 🚨 Issues Found")
		fmt.Println()
		fmt.Println("The following types should be reviewed:")
		for _, typeName := range problemTypes {
			methods := typeGroups[typeName]
			structInfo := structs[typeName]
			
			fmt.Printf("#### %s\n", typeName)
			
			hasValue := false
			hasPointer := false
			for _, method := range methods {
				if method.IsPointer {
					hasPointer = true
				} else {
					hasValue = true
				}
			}
			
			if hasValue && hasPointer {
				fmt.Println("**Issue**: Mixed receiver types")
				fmt.Println("**Recommendation**: Choose either all value or all pointer receivers")
				fmt.Println("**Guideline**: Use pointer receivers if:")
				fmt.Println("- The struct is large (>3-4 fields)")
				fmt.Println("- Methods need to modify the receiver")
				fmt.Println("- The struct contains sync.Mutex or similar fields")
				fmt.Println("- Other methods already use pointer receivers")
			} else if hasValue && structInfo.Size > 3 {
				fmt.Println("**Issue**: Value receivers on potentially large struct")
				fmt.Println("**Recommendation**: Consider switching to pointer receivers to avoid copying")
				fmt.Printf("**Details**: %d fields may be expensive to copy\n", structInfo.Size)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("### ✅ No Issues Found")
		fmt.Println()
		fmt.Println("All types consistently use appropriate receiver types!")
	}

	fmt.Println("### General Guidelines")
	fmt.Println()
	fmt.Println("**Use pointer receivers when:**")
	fmt.Println("- The method needs to modify the receiver")
	fmt.Println("- The struct is large (typically >3-4 fields)")
	fmt.Println("- The struct contains sync.Mutex or similar types")
	fmt.Println("- To maintain consistency if some methods already use pointer receivers")
	fmt.Println()
	fmt.Println("**Use value receivers when:**")
	fmt.Println("- The struct is small (1-3 simple fields)")
	fmt.Println("- The method doesn't modify the receiver")
	fmt.Println("- You want to ensure the receiver is immutable")
	fmt.Println("- The type is a basic type like int, string, or small struct")
}