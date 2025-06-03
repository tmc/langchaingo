# Building Smart Documentation Generator

Create an AI-powered tool that automatically generates and maintains technical documentation from your codebase.

## What You'll Build

A CLI tool that:
- Analyzes Go code to understand structure and purpose
- Generates comprehensive API documentation
- Creates usage examples automatically
- Updates documentation when code changes
- Generates tutorials and guides from code patterns

## Prerequisites

- Go 1.21+
- LLM API key
- A Go project to document

## Step 1: Project Setup

```bash
mkdir smart-docs
cd smart-docs
go mod init smart-docs
go get github.com/tmc/langchaingo
go get golang.org/x/tools/go/packages
go get golang.org/x/tools/go/ast/astutil
```

## Step 2: Code Analysis Engine

Create `analyzer.go`:

```go
package main

import (
    "fmt"
    "go/ast"
    "go/doc"
    "go/parser"
    "go/token"
    "path/filepath"
    "strings"
    "unicode"
)

type CodeAnalyzer struct {
    fset *token.FileSet
}

type PackageInfo struct {
    Name        string          `json:"name"`
    Path        string          `json:"path"`
    Description string          `json:"description"`
    Functions   []FunctionInfo  `json:"functions"`
    Types       []TypeInfo      `json:"types"`
    Constants   []ConstantInfo  `json:"constants"`
    Variables   []VariableInfo  `json:"variables"`
    Examples    []ExampleInfo   `json:"examples"`
    Imports     []string        `json:"imports"`
}

type FunctionInfo struct {
    Name        string       `json:"name"`
    Signature   string       `json:"signature"`
    Description string       `json:"description"`
    Parameters  []ParamInfo  `json:"parameters"`
    Returns     []ReturnInfo `json:"returns"`
    Examples    []string     `json:"examples"`
    IsExported  bool         `json:"is_exported"`
    IsMethod    bool         `json:"is_method"`
    Receiver    string       `json:"receiver,omitempty"`
}

type TypeInfo struct {
    Name        string      `json:"name"`
    Kind        string      `json:"kind"` // struct, interface, alias, etc.
    Description string      `json:"description"`
    Fields      []FieldInfo `json:"fields,omitempty"`
    Methods     []string    `json:"methods,omitempty"`
    IsExported  bool        `json:"is_exported"`
}

type FieldInfo struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Tag         string `json:"tag,omitempty"`
    Description string `json:"description"`
}

type ParamInfo struct {
    Name string `json:"name"`
    Type string `json:"type"`
}

type ReturnInfo struct {
    Type        string `json:"type"`
    Description string `json:"description"`
}

type ConstantInfo struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Value       string `json:"value"`
    Description string `json:"description"`
    IsExported  bool   `json:"is_exported"`
}

type VariableInfo struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Description string `json:"description"`
    IsExported  bool   `json:"is_exported"`
}

type ExampleInfo struct {
    Name string `json:"name"`
    Code string `json:"code"`
    Doc  string `json:"doc"`
}

func NewCodeAnalyzer() *CodeAnalyzer {
    return &CodeAnalyzer{
        fset: token.NewFileSet(),
    }
}

func (ca *CodeAnalyzer) AnalyzePackage(dir string) (*PackageInfo, error) {
    pkgs, err := parser.ParseDir(ca.fset, dir, nil, parser.ParseComments)
    if err != nil {
        return nil, fmt.Errorf("parsing directory: %w", err)
    }

    // Find the main package (non-test)
    var pkg *ast.Package
    for name, p := range pkgs {
        if !strings.HasSuffix(name, "_test") {
            pkg = p
            break
        }
    }

    if pkg == nil {
        return nil, fmt.Errorf("no Go package found in %s", dir)
    }

    // Create documentation
    docPkg := doc.New(pkg, "./", 0)
    
    info := &PackageInfo{
        Name:        pkg.Name,
        Path:        dir,
        Description: cleanDoc(docPkg.Doc),
        Imports:     ca.extractImports(pkg),
    }

    // Analyze functions
    for _, fn := range docPkg.Funcs {
        fnInfo := ca.analyzeFunctionDecl(fn)
        info.Functions = append(info.Functions, fnInfo)
    }

    // Analyze types
    for _, typ := range docPkg.Types {
        typeInfo := ca.analyzeTypeDecl(typ)
        info.Types = append(info.Types, typeInfo)
        
        // Add methods to functions list
        for _, method := range typ.Methods {
            methodInfo := ca.analyzeFunctionDecl(method)
            methodInfo.IsMethod = true
            methodInfo.Receiver = typ.Name
            info.Functions = append(info.Functions, methodInfo)
        }
    }

    // Analyze constants and variables
    for _, c := range docPkg.Consts {
        constInfo := ca.analyzeConstantDecl(c)
        info.Constants = append(info.Constants, constInfo...)
    }

    for _, v := range docPkg.Vars {
        varInfo := ca.analyzeVariableDecl(v)
        info.Variables = append(info.Variables, varInfo...)
    }

    return info, nil
}

func (ca *CodeAnalyzer) analyzeFunctionDecl(fn *doc.Func) FunctionInfo {
    info := FunctionInfo{
        Name:        fn.Name,
        Description: cleanDoc(fn.Doc),
        IsExported:  ast.IsExported(fn.Name),
        Examples:    ca.extractExamples(fn.Doc),
    }

    if fn.Decl != nil && fn.Decl.Type != nil {
        info.Signature = ca.getFunctionSignature(fn.Decl)
        info.Parameters = ca.extractParameters(fn.Decl.Type.Params)
        info.Returns = ca.extractReturns(fn.Decl.Type.Results)
    }

    return info
}

func (ca *CodeAnalyzer) analyzeTypeDecl(typ *doc.Type) TypeInfo {
    info := TypeInfo{
        Name:        typ.Name,
        Description: cleanDoc(typ.Doc),
        IsExported:  ast.IsExported(typ.Name),
    }

    if typ.Decl != nil {
        for _, spec := range typ.Decl.Specs {
            if ts, ok := spec.(*ast.TypeSpec); ok {
                info.Kind = ca.getTypeKind(ts.Type)
                if structType, ok := ts.Type.(*ast.StructType); ok {
                    info.Fields = ca.extractFields(structType)
                }
            }
        }
    }

    // Extract method names
    for _, method := range typ.Methods {
        info.Methods = append(info.Methods, method.Name)
    }

    return info
}

func (ca *CodeAnalyzer) analyzeConstantDecl(c *doc.Value) []ConstantInfo {
    var constants []ConstantInfo
    
    for _, spec := range c.Decl.Specs {
        if vs, ok := spec.(*ast.ValueSpec); ok {
            for i, name := range vs.Names {
                constInfo := ConstantInfo{
                    Name:        name.Name,
                    Description: cleanDoc(c.Doc),
                    IsExported:  ast.IsExported(name.Name),
                }
                
                if vs.Type != nil {
                    constInfo.Type = ca.typeToString(vs.Type)
                }
                
                if i < len(vs.Values) && vs.Values[i] != nil {
                    constInfo.Value = ca.exprToString(vs.Values[i])
                }
                
                constants = append(constants, constInfo)
            }
        }
    }
    
    return constants
}

func (ca *CodeAnalyzer) analyzeVariableDecl(v *doc.Value) []VariableInfo {
    var variables []VariableInfo
    
    for _, spec := range v.Decl.Specs {
        if vs, ok := spec.(*ast.ValueSpec); ok {
            for _, name := range vs.Names {
                varInfo := VariableInfo{
                    Name:        name.Name,
                    Description: cleanDoc(v.Doc),
                    IsExported:  ast.IsExported(name.Name),
                }
                
                if vs.Type != nil {
                    varInfo.Type = ca.typeToString(vs.Type)
                }
                
                variables = append(variables, varInfo)
            }
        }
    }
    
    return variables
}

func (ca *CodeAnalyzer) extractImports(pkg *ast.Package) []string {
    importSet := make(map[string]bool)
    
    for _, file := range pkg.Files {
        for _, imp := range file.Imports {
            path := strings.Trim(imp.Path.Value, `"`)
            importSet[path] = true
        }
    }
    
    var imports []string
    for imp := range importSet {
        imports = append(imports, imp)
    }
    
    return imports
}

func (ca *CodeAnalyzer) extractParameters(fields *ast.FieldList) []ParamInfo {
    if fields == nil {
        return nil
    }
    
    var params []ParamInfo
    for _, field := range fields.List {
        paramType := ca.typeToString(field.Type)
        
        if len(field.Names) == 0 {
            // Anonymous parameter
            params = append(params, ParamInfo{
                Name: "",
                Type: paramType,
            })
        } else {
            for _, name := range field.Names {
                params = append(params, ParamInfo{
                    Name: name.Name,
                    Type: paramType,
                })
            }
        }
    }
    
    return params
}

func (ca *CodeAnalyzer) extractReturns(fields *ast.FieldList) []ReturnInfo {
    if fields == nil {
        return nil
    }
    
    var returns []ReturnInfo
    for _, field := range fields.List {
        returns = append(returns, ReturnInfo{
            Type: ca.typeToString(field.Type),
        })
    }
    
    return returns
}

func (ca *CodeAnalyzer) extractFields(structType *ast.StructType) []FieldInfo {
    var fields []FieldInfo
    
    for _, field := range structType.Fields.List {
        fieldType := ca.typeToString(field.Type)
        var tag string
        if field.Tag != nil {
            tag = field.Tag.Value
        }
        
        if len(field.Names) == 0 {
            // Embedded field
            fields = append(fields, FieldInfo{
                Name: "",
                Type: fieldType,
                Tag:  tag,
            })
        } else {
            for _, name := range field.Names {
                fields = append(fields, FieldInfo{
                    Name: name.Name,
                    Type: fieldType,
                    Tag:  tag,
                })
            }
        }
    }
    
    return fields
}

func (ca *CodeAnalyzer) getFunctionSignature(decl *ast.FuncDecl) string {
    // This is a simplified version - you'd want more sophisticated formatting
    var parts []string
    
    parts = append(parts, "func")
    
    if decl.Recv != nil {
        recv := ca.fieldListToString(decl.Recv)
        parts = append(parts, fmt.Sprintf("(%s)", recv))
    }
    
    parts = append(parts, decl.Name.Name)
    
    if decl.Type.Params != nil {
        params := ca.fieldListToString(decl.Type.Params)
        parts = append(parts, fmt.Sprintf("(%s)", params))
    } else {
        parts = append(parts, "()")
    }
    
    if decl.Type.Results != nil {
        results := ca.fieldListToString(decl.Type.Results)
        if len(decl.Type.Results.List) == 1 && len(decl.Type.Results.List[0].Names) == 0 {
            parts = append(parts, results)
        } else {
            parts = append(parts, fmt.Sprintf("(%s)", results))
        }
    }
    
    return strings.Join(parts, " ")
}

func (ca *CodeAnalyzer) fieldListToString(fields *ast.FieldList) string {
    if fields == nil {
        return ""
    }
    
    var parts []string
    for _, field := range fields.List {
        fieldType := ca.typeToString(field.Type)
        if len(field.Names) == 0 {
            parts = append(parts, fieldType)
        } else {
            for _, name := range field.Names {
                parts = append(parts, fmt.Sprintf("%s %s", name.Name, fieldType))
            }
        }
    }
    
    return strings.Join(parts, ", ")
}

func (ca *CodeAnalyzer) typeToString(expr ast.Expr) string {
    // Simplified type-to-string conversion
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        return "*" + ca.typeToString(t.X)
    case *ast.ArrayType:
        return "[]" + ca.typeToString(t.Elt)
    case *ast.MapType:
        return fmt.Sprintf("map[%s]%s", ca.typeToString(t.Key), ca.typeToString(t.Value))
    case *ast.SelectorExpr:
        return fmt.Sprintf("%s.%s", ca.typeToString(t.X), t.Sel.Name)
    case *ast.InterfaceType:
        return "interface{}"
    default:
        return "unknown"
    }
}

func (ca *CodeAnalyzer) exprToString(expr ast.Expr) string {
    switch e := expr.(type) {
    case *ast.BasicLit:
        return e.Value
    case *ast.Ident:
        return e.Name
    default:
        return "..."
    }
}

func (ca *CodeAnalyzer) getTypeKind(expr ast.Expr) string {
    switch expr.(type) {
    case *ast.StructType:
        return "struct"
    case *ast.InterfaceType:
        return "interface"
    case *ast.ArrayType:
        return "array"
    case *ast.MapType:
        return "map"
    case *ast.ChanType:
        return "channel"
    case *ast.FuncType:
        return "function"
    default:
        return "alias"
    }
}

func (ca *CodeAnalyzer) extractExamples(doc string) []string {
    // Extract code examples from documentation
    var examples []string
    lines := strings.Split(doc, "\n")
    
    var inExample bool
    var currentExample strings.Builder
    
    for _, line := range lines {
        trimmed := strings.TrimSpace(line)
        
        if strings.HasPrefix(trimmed, "Example:") || 
           strings.HasPrefix(trimmed, "Usage:") ||
           strings.Contains(trimmed, "```go") {
            inExample = true
            currentExample.Reset()
            continue
        }
        
        if inExample {
            if strings.Contains(trimmed, "```") || 
               (trimmed == "" && currentExample.Len() > 0) {
                if currentExample.Len() > 0 {
                    examples = append(examples, currentExample.String())
                    currentExample.Reset()
                }
                inExample = false
                continue
            }
            
            if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
                currentExample.WriteString(strings.TrimPrefix(strings.TrimPrefix(line, "    "), "\t"))
                currentExample.WriteString("\n")
            }
        }
    }
    
    return examples
}

func cleanDoc(doc string) string {
    if doc == "" {
        return ""
    }
    
    // Remove leading/trailing whitespace and normalize line endings
    doc = strings.TrimSpace(doc)
    doc = strings.ReplaceAll(doc, "\r\n", "\n")
    
    // Remove common documentation artifacts
    lines := strings.Split(doc, "\n")
    var cleaned []string
    
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" {
            cleaned = append(cleaned, line)
        }
    }
    
    return strings.Join(cleaned, "\n")
}
```

## Step 3: Documentation Generator

Create `generator.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "path/filepath"
    "strings"
    "text/template"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/prompts"
)

type DocGenerator struct {
    llm llms.Model
    templates map[string]*template.Template
}

type DocConfig struct {
    ProjectName    string `json:"project_name"`
    ProjectDesc    string `json:"project_description"`
    OutputDir      string `json:"output_dir"`
    IncludePrivate bool   `json:"include_private"`
    GenerateExamples bool `json:"generate_examples"`
    Style          string `json:"style"` // "godoc", "markdown", "html"
}

func NewDocGenerator() (*DocGenerator, error) {
    llm, err := openai.New()
    if err != nil {
        return nil, fmt.Errorf("creating LLM: %w", err)
    }

    dg := &DocGenerator{
        llm: llm,
        templates: make(map[string]*template.Template),
    }

    if err := dg.loadTemplates(); err != nil {
        return nil, fmt.Errorf("loading templates: %w", err)
    }

    return dg, nil
}

func (dg *DocGenerator) loadTemplates() error {
    // Package documentation template
    packageTmpl := `# {{.Name}}

{{.Description}}

## Installation

'''bash
go get {{.Path}}
'''

## Usage

{{if .Examples}}
{{range .Examples}}
'''go
{{.Code}}
'''
{{end}}
{{end}}

## API Reference

{{if .Functions}}
### Functions

{{range .Functions}}
{{if .IsExported}}
#### {{.Name}}

'''go
{{.Signature}}
'''

{{.Description}}

{{if .Parameters}}
**Parameters:**
{{range .Parameters}}
- '{{.Name}}' ({{.Type}})
{{end}}
{{end}}

{{if .Returns}}
**Returns:**
{{range .Returns}}
- {{.Type}}{{if .Description}} - {{.Description}}{{end}}
{{end}}
{{end}}

{{if .Examples}}
**Example:**
{{range .Examples}}
'''go
{{.}}
'''
{{end}}
{{end}}

{{end}}
{{end}}
{{end}}

{{if .Types}}
### Types

{{range .Types}}
{{if .IsExported}}
#### {{.Name}}

'''go
type {{.Name}} {{.Kind}}
'''

{{.Description}}

{{if .Fields}}
**Fields:**
{{range .Fields}}
- '{{.Name}}' {{.Type}}{{if .Description}} - {{.Description}}{{end}}
{{end}}
{{end}}

{{if .Methods}}
**Methods:**
{{range .Methods}}
- [{{.}}](#{{.}})
{{end}}
{{end}}

{{end}}
{{end}}
{{end}}
`

    tmpl, err := template.New("package").Parse(packageTmpl)
    if err != nil {
        return fmt.Errorf("parsing package template: %w", err)
    }
    dg.templates["package"] = tmpl

    return nil
}

func (dg *DocGenerator) GeneratePackageDoc(pkg *PackageInfo, config DocConfig) (string, error) {
    // Enhance descriptions with AI
    if err := dg.enhanceDescriptions(pkg); err != nil {
        return "", fmt.Errorf("enhancing descriptions: %w", err)
    }

    // Generate usage examples
    if config.GenerateExamples {
        if err := dg.generateExamples(pkg); err != nil {
            return "", fmt.Errorf("generating examples: %w", err)
        }
    }

    // Apply template
    var result strings.Builder
    if err := dg.templates["package"].Execute(&result, pkg); err != nil {
        return "", fmt.Errorf("executing template: %w", err)
    }

    return result.String(), nil
}

func (dg *DocGenerator) enhanceDescriptions(pkg *PackageInfo) error {
    ctx := context.Background()

    // Enhance package description if empty or too brief
    if len(pkg.Description) < 50 {
        enhanced, err := dg.enhancePackageDescription(ctx, pkg)
        if err == nil && enhanced != "" {
            pkg.Description = enhanced
        }
    }

    // Enhance function descriptions
    for i := range pkg.Functions {
        if len(pkg.Functions[i].Description) < 20 {
            enhanced, err := dg.enhanceFunctionDescription(ctx, &pkg.Functions[i])
            if err == nil && enhanced != "" {
                pkg.Functions[i].Description = enhanced
            }
        }
    }

    // Enhance type descriptions
    for i := range pkg.Types {
        if len(pkg.Types[i].Description) < 20 {
            enhanced, err := dg.enhanceTypeDescription(ctx, &pkg.Types[i])
            if err == nil && enhanced != "" {
                pkg.Types[i].Description = enhanced
            }
        }
    }

    return nil
}

func (dg *DocGenerator) enhancePackageDescription(ctx context.Context, pkg *PackageInfo) (string, error) {
    template := prompts.NewPromptTemplate(`
Analyze this Go package and write a clear, concise description (2-3 sentences):

Package: {{.name}}
Path: {{.path}}

Functions: {{range .functions}}{{.name}}, {{end}}
Types: {{range .types}}{{.name}}, {{end}}

Write a professional description that explains:
1. What this package does
2. Who would use it
3. Key capabilities

Keep it under 200 words and avoid marketing language.`, 
        []string{"name", "path", "functions", "types"})

    prompt, err := template.Format(map[string]any{
        "name":      pkg.Name,
        "path":      pkg.Path,
        "functions": pkg.Functions,
        "types":     pkg.Types,
    })
    if err != nil {
        return "", err
    }

    response, err := dg.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(response.Choices[0].Content), nil
}

func (dg *DocGenerator) enhanceFunctionDescription(ctx context.Context, fn *FunctionInfo) (string, error) {
    template := prompts.NewPromptTemplate(`
Write a clear description for this Go function:

Function: {{.name}}
Signature: {{.signature}}
{{if .parameters}}Parameters: {{range .parameters}}{{.name}} {{.type}}, {{end}}{{end}}
{{if .returns}}Returns: {{range .returns}}{{.type}}, {{end}}{{end}}

Describe what it does, when to use it, and any important behavior.
Keep it concise (1-2 sentences).`, 
        []string{"name", "signature", "parameters", "returns"})

    prompt, err := template.Format(map[string]any{
        "name":       fn.Name,
        "signature":  fn.Signature,
        "parameters": fn.Parameters,
        "returns":    fn.Returns,
    })
    if err != nil {
        return "", err
    }

    response, err := dg.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(response.Choices[0].Content), nil
}

func (dg *DocGenerator) enhanceTypeDescription(ctx context.Context, typ *TypeInfo) (string, error) {
    template := prompts.NewPromptTemplate(`
Write a clear description for this Go type:

Type: {{.name}} ({{.kind}})
{{if .fields}}Fields: {{range .fields}}{{.name}} {{.type}}, {{end}}{{end}}
{{if .methods}}Methods: {{range .methods}}{{.}}, {{end}}{{end}}

Describe what it represents and how it's used.
Keep it concise (1-2 sentences).`, 
        []string{"name", "kind", "fields", "methods"})

    prompt, err := template.Format(map[string]any{
        "name":    typ.Name,
        "kind":    typ.Kind,
        "fields":  typ.Fields,
        "methods": typ.Methods,
    })
    if err != nil {
        return "", err
    }

    response, err := dg.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(response.Choices[0].Content), nil
}

func (dg *DocGenerator) generateExamples(pkg *PackageInfo) error {
    ctx := context.Background()

    // Generate package-level usage example
    if len(pkg.Examples) == 0 {
        example, err := dg.generatePackageExample(ctx, pkg)
        if err == nil && example != "" {
            pkg.Examples = append(pkg.Examples, ExampleInfo{
                Name: "Basic Usage",
                Code: example,
                Doc:  "Basic usage example",
            })
        }
    }

    // Generate function examples
    for i := range pkg.Functions {
        if len(pkg.Functions[i].Examples) == 0 && pkg.Functions[i].IsExported {
            example, err := dg.generateFunctionExample(ctx, &pkg.Functions[i], pkg)
            if err == nil && example != "" {
                pkg.Functions[i].Examples = append(pkg.Functions[i].Examples, example)
            }
        }
    }

    return nil
}

func (dg *DocGenerator) generatePackageExample(ctx context.Context, pkg *PackageInfo) (string, error) {
    template := prompts.NewPromptTemplate(`
Create a realistic Go code example showing how to use this package:

Package: {{.name}}
Description: {{.description}}
Key Functions: {{range .functions}}{{if .is_exported}}{{.name}}, {{end}}{{end}}
Key Types: {{range .types}}{{if .is_exported}}{{.name}}, {{end}}{{end}}

Write a complete, runnable example that shows:
1. Import statement
2. Basic usage
3. Error handling
4. Realistic use case

Return only the Go code, no explanations.`, 
        []string{"name", "description", "functions", "types"})

    prompt, err := template.Format(map[string]any{
        "name":        pkg.Name,
        "description": pkg.Description,
        "functions":   pkg.Functions,
        "types":       pkg.Types,
    })
    if err != nil {
        return "", err
    }

    response, err := dg.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(response.Choices[0].Content), nil
}

func (dg *DocGenerator) generateFunctionExample(ctx context.Context, fn *FunctionInfo, pkg *PackageInfo) (string, error) {
    template := prompts.NewPromptTemplate(`
Create a Go code example for this function:

Function: {{.name}}
Signature: {{.signature}}
Package: {{.package}}
{{if .parameters}}Parameters: {{range .parameters}}{{.name}} {{.type}}, {{end}}{{end}}

Write a realistic example showing how to call this function.
Include proper error handling if needed.
Return only the Go code snippet.`, 
        []string{"name", "signature", "package", "parameters"})

    prompt, err := template.Format(map[string]any{
        "name":       fn.Name,
        "signature":  fn.Signature,
        "package":    pkg.Name,
        "parameters": fn.Parameters,
    })
    if err != nil {
        return "", err
    }

    response, err := dg.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return "", err
    }

    return strings.TrimSpace(response.Choices[0].Content), nil
}
```

## Step 4: Main Application

Create `main.go`:

```go
package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
)

func main() {
    var (
        projectDir  = flag.String("dir", ".", "Project directory to analyze")
        outputDir   = flag.String("output", "./docs", "Output directory for documentation")
        configFile  = flag.String("config", "", "Configuration file")
        watch       = flag.Bool("watch", false, "Watch for changes and regenerate")
        packageName = flag.String("package", "", "Specific package to document")
    )
    flag.Parse()

    // Load configuration
    config := DocConfig{
        OutputDir:        *outputDir,
        IncludePrivate:   false,
        GenerateExamples: true,
        Style:           "markdown",
    }

    if *configFile != "" {
        if err := loadConfig(*configFile, &config); err != nil {
            log.Printf("Warning: Could not load config file: %v", err)
        }
    }

    // Initialize components
    analyzer := NewCodeAnalyzer()
    generator, err := NewDocGenerator()
    if err != nil {
        log.Fatal(err)
    }

    if *watch {
        if err := watchAndGenerate(analyzer, generator, *projectDir, config); err != nil {
            log.Fatal(err)
        }
    } else {
        if err := generateDocs(analyzer, generator, *projectDir, config, *packageName); err != nil {
            log.Fatal(err)
        }
    }
}

func generateDocs(analyzer *CodeAnalyzer, generator *DocGenerator, projectDir string, config DocConfig, packageName string) error {
    if packageName != "" {
        // Document specific package
        return generatePackageDocs(analyzer, generator, filepath.Join(projectDir, packageName), config)
    }

    // Document all packages
    return filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if !info.IsDir() {
            return nil
        }

        // Skip vendor, .git, and test directories
        if shouldSkipDir(path) {
            return filepath.SkipDir
        }

        // Check if directory contains Go files
        hasGoFiles, err := hasGoSourceFiles(path)
        if err != nil {
            return err
        }

        if hasGoFiles {
            if err := generatePackageDocs(analyzer, generator, path, config); err != nil {
                log.Printf("Error documenting package %s: %v", path, err)
            }
        }

        return nil
    })
}

func generatePackageDocs(analyzer *CodeAnalyzer, generator *DocGenerator, packageDir string, config DocConfig) error {
    fmt.Printf("Analyzing package: %s\n", packageDir)

    // Analyze package
    pkg, err := analyzer.AnalyzePackage(packageDir)
    if err != nil {
        return fmt.Errorf("analyzing package: %w", err)
    }

    // Generate documentation
    doc, err := generator.GeneratePackageDoc(pkg, config)
    if err != nil {
        return fmt.Errorf("generating documentation: %w", err)
    }

    // Write to file
    outputPath := filepath.Join(config.OutputDir, pkg.Name+".md")
    if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
        return fmt.Errorf("creating output directory: %w", err)
    }

    if err := os.WriteFile(outputPath, []byte(doc), 0644); err != nil {
        return fmt.Errorf("writing documentation: %w", err)
    }

    fmt.Printf("Generated documentation: %s\n", outputPath)
    return nil
}

func watchAndGenerate(analyzer *CodeAnalyzer, generator *DocGenerator, projectDir string, config DocConfig) error {
    // Simplified file watching - you'd want to use fsnotify for production
    fmt.Printf("Watching %s for changes...\n", projectDir)
    
    for {
        if err := generateDocs(analyzer, generator, projectDir, config, ""); err != nil {
            log.Printf("Error generating docs: %v", err)
        }
        time.Sleep(30 * time.Second)
    }
}

func shouldSkipDir(path string) bool {
    base := filepath.Base(path)
    return base == "vendor" || 
           base == ".git" || 
           base == "testdata" ||
           strings.HasSuffix(base, "_test")
}

func hasGoSourceFiles(dir string) (bool, error) {
    files, err := os.ReadDir(dir)
    if err != nil {
        return false, err
    }

    for _, file := range files {
        if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") && 
           !strings.HasSuffix(file.Name(), "_test.go") {
            return true, nil
        }
    }

    return false, nil
}

func loadConfig(filename string, config *DocConfig) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return err
    }

    return json.Unmarshal(data, config)
}
```

## Step 5: Run the Documentation Generator

Create a sample package to document:

```bash
mkdir example-package
cd example-package
cat > user.go << 'EOF'
// Package user provides user management functionality.
package user

import (
    "errors"
    "time"
)

// User represents a system user.
type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Created  time.Time `json:"created"`
}

// UserService handles user operations.
type UserService struct {
    users map[int]*User
}

// NewUserService creates a new user service.
func NewUserService() *UserService {
    return &UserService{
        users: make(map[int]*User),
    }
}

// CreateUser creates a new user.
func (s *UserService) CreateUser(name, email string) (*User, error) {
    if name == "" {
        return nil, errors.New("name cannot be empty")
    }
    
    user := &User{
        ID:      len(s.users) + 1,
        Name:    name,
        Email:   email,
        Created: time.Now(),
    }
    
    s.users[user.ID] = user
    return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(id int) (*User, error) {
    user, exists := s.users[id]
    if !exists {
        return nil, errors.New("user not found")
    }
    return user, nil
}
EOF
```

Run the documentation generator:

```bash
cd ..
export OPENAI_API_KEY="your-openai-api-key-here"
go run *.go -dir=example-package -output=./generated-docs
```

## Step 6: Configuration File

Create `docgen.json`:

```json
{
    "project_name": "My Go Project",
    "project_description": "A sample Go project for documentation",
    "output_dir": "./docs",
    "include_private": false,
    "generate_examples": true,
    "style": "markdown"
}
```

## Advanced Features

### Integration with CI/CD

Create `.github/workflows/docs.yml`:

```yaml
name: Generate Documentation
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Generate Documentation
      env:
        OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
      run: |
        go run cmd/smart-docs/*.go -dir=. -output=./docs
    
    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: ./docs
```

### Multiple Output Formats

Add support for different output formats:

```go
func (dg *DocGenerator) GenerateHTML(pkg *PackageInfo) (string, error) {
    // Generate HTML documentation
}

func (dg *DocGenerator) GenerateOpenAPI(pkg *PackageInfo) (string, error) {
    // Generate OpenAPI spec from HTTP handlers
}
```

## Use Cases

This smart documentation generator can:

1. **Automate API documentation** for web services
2. **Generate SDK documentation** for client libraries  
3. **Create internal documentation** for large codebases
4. **Maintain up-to-date docs** in CI/CD pipelines
5. **Generate examples** for complex APIs

This tutorial shows how LangChainGo can create sophisticated development tools that save significant time and improve code quality!