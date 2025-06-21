# Building an AI code reviewer

Create an intelligent code review assistant that analyzes Go code for bugs, style issues, and performance improvements.

## What you'll build

A CLI tool that:
- Analyzes Go source files for potential issues
- Suggests improvements and best practices
- Integrates with git to review changed files
- Provides explanations for its recommendations

## Prerequisites

- Go 1.21+
- OpenAI or Anthropic API key
- Git installed

## Step 1: Project setup

```bash
mkdir ai-code-reviewer
cd ai-code-reviewer
go mod init code-reviewer
go get github.com/tmc/langchaingo
```

## Step 2: Core reviewer structure

Create `main.go`:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/prompts"
)

type CodeReviewer struct {
    llm llms.Model
    template *prompts.PromptTemplate
}

func NewCodeReviewer() (*CodeReviewer, error) {
    llm, err := openai.New()
    if err != nil {
        return nil, err
    }

    template := prompts.NewPromptTemplate(`
You are an expert Go code reviewer. Analyze this Go code for:

1. **Bugs and Logic Issues**: Potential runtime errors, nil pointer dereferences, race conditions
2. **Performance**: Inefficient algorithms, unnecessary allocations, string concatenation issues
3. **Style**: Go idioms, naming conventions, error handling patterns
4. **Security**: Input validation, sensitive data handling

Code to review:
'''go
{{.code}}
'''

File: {{.filename}}

Provide specific, actionable feedback. For each issue:
- Explain WHY it's a problem
- Show HOW to fix it with code examples
- Rate severity: Critical, Warning, Suggestion

Focus on the most important issues first.`, 
        []string{"code", "filename"})

    return &CodeReviewer{
        llm: llm,
        template: &template,
    }, nil
}

func (cr *CodeReviewer) ReviewFile(filename string) error {
    content, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("reading file: %w", err)
    }

    // Parse Go code to ensure it's valid
    fset := token.NewFileSet()
    _, err = parser.ParseFile(fset, filename, content, parser.ParseComments)
    if err != nil {
        return fmt.Errorf("parsing Go file: %w", err)
    }

    prompt, err := cr.template.Format(map[string]any{
        "code":     string(content),
        "filename": filename,
    })
    if err != nil {
        return fmt.Errorf("formatting prompt: %w", err)
    }

    ctx := context.Background()
    response, err := cr.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    })
    if err != nil {
        return fmt.Errorf("generating review: %w", err)
    }

    fmt.Printf("\n=== Review for %s ===\n", filename)
    fmt.Println(strings.Repeat("=", 80))
    fmt.Println(response.Choices[0].Content)
    fmt.Println(strings.Repeat("=", 80))

    return nil
}

func main() {
    var (
        file = flag.String("file", "", "Go file to review")
        dir  = flag.String("dir", "", "Directory to review (all .go files)")
        git  = flag.Bool("git", false, "Review files changed in git working directory")
    )
    flag.Parse()

    reviewer, err := NewCodeReviewer()
    if err != nil {
        log.Fatal(err)
    }

    switch {
    case *file != "":
        if err := reviewer.ReviewFile(*file); err != nil {
            log.Fatal(err)
        }
    case *dir != "":
        if err := reviewDirectory(reviewer, *dir); err != nil {
            log.Fatal(err)
        }
    case *git:
        if err := reviewGitChanges(reviewer); err != nil {
            log.Fatal(err)
        }
    default:
        fmt.Println("Usage:")
        fmt.Println("  code-reviewer -file=main.go")
        fmt.Println("  code-reviewer -dir=./pkg")
        fmt.Println("  code-reviewer -git")
        os.Exit(1)
    }
}

func reviewDirectory(reviewer *CodeReviewer, dir string) error {
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor/") {
            return reviewer.ReviewFile(path)
        }
        return nil
    })
}

func reviewGitChanges(reviewer *CodeReviewer) error {
    // This is a simplified version - you'd want to use a proper git library
    cmd := exec.Command("git", "diff", "--name-only", "HEAD")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("getting git changes: %w", err)
    }

    files := strings.Split(strings.TrimSpace(string(output)), "\n")
    for _, file := range files {
        if strings.HasSuffix(file, ".go") && file != "" {
            if err := reviewer.ReviewFile(file); err != nil {
                log.Printf("Error reviewing %s: %v", file, err)
            }
        }
    }
    return nil
}
```

## Step 3: Test with sample code

Create `sample.go` to test the reviewer:

```go
package main

import "fmt"

func badCode() {
    // This has several issues
    var users []string
    for i := 0; i < len(users); i++ {
        fmt.Println(users[i]) // potential index out of bounds
    }
    
    // String concatenation in loop
    var result string
    for i := 0; i < 1000; i++ {
        result += fmt.Sprintf("item-%d,", i)
    }
    
    // Ignoring errors
    file, _ := os.Open("nonexistent.txt")
    file.Read(make([]byte, 100))
}
```

## Step 4: Run the code reviewer

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
go run main.go -file=sample.go
```

## Step 5: Enhanced version with structured output

Create `reviewer.go` for more sophisticated analysis:

```go
package main

import (
    "encoding/json"
    "fmt"
    "go/ast"
    "go/parser"
    "go/token"
)

type Issue struct {
    Severity    string `json:"severity"`
    Type        string `json:"type"`
    Line        int    `json:"line"`
    Description string `json:"description"`
    Suggestion  string `json:"suggestion"`
}

type ReviewResult struct {
    Filename string  `json:"filename"`
    Issues   []Issue `json:"issues"`
    Score    int     `json:"score"` // 0-100
}

func (cr *CodeReviewer) ReviewFileStructured(filename string) (*ReviewResult, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("reading file: %w", err)
    }

    // Parse for line numbers
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
    if err != nil {
        return nil, fmt.Errorf("parsing Go file: %w", err)
    }

    template := prompts.NewPromptTemplate(`
Analyze this Go code and return a JSON response with this exact structure:

{
  "filename": "{{.filename}}",
  "issues": [
    {
      "severity": "critical|warning|suggestion",
      "type": "bug|performance|style|security",
      "line": 42,
      "description": "Detailed issue description",
      "suggestion": "How to fix this issue"
    }
  ],
  "score": 85
}

Code to analyze:
'''go
{{.code}}
'''

Focus on real issues. Score: 100 = perfect, 0 = many serious issues.`, 
        []string{"code", "filename"})

    prompt, err := template.Format(map[string]any{
        "code":     string(content),
        "filename": filename,
    })
    if err != nil {
        return nil, fmt.Errorf("formatting prompt: %w", err)
    }

    ctx := context.Background()
    response, err := cr.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    }, llms.WithJSONMode())
    if err != nil {
        return nil, fmt.Errorf("generating review: %w", err)
    }

    var result ReviewResult
    if err := json.Unmarshal([]byte(response.Choices[0].Content), &result); err != nil {
        return nil, fmt.Errorf("parsing JSON response: %w", err)
    }

    return &result, nil
}
```

## Step 6: Create a Git hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running AI code review..."
./code-reviewer -git
if [ $? -ne 0 ]; then
    echo "Code review found issues. Fix them or use --no-verify to skip."
    exit 1
fi
echo "Code review passed!"
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Advanced features

### Custom rules engine

Add specific checks for your codebase:

```go
type RuleEngine struct {
    rules []Rule
}

type Rule interface {
    Check(node ast.Node, fset *token.FileSet) []Issue
}

type NoGlobalVarsRule struct{}

func (r NoGlobalVarsRule) Check(node ast.Node, fset *token.FileSet) []Issue {
    var issues []Issue
    ast.Inspect(node, func(n ast.Node) bool {
        if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
            for _, spec := range genDecl.Specs {
                if valueSpec, ok := spec.(*ast.ValueSpec); ok {
                    pos := fset.Position(valueSpec.Pos())
                    issues = append(issues, Issue{
                        Severity:    "warning",
                        Type:        "style",
                        Line:        pos.Line,
                        Description: "Global variable found",
                        Suggestion:  "Consider using dependency injection or configuration structs",
                    })
                }
            }
        }
        return true
    })
    return issues
}
```

## Integration with CI/CD

Create `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o code-reviewer

FROM alpine:latest
RUN apk --no-cache add git
WORKDIR /root/
COPY --from=builder /app/code-reviewer .
ENTRYPOINT ["./code-reviewer"]
```

Use in GitHub Actions:

```yaml
name: AI Code Review
on: [pull_request]
jobs:
  review:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Run AI Code Review
      env:
        OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
      run: |
        docker build -t code-reviewer .
        docker run --rm -v $PWD:/code code-reviewer -dir=/code
```

## Next steps

- Add support for other languages
- Implement learning from feedback
- Create a web interface for team reviews
- Add integration with popular IDEs

This tutorial shows how LangChainGo can power practical development tools that go far beyond simple chatbots!