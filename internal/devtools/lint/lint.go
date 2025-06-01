// Command lint runs various linters on the codebase.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var (
	flagFix         = flag.Bool("fix", false, "fix issues found by linters")
	flagPrepush     = flag.Bool("prepush", false, "run additional linters that need to pass before pushing to GitHub")
	flagTesting     = flag.Bool("testing", false, "run testing-specific linters (httprr patterns)")
	flagNoChdir     = flag.Bool("no-chdir", false, "don't automatically change to repository root directory")
	flagVerbose     = flag.Bool("v", false, "enable verbose output")
	flagVeryVerbose = flag.Bool("vv", false, "enable very verbose output")
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("lint: ")
	flag.Parse()

	// Auto change to repo root unless disabled
	if !*flagNoChdir {
		if err := changeToRepoRoot(); err != nil {
			log.Fatal(err)
		}
	}

	if err := run(*flagFix); err != nil {
		log.Fatal(err)
	}
}

// changeToRepoRoot attempts to find and change to the repository root directory
// by looking for a go.mod file with the root module.
func changeToRepoRoot() error {
	// First try to find go.mod in current directory
	if _, err := os.Stat("go.mod"); err == nil {
		// Check if this is the root module
		content, err := os.ReadFile("go.mod")
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "module github.com/tmc/langchaingo") &&
					!strings.HasPrefix(line, "module github.com/tmc/langchaingo/") {
					// Already at the root
					if *flagVerbose {
						log.Println("already at repository root")
					}
					return nil
				}
			}
		}
	}

	// Not at root, traverse upward to find it
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Keep track of original directory to report the change
	originalDir := currentDir

	// Try to find the repo root by going up until we find the root go.mod
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			content, err := os.ReadFile(filepath.Join(currentDir, "go.mod"))
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "module github.com/tmc/langchaingo") &&
						!strings.HasPrefix(line, "module github.com/tmc/langchaingo/") {
						// Found the root, change directory
						if err := os.Chdir(currentDir); err != nil {
							return fmt.Errorf("failed to change to repository root: %w", err)
						}
						if *flagVerbose {
							log.Printf("changed directory from %s to %s", originalDir, currentDir)
						}
						return nil
					}
				}
			}
		}

		// Go up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the filesystem root without finding the repo root
			return fmt.Errorf("could not find repository root")
		}
		currentDir = parentDir
	}
}

func run(fix bool) error {
	// For testing mode, run httprr pattern checks
	if *flagTesting {
		if *flagVerbose {
			log.Println("running in testing mode, checking test patterns and practices")
		}
		if err := checkHttprrTestPatterns(fix); err != nil {
			return fmt.Errorf("checkHttprrTestPatterns: %w", err)
		}
		return nil
	}

	// For prepush mode, run additional linters needed before pushing to GitHub
	if *flagPrepush {
		if *flagVerbose {
			log.Println("running in prepush mode, checking critical linters required before pushing")
		}
		if err := checkNoReplaceDirectives(fix); err != nil {
			return fmt.Errorf("checkNoReplaceDirectives: %w", err)
		}
		if err := checkHttprrCompression(fix); err != nil {
			return fmt.Errorf("checkHttprrCompression: %w", err)
		}
		// Additional pre-push checks can be added here in the future
		return nil
	}

	// Otherwise run all standard checks
	if err := checkMissingExampleGoModFiles(fix); err != nil {
		return fmt.Errorf("checkMissingExampleGoModFiles: %w", err)
	}
	if err := checkModuleNameMatchesDirectory(fix); err != nil {
		return fmt.Errorf("checkModuleNameMatchesDirectory: %w", err)
	}
	// Note: We don't check for replace directives in standard mode since they should
	// be absent for publishing. Use a separate command if you need them for development.
	return nil
}

// mapKeys returns a slice of keys from a map[string]bool
func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// checkMissingExampleGoModFiles checks if the example directories have go.mod files.
func checkMissingExampleGoModFiles(fix bool) error {
	examplePaths, _ := filepath.Glob("./examples/*/*.go")
	examplePathsWithMod, _ := filepath.Glob("./examples/*/go.mod")
	// Create maps of all directories and directories with go.mod
	allDirs := map[string]bool{}
	okDirs := map[string]bool{}
	for _, path := range examplePaths {
		allDirs[filepath.Dir(path)] = true
	}
	for _, path := range examplePathsWithMod {
		okDirs[filepath.Dir(path)] = true
		if *flagVerbose {
			log.Printf("found go.mod file in %s", filepath.Dir(path))
		}
	}
	// Find missing go.mod directories
	var errs []error
	dirs := mapKeys(allDirs)
	sort.Strings(dirs)
	for _, dir := range dirs {
		if _, ok := okDirs[dir]; !ok {
			if fix {
				errs = append(errs, fixMissingExampleGoModFile(dir))
			} else {
				errs = append(errs, fmt.Errorf("missing go.mod file in %s", dir))
			}
		}
	}
	return errors.Join(errs...)
}

func fixMissingExampleGoModFile(dir string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && go mod init && go mod tidy", dir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fixMissingExampleGoModFile failed: %w. Output: %s", err, output)
	}
	if *flagVerbose {
		log.Printf("fixed missing go.mod file in %s", dir)
	}
	return nil
}

// checkModuleNameMatchesDirectory checks if Go module names match their directory structure.
func checkModuleNameMatchesDirectory(fix bool) error {
	var errs []error
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a go.mod file
		if d.Name() != "go.mod" {
			return nil
		}

		// Skip root go.mod
		if path == "go.mod" {
			return nil
		}

		// Read go.mod content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse module line
		lines := strings.Split(string(content), "\n")
		var moduleName string
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}

		if moduleName == "" {
			errs = append(errs, fmt.Errorf("no module declaration in %s", path))
			return nil
		}

		// Get expected module name based on directory
		dir := filepath.Dir(path)
		expectedModuleName := fmt.Sprintf("github.com/tmc/langchaingo/%s", dir)

		// Handle examples differently
		if strings.Contains(dir, "examples/") {
			// Extract example name from directory path
			exampleName := filepath.Base(dir)

			// For examples, module should be github.com/tmc/langchaingo/examples/example-name
			// Or at least the basename should be consistent with directory name
			if !strings.HasSuffix(moduleName, exampleName) && !strings.Contains(moduleName, exampleName) {
				if *flagVeryVerbose {
					log.Printf("example module name mismatch in %s: directory is '%s', but module is '%s'",
						path, exampleName, moduleName)
				}

				if fix {
					// For examples, just use a simple module name based on directory
					expectedExampleModule := fmt.Sprintf("github.com/tmc/langchaingo/examples/%s", exampleName)
					errs = append(errs, fixModuleName(dir, expectedExampleModule))
				} else {
					errs = append(errs, fmt.Errorf("example module name mismatch in %s: directory is '%s', but module is '%s'",
						path, exampleName, moduleName))
				}
			} else if *flagVerbose {
				log.Printf("example module name ok in %s", path)
			}
			return nil
		}

		// For regular modules
		if moduleName != expectedModuleName {
			if *flagVeryVerbose {
				log.Printf("module name mismatch in %s: expected '%s', got '%s'", path, expectedModuleName, moduleName)
			}

			if fix {
				errs = append(errs, fixModuleName(dir, expectedModuleName))
			} else {
				errs = append(errs, fmt.Errorf("module name mismatch in %s: expected '%s', got '%s'", path, expectedModuleName, moduleName))
			}
		} else if *flagVerbose {
			log.Printf("module name ok in %s", path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return errors.Join(errs...)
}

func fixModuleName(dir, expectedModuleName string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && go mod edit -module=%s", dir, expectedModuleName))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fixModuleName failed: %w. Output: %s", err, output)
	}
	if *flagVerbose {
		log.Printf("fixed module name in %s to %s", dir, expectedModuleName)
	}
	return nil
}

// checkMissingReplaceDirectives checks if go.mod files have appropriate replace directives.
// It finds all go.mod files and ensures they have a replace directive pointing to the root module.
// This is critical for local development to ensure that changes in the root module are reflected
// in the dependent modules without requiring publishing a new version.
func checkMissingReplaceDirectives(fix bool) error {
	// Find all go.mod files
	var goModPaths []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a go.mod file
		if d.Name() != "go.mod" {
			return nil
		}

		// Skip root go.mod
		if path == "go.mod" {
			return nil
		}

		goModPaths = append(goModPaths, path)
		return nil
	})
	if err != nil {
		return err
	}

	// Get root module name
	rootModContent, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read root go.mod: %w", err)
	}

	rootModName := ""
	rootLines := strings.Split(string(rootModContent), "\n")
	for _, line := range rootLines {
		if strings.HasPrefix(line, "module ") {
			rootModName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			break
		}
	}

	if rootModName == "" {
		return fmt.Errorf("could not determine root module name")
	}

	if *flagVerbose {
		log.Printf("root module name: %s", rootModName)
		log.Printf("found %d go.mod files to check", len(goModPaths))
	}

	var errs []error

	// Check each go.mod file for replace directive
	for _, path := range goModPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read %s: %w", path, err))
			continue
		}

		lines := strings.Split(string(content), "\n")

		// Get module name from the go.mod file
		var moduleName string
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
				break
			}
		}

		if moduleName == "" {
			errs = append(errs, fmt.Errorf("no module declaration in %s", path))
			continue
		}

		// Check if it already has a replace directive for the root module
		hasReplace := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "replace "+rootModName+" => ") ||
				strings.HasPrefix(line, "replace "+rootModName+" v") {
				hasReplace = true
				break
			}
		}

		if !hasReplace {
			dir := filepath.Dir(path)
			if fix {
				if err := fixAddReplaceDirective(dir, rootModName); err != nil {
					errs = append(errs, err)
				}
			} else {
				errs = append(errs, fmt.Errorf("missing replace directive for root module in %s", path))
			}
		} else if *flagVerbose {
			log.Printf("replace directive found in %s", path)
		}
	}

	return errors.Join(errs...)
}

// fixAddReplaceDirective adds a replace directive to the go.mod file in the specified directory.
func fixAddReplaceDirective(dir, rootModName string) error {
	// Calculate relative path from module to root
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", dir, err)
	}

	absRoot, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get absolute path for root: %w", err)
	}

	// Calculate relative path from module to root directory
	relPath, err := filepath.Rel(absDir, absRoot)
	if err != nil {
		return fmt.Errorf("failed to get relative path from %s to root: %w", dir, err)
	}

	// Create the replace directive
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cd %s && go mod edit -replace=%s=%s", dir, rootModName, relPath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fixAddReplaceDirective failed for %s: %w. Output: %s", dir, err, output)
	}

	if *flagVerbose {
		log.Printf("added replace directive in %s: %s => %s", dir, rootModName, relPath)
	}

	return nil
}

// checkNoReplaceDirectives ensures that go.mod files do NOT contain replace directives.
// This is for pre-push checks to ensure published modules don't contain development replace directives.
func checkNoReplaceDirectives(fix bool) error {
	// Find all go.mod files
	var goModPaths []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a go.mod file
		if d.Name() != "go.mod" {
			return nil
		}

		// Skip root go.mod
		if path == "go.mod" {
			return nil
		}

		goModPaths = append(goModPaths, path)
		return nil
	})
	if err != nil {
		return err
	}

	if *flagVerbose {
		log.Printf("found %d go.mod files to check for replace directives", len(goModPaths))
	}

	var errs []error

	// Check each go.mod file for replace directives
	for _, path := range goModPaths {
		content, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read %s: %w", path, err))
			continue
		}

		lines := strings.Split(string(content), "\n")

		// Check for replace directives
		var replaceLines []string
		for i, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "replace ") {
				replaceLines = append(replaceLines, fmt.Sprintf("line %d: %s", i+1, line))
			}
		}

		if len(replaceLines) > 0 {
			if fix {
				if err := fixRemoveReplaceDirectives(filepath.Dir(path)); err != nil {
					errs = append(errs, err)
				}
			} else {
				errs = append(errs, fmt.Errorf("replace directives found in %s (should be removed before push):\n  %s",
					path, strings.Join(replaceLines, "\n  ")))
			}
		} else if *flagVerbose {
			log.Printf("no replace directives found in %s", path)
		}
	}

	return errors.Join(errs...)
}

// fixRemoveReplaceDirectives removes all replace directives from the go.mod file in the specified directory.
func fixRemoveReplaceDirectives(dir string) error {
	// Read the go.mod file
	gomodPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(gomodPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod in %s: %w", dir, err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	// Remove lines that start with "replace "
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "replace ") {
			newLines = append(newLines, line)
		}
	}

	// Write back the modified content
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(gomodPath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("failed to write go.mod in %s: %w", dir, err)
	}

	if *flagVerbose {
		log.Printf("removed replace directives from %s", gomodPath)
	}

	return nil
}

// checkHttprrCompression verifies that all httprr files are in compressed format.
// This ensures the repository stays clean and git grep doesn't match against large trace files.
func checkHttprrCompression(fix bool) error {
	if *flagVerbose {
		log.Println("Checking httprr file compression status")
	}

	var uncompressedFiles []string

	// Walk through the repository looking for .httprr files
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor":
				return fs.SkipDir
			}
			return nil
		}

		// Check for uncompressed httprr files
		if strings.HasSuffix(path, ".httprr") && !strings.HasSuffix(path, ".httprr.gz") {
			// Skip files in internal/devtools/httprr-convert as those are for testing
			if !strings.Contains(path, "internal/devtools/httprr-convert") {
				uncompressedFiles = append(uncompressedFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree: %w", err)
	}

	if len(uncompressedFiles) == 0 {
		if *flagVerbose {
			log.Println("All httprr files are properly compressed")
		}
		return nil
	}

	if fix {
		if *flagVerbose {
			log.Printf("Found %d uncompressed httprr files, compressing them...", len(uncompressedFiles))
		}

		// Group files by directory and compress them
		dirFiles := make(map[string][]string)
		for _, file := range uncompressedFiles {
			dir := filepath.Dir(file)
			dirFiles[dir] = append(dirFiles[dir], file)
		}

		for dir := range dirFiles {
			cmd := exec.Command("go", "run", "./internal/devtools/httprr-convert", "-compress", "-dir", dir)
			if *flagVerbose {
				log.Printf("Running: %s", strings.Join(cmd.Args, " "))
			}

			if output, err := cmd.CombinedOutput(); err != nil {
				log.Printf("Failed to compress httprr files in %s: %v\nOutput: %s", dir, err, output)
				return fmt.Errorf("failed to compress httprr files in %s: %w", dir, err)
			}
		}

		if *flagVerbose {
			log.Printf("Successfully compressed %d httprr files", len(uncompressedFiles))
		}
		return nil
	}

	// Report the issue without fixing
	var errorLines []string
	errorLines = append(errorLines, fmt.Sprintf("Found %d uncompressed httprr files:", len(uncompressedFiles)))
	for _, file := range uncompressedFiles {
		errorLines = append(errorLines, fmt.Sprintf("  - %s", file))
	}
	errorLines = append(errorLines, "")
	errorLines = append(errorLines, "To fix this issue, run:")
	errorLines = append(errorLines, "  go run ./internal/devtools/lint -prepush -fix")
	errorLines = append(errorLines, "Or manually compress files:")
	errorLines = append(errorLines, "  go run ./internal/devtools/rrtool pack -r")

	return fmt.Errorf("%s", strings.Join(errorLines, "\n"))
}

// checkHttprrTestPatterns checks for incorrect httprr usage patterns in test files using AST analysis.
// It identifies test files that use the old pattern where WithToken("test-api-key") is
// called unconditionally, even during recording mode, which causes authentication errors.
func checkHttprrTestPatterns(fix bool) error {
	if *flagVerbose {
		log.Println("Checking httprr test patterns using AST analysis")
	}

	var allIssues []HttprrIssue
	var fixedFiles []string

	// Walk through the repository looking for test files
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip certain directories
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor":
				return fs.SkipDir
			}
			return nil
		}

		// Only check Go test files
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse and analyze the file
		fileIssues, err := analyzeHttprrPatternsInFile(path)
		if err != nil {
			if *flagVerbose {
				log.Printf("Error analyzing %s: %v", path, err)
			}
			return nil // Skip files we can't parse
		}

		if len(fileIssues) == 0 {
			return nil // No issues in this file
		}

		allIssues = append(allIssues, fileIssues...)

		// If fixing is requested, attempt to fix this file
		if fix {
			if err := fixHttprrPatternsInFileAST(path, fileIssues); err != nil {
				log.Printf("Failed to fix httprr patterns in %s: %v", path, err)
			} else {
				fixedFiles = append(fixedFiles, path)
				if *flagVerbose {
					log.Printf("Fixed httprr patterns in %s", path)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory tree: %w", err)
	}

	if len(allIssues) == 0 {
		if *flagVerbose {
			log.Println("All httprr test patterns are correct")
		}
		return nil
	}

	if fix {
		if *flagVerbose && len(fixedFiles) > 0 {
			log.Printf("Successfully fixed httprr patterns in %d files:", len(fixedFiles))
			for _, file := range fixedFiles {
				log.Printf("  - %s", file)
			}
		}
		return nil
	}

	// Report the issues without fixing
	var errorLines []string
	errorLines = append(errorLines, fmt.Sprintf("Found %d httprr test pattern issues:", len(allIssues)))
	
	// Group issues by type
	tokenIssues := 0
	cleanupIssues := 0
	for _, issue := range allIssues {
		errorLines = append(errorLines, fmt.Sprintf("  - %s:%d: %s", issue.File, issue.Line, issue.Description))
		switch issue.Type {
		case IssueHardcodedToken:
			tokenIssues++
		case IssueRedundantCleanup:
			cleanupIssues++
		}
	}
	
	errorLines = append(errorLines, "")
	
	if tokenIssues > 0 {
		errorLines = append(errorLines, "Some files use the old httprr pattern where WithToken(\"test-api-key\") is called")
		errorLines = append(errorLines, "unconditionally, even during recording mode, which causes authentication errors.")
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, "The correct pattern is:")
		errorLines = append(errorLines, "  // Only add fake token when NOT recording (i.e., during replay)")
		errorLines = append(errorLines, "  if !rr.Recording() {")
		errorLines = append(errorLines, "      opts = append(opts, provider.WithToken(\"test-api-key\"))")
		errorLines = append(errorLines, "  }")
		errorLines = append(errorLines, "")
	}
	
	if cleanupIssues > 0 {
		errorLines = append(errorLines, "Some files have redundant cleanup calls. httprr.OpenForTest automatically")
		errorLines = append(errorLines, "handles cleanup, so t.Cleanup(func() { rr.Close() }) is not needed.")
		errorLines = append(errorLines, "")
	}
	
	errorLines = append(errorLines, "To fix these issues automatically, run:")
	errorLines = append(errorLines, "  go run ./internal/devtools/lint -testing -fix")

	return fmt.Errorf("%s", strings.Join(errorLines, "\n"))
}

// HttprrIssue represents a specific httprr pattern issue found in a file.
type HttprrIssue struct {
	File        string
	Line        int
	Type        HttprrIssueType
	Description string
	Node        ast.Node // The AST node that has the issue
}

// HttprrIssueType represents the type of httprr issue.
type HttprrIssueType int

const (
	IssueHardcodedToken HttprrIssueType = iota
	IssueParallelBeforeHttprr
	IssueRedundantCleanup
)

// analyzeHttprrPatternsInFile uses AST analysis to find httprr pattern issues in a file.
func analyzeHttprrPatternsInFile(filePath string) ([]HttprrIssue, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Skip files that don't use httprr
	if !strings.Contains(string(content), "httprr.OpenForTest") {
		return nil, nil
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analyzer := &HttprrAnalyzer{
		fset:   fset,
		file:   filePath,
		issues: []HttprrIssue{},
	}

	// Walk the AST and find issues
	ast.Walk(analyzer, node)

	return analyzer.issues, nil
}

// HttprrAnalyzer implements ast.Visitor to analyze httprr patterns.
type HttprrAnalyzer struct {
	fset          *token.FileSet
	file          string
	issues        []HttprrIssue
	currentFunc   *ast.FuncDecl
	httprOpenCall ast.Node // Track httprr.OpenForTest calls
	inRecordingCheck bool   // Track if we're inside a !rr.Recording() check
}

// Visit implements ast.Visitor.
func (a *HttprrAnalyzer) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		// Track current function for context
		a.currentFunc = n
		
		// Check for t.Parallel() calls in test functions
		if strings.HasPrefix(n.Name.Name, "Test") && len(n.Type.Params.List) > 0 {
			a.checkParallelUsage(n)
		}

	case *ast.IfStmt:
		// Check if this is a !rr.Recording() check
		if a.isRecordingCheck(n) {
			oldInCheck := a.inRecordingCheck
			a.inRecordingCheck = true
			// Visit the body of the if statement
			ast.Walk(a, n.Body)
			a.inRecordingCheck = oldInCheck
			// Don't visit the else clause
			return nil
		}

	case *ast.CallExpr:
		// Check for httprr.OpenForTest calls
		if a.isHttprrOpenForTest(n) {
			a.httprOpenCall = n
		}
		
		// Check for hardcoded WithToken/WithAPIKey calls
		if a.isWithTokenCall(n) {
			a.checkTokenUsage(n)
		}
		
		// Check for t.Cleanup calls
		if a.isCleanupCall(n) {
			a.checkCleanupUsage(n)
		}
	}

	return a
}

// isHttprrOpenForTest checks if a call expression is httprr.OpenForTest.
func (a *HttprrAnalyzer) isHttprrOpenForTest(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "httprr" && sel.Sel.Name == "OpenForTest"
		}
	}
	return false
}

// isWithTokenCall checks if a call expression is a WithToken or WithAPIKey call with "test-api-key".
func (a *HttprrAnalyzer) isWithTokenCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		methodName := sel.Sel.Name
		if methodName == "WithToken" || methodName == "WithAPIKey" {
			// Check if the argument is "test-api-key"
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok {
					return lit.Kind == token.STRING && 
						   (lit.Value == `"test-api-key"` || lit.Value == "'test-api-key'")
				}
			}
		}
	}
	return false
}

// isRecordingCheck checks if an if statement is checking !rr.Recording()
func (a *HttprrAnalyzer) isRecordingCheck(ifStmt *ast.IfStmt) bool {
	// Check for !rr.Recording()
	if unary, ok := ifStmt.Cond.(*ast.UnaryExpr); ok && unary.Op == token.NOT {
		if call, ok := unary.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					return ident.Name == "rr" && sel.Sel.Name == "Recording"
				}
			}
		}
	}
	return false
}

// checkTokenUsage checks if a WithToken call is properly wrapped in a Recording() check.
func (a *HttprrAnalyzer) checkTokenUsage(call *ast.CallExpr) {
	// Only report an issue if we're NOT inside a !rr.Recording() check
	if !a.inRecordingCheck {
		pos := a.fset.Position(call.Pos())
		a.issues = append(a.issues, HttprrIssue{
			File:        a.file,
			Line:        pos.Line,
			Type:        IssueHardcodedToken,
			Description: "hardcoded test token without Recording() check",
			Node:        call,
		})
	}
}

// checkParallelUsage checks for t.Parallel() calls that occur before httprr setup.
func (a *HttprrAnalyzer) checkParallelUsage(fn *ast.FuncDecl) {
	if fn.Body == nil {
		return
	}

	var parallelCall ast.Node
	var httprCall ast.Node

	// Walk through statements to find t.Parallel() and httprr.OpenForTest
	for _, stmt := range fn.Body.List {
		ast.Inspect(stmt, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				// Check for t.Parallel()
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name == "t" && sel.Sel.Name == "Parallel" {
							parallelCall = call
						}
					}
				}
				
				// Check for httprr.OpenForTest
				if a.isHttprrOpenForTest(call) {
					httprCall = call
				}
			}
			return true
		})
		
		// If we found both, check the order
		if parallelCall != nil && httprCall != nil {
			parallelPos := a.fset.Position(parallelCall.Pos())
			httprPos := a.fset.Position(httprCall.Pos())
			
			// If t.Parallel() comes before httprr.OpenForTest, it's an issue
			if parallelPos.Line < httprPos.Line {
				a.issues = append(a.issues, HttprrIssue{
					File:        a.file,
					Line:        parallelPos.Line,
					Type:        IssueParallelBeforeHttprr,
					Description: "t.Parallel() called before httprr setup, should be conditional on !rr.Recording()",
					Node:        parallelCall,
				})
			}
			break
		}
	}
}

// isCleanupCall checks if a call expression is t.Cleanup
func (a *HttprrAnalyzer) isCleanupCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "t" && sel.Sel.Name == "Cleanup"
		}
	}
	return false
}

// checkCleanupUsage checks if a t.Cleanup call contains redundant rr.Close()
func (a *HttprrAnalyzer) checkCleanupUsage(call *ast.CallExpr) {
	// Skip if we haven't seen an httprr.OpenForTest call yet
	if a.httprOpenCall == nil {
		return
	}
	
	// Check if the cleanup function contains rr.Close()
	if len(call.Args) > 0 {
		if fn, ok := call.Args[0].(*ast.FuncLit); ok {
			if fn.Body != nil {
				for _, stmt := range fn.Body.List {
					if a.containsRRClose(stmt) {
						pos := a.fset.Position(call.Pos())
						a.issues = append(a.issues, HttprrIssue{
							File:        a.file,
							Line:        pos.Line,
							Type:        IssueRedundantCleanup,
							Description: "redundant t.Cleanup(func() { rr.Close() }) - httprr.OpenForTest handles cleanup automatically",
							Node:        call,
						})
						return
					}
				}
			}
		}
	}
}

// containsRRClose checks if a statement contains rr.Close()
func (a *HttprrAnalyzer) containsRRClose(stmt ast.Stmt) bool {
	contains := false
	ast.Inspect(stmt, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "rr" && sel.Sel.Name == "Close" {
						contains = true
						return false
					}
				}
			}
		}
		return true
	})
	return contains
}

// fixHttprrPatternsInFileAST attempts to automatically fix httprr patterns in a file using AST manipulation.
func fixHttprrPatternsInFileAST(filePath string, issues []HttprrIssue) error {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return err
	}

	// Apply fixes to the AST
	fixer := &HttprrFixer{
		fset:   fset,
		issues: issues,
	}

	// Walk the AST and apply fixes
	ast.Walk(fixer, node)

	// Format and write back the modified AST
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return fmt.Errorf("failed to format modified AST: %w", err)
	}

	// Write the fixed content back to the file
	if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write fixed file: %w", err)
	}

	return nil
}

// HttprrFixer implements ast.Visitor to fix httprr patterns.
type HttprrFixer struct {
	fset   *token.FileSet
	issues []HttprrIssue
}

// containsRRClose checks if a statement contains rr.Close()
func (f *HttprrFixer) containsRRClose(stmt ast.Stmt) bool {
	contains := false
	ast.Inspect(stmt, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "rr" && sel.Sel.Name == "Close" {
						contains = true
						return false
					}
				}
			}
		}
		return true
	})
	return contains
}

// Visit implements ast.Visitor.
func (f *HttprrFixer) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.FuncDecl:
		// Handle test functions that need parallel execution fixes
		if strings.HasPrefix(n.Name.Name, "Test") && len(n.Type.Params.List) > 0 {
			f.fixParallelUsageInFunction(n)
			f.fixRedundantCleanupInFunction(n)
		}

	case *ast.CallExpr:
		// Handle hardcoded token issues by wrapping them in Recording() checks
		f.fixTokenUsageInCall(n)
	}

	return f
}

// fixParallelUsageInFunction fixes t.Parallel() calls in test functions.
func (f *HttprrFixer) fixParallelUsageInFunction(fn *ast.FuncDecl) {
	if fn.Body == nil {
		return
	}

	// Find issues that relate to this function
	fnPos := f.fset.Position(fn.Pos())
	var parallelIssues []HttprrIssue
	for _, issue := range f.issues {
		if issue.Type == IssueParallelBeforeHttprr && issue.File == fnPos.Filename {
			parallelIssues = append(parallelIssues, issue)
		}
	}

	if len(parallelIssues) == 0 {
		return
	}

	// Find and remove t.Parallel() calls
	var newStmts []ast.Stmt
	for _, stmt := range fn.Body.List {
		keep := true
		
		// Check if this statement contains t.Parallel()
		ast.Inspect(stmt, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name == "t" && sel.Sel.Name == "Parallel" {
							keep = false
							return false
						}
					}
				}
			}
			return true
		})
		
		if keep {
			newStmts = append(newStmts, stmt)
		}
	}

	fn.Body.List = newStmts
}

// fixRedundantCleanupInFunction removes redundant t.Cleanup(func() { rr.Close() }) calls.
func (f *HttprrFixer) fixRedundantCleanupInFunction(fn *ast.FuncDecl) {
	if fn.Body == nil {
		return
	}
	
	// Find cleanup issues that relate to this function
	fnPos := f.fset.Position(fn.Pos())
	var cleanupIssues []HttprrIssue
	for _, issue := range f.issues {
		if issue.Type == IssueRedundantCleanup && issue.File == fnPos.Filename {
			// Check if the issue is within this function
			issuePos := f.fset.Position(issue.Node.Pos())
			if issuePos.Line >= fnPos.Line && issuePos.Line <= f.fset.Position(fn.End()).Line {
				cleanupIssues = append(cleanupIssues, issue)
			}
		}
	}
	
	if len(cleanupIssues) == 0 {
		return
	}
	
	// Remove the redundant cleanup statements
	var newStmts []ast.Stmt
	for _, stmt := range fn.Body.List {
		keep := true
		
		// Check if this statement is a redundant cleanup call
		if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := exprStmt.X.(*ast.CallExpr); ok {
				// Check if this is a t.Cleanup call with rr.Close()
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name == "t" && sel.Sel.Name == "Cleanup" && f.containsRRClose(exprStmt) {
							stmtPos := f.fset.Position(stmt.Pos())
							for _, issue := range cleanupIssues {
								if issue.Line == stmtPos.Line {
									keep = false
									break
								}
							}
						}
					}
				}
			}
		}
		
		if keep {
			newStmts = append(newStmts, stmt)
		}
	}
	
	fn.Body.List = newStmts
}

// fixTokenUsageInCall fixes hardcoded token issues by modifying the AST structure.
// Note: This is a simplified implementation. A more sophisticated approach would
// analyze the containing context and properly wrap token calls in Recording() checks.
func (f *HttprrFixer) fixTokenUsageInCall(call *ast.CallExpr) {
	// For now, we'll focus on the more common parallel execution issues
	// Token usage issues are more complex to fix automatically since they require
	// restructuring the code to add conditional logic
	
	// This could be implemented by:
	// 1. Finding the assignment statement containing the WithToken call
	// 2. Converting it to an if-else structure
	// 3. Adding the appropriate Recording() check
	// 
	// This level of AST manipulation is complex and may be better left for manual fixing
	// or a more sophisticated code transformation tool
}
