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
	flagFix          = flag.Bool("fix", false, "fix issues found by linters")
	flagPrepush      = flag.Bool("prepush", false, "run additional linters that need to pass before pushing to GitHub")
	flagTesting      = flag.Bool("testing", false, "run testing-specific linters (httprr patterns)")
	flagArchitecture = flag.Bool("architecture", false, "run architectural linters (dependency rules, interface patterns)")
	flagNoChdir      = flag.Bool("no-chdir", false, "don't automatically change to repository root directory")
	flagVerbose      = flag.Bool("v", false, "enable verbose output")
	flagVeryVerbose  = flag.Bool("vv", false, "enable very verbose output")
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

	// For architecture mode, run architectural linters
	if *flagArchitecture {
		if *flagVerbose {
			log.Println("running in architecture mode, checking architectural rules and patterns")
		}
		if err := checkArchitecturalRules(fix); err != nil {
			return fmt.Errorf("checkArchitecturalRules: %w", err)
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
	osGetenvIssues := 0
	notRecordingIssues := 0
	for _, issue := range allIssues {
		errorLines = append(errorLines, fmt.Sprintf("  - %s:%d: %s", issue.File, issue.Line, issue.Description))
		switch issue.Type {
		case IssueHardcodedToken:
			tokenIssues++
		case IssueRedundantCleanup:
			cleanupIssues++
		case IssueDirectOsGetenv:
			osGetenvIssues++
		case IssueNotRecordingPattern:
			notRecordingIssues++
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

	if osGetenvIssues > 0 {
		errorLines = append(errorLines, "Some files use direct os.Getenv() calls to check for API keys in tests.")
		errorLines = append(errorLines, "This should be replaced with httprr.SkipIfNoCredentialsAndRecordingMissing().")
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, "The correct pattern is:")
		errorLines = append(errorLines, "  httprr.SkipIfNoCredentialsAndRecordingMissing(t, \"API_KEY\")")
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, "  var opts []ProviderOption")
		errorLines = append(errorLines, "  opts = append(opts, WithModel(\"model-name\"))")
		errorLines = append(errorLines, "  if rr.Replaying() {")
		errorLines = append(errorLines, "      opts = append(opts, WithAPIKey(\"test-api-key\"))")
		errorLines = append(errorLines, "  }")
		errorLines = append(errorLines, "  provider, err := NewProvider(opts...)")
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, "If you need to keep the os.Getenv() call, add a trailing comment:")
		errorLines = append(errorLines, "  if os.Getenv(\"API_KEY\") == \"\" { // some explanation")
		errorLines = append(errorLines, "")
	}

	if notRecordingIssues > 0 {
		errorLines = append(errorLines, "Some files use !rr.Recording() instead of rr.Replaying().")
		errorLines = append(errorLines, "For better readability, use the positive condition:")
		errorLines = append(errorLines, "")
		errorLines = append(errorLines, "  if rr.Replaying() {")
		errorLines = append(errorLines, "      // replay-specific logic")
		errorLines = append(errorLines, "  }")
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
	IssueDirectOsGetenv
	IssueNotRecordingPattern
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
		ast:    node,
	}

	// Walk the AST and find issues
	ast.Walk(analyzer, node)

	return analyzer.issues, nil
}

// HttprrAnalyzer implements ast.Visitor to analyze httprr patterns.
type HttprrAnalyzer struct {
	fset             *token.FileSet
	file             string
	issues           []HttprrIssue
	currentFunc      *ast.FuncDecl
	httprOpenCall    ast.Node  // Track httprr.OpenForTest calls
	inRecordingCheck bool      // Track if we're inside a !rr.Recording() check
	ast              *ast.File // The parsed AST file for comment access
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
		// Check if this is a !rr.Recording() pattern that should use rr.Replaying()
		if a.isNotRecordingCheck(n) {
			a.checkNotRecordingUsage(n)
		}

		// Check if this is a !rr.Recording() check (for token usage analysis)
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

		// Check for os.Getenv calls
		if a.isOsGetenvCall(n) {
			a.checkOsGetenvUsage(n)
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

// isRecordingCheck checks if an if statement is checking !rr.Recording() or rr.Replaying()
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

	// Check for rr.Replaying()
	if call, ok := ifStmt.Cond.(*ast.CallExpr); ok {
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "rr" && sel.Sel.Name == "Replaying"
			}
		}
	}

	return false
}

// isNotRecordingCheck checks if an if statement is checking !rr.Recording() (for linting purposes)
func (a *HttprrAnalyzer) isNotRecordingCheck(ifStmt *ast.IfStmt) bool {
	// Only flag this in test functions
	if a.currentFunc == nil || !strings.HasPrefix(a.currentFunc.Name.Name, "Test") {
		return false
	}

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

// checkNotRecordingUsage suggests using rr.Replaying() instead of !rr.Recording()
func (a *HttprrAnalyzer) checkNotRecordingUsage(ifStmt *ast.IfStmt) {
	pos := a.fset.Position(ifStmt.Pos())
	a.issues = append(a.issues, HttprrIssue{
		File:        a.file,
		Line:        pos.Line,
		Type:        IssueNotRecordingPattern,
		Description: "use rr.Replaying() instead of !rr.Recording() for better readability",
		Node:        ifStmt,
	})
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

// isOsGetenvCall checks if a call expression is os.Getenv
func (a *HttprrAnalyzer) isOsGetenvCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "os" && sel.Sel.Name == "Getenv"
		}
	}
	return false
}

// checkOsGetenvUsage checks if an os.Getenv call should be replaced with httprr.SkipIfNoCredentialsAndRecordingMissing
func (a *HttprrAnalyzer) checkOsGetenvUsage(call *ast.CallExpr) {
	// Only report in test functions
	if a.currentFunc == nil || !strings.HasPrefix(a.currentFunc.Name.Name, "Test") {
		return
	}

	// Check if this is checking for an API key or credential
	if len(call.Args) > 0 {
		if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
			envVar := strings.Trim(lit.Value, `"'`)
			// Check if this looks like an API key or credential
			if a.isCredentialEnvVar(envVar) {
				// Check if there's a trailing comment on the same line
				if a.hasTrailingComment(call) {
					return // Skip reporting if comment is present
				}

				// Check if this os.Getenv is part of a condition that also checks rr.Recording()
				if a.isInRecordingCondition(call) {
					return // Skip reporting if it's in a Recording() condition (correct pattern)
				}

				pos := a.fset.Position(call.Pos())
				a.issues = append(a.issues, HttprrIssue{
					File:        a.file,
					Line:        pos.Line,
					Type:        IssueDirectOsGetenv,
					Description: fmt.Sprintf("direct os.Getenv(%q) call should use httprr.SkipIfNoCredentialsAndRecordingMissing", envVar),
					Node:        call,
				})
			}
		}
	}
}

// isCredentialEnvVar checks if an environment variable name looks like a credential
func (a *HttprrAnalyzer) isCredentialEnvVar(envVar string) bool {
	envVar = strings.ToUpper(envVar)
	credentialPatterns := []string{
		"API_KEY", "APIKEY", "TOKEN", "SECRET", "PASSWORD", "CREDENTIALS",
		"CLIENT_ID", "CLIENT_SECRET", "AUTH", "BEARER", "ACCESS_KEY",
	}

	for _, pattern := range credentialPatterns {
		if strings.Contains(envVar, pattern) {
			return true
		}
	}

	// Also check for specific known API key patterns in this codebase
	knownKeys := []string{
		"JINA_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY",
		"GOOGLE_API_KEY", "HUGGINGFACE_API_TOKEN", "HF_TOKEN",
		"COHERE_API_KEY", "MISTRAL_API_KEY", "VOYAGEAI_API_KEY",
	}

	for _, key := range knownKeys {
		if envVar == key {
			return true
		}
	}

	return false
}

// hasTrailingComment checks if there's a comment on the same line as the os.Getenv call
func (a *HttprrAnalyzer) hasTrailingComment(call *ast.CallExpr) bool {
	if a.ast == nil || a.ast.Comments == nil {
		return false
	}

	callLine := a.fset.Position(call.Pos()).Line

	// Check all comment groups for comments on the same line
	for _, commentGroup := range a.ast.Comments {
		for _, comment := range commentGroup.List {
			commentLine := a.fset.Position(comment.Pos()).Line
			if commentLine == callLine {
				return true // Found a comment on the same line
			}
		}
	}

	return false
}

// isInRecordingCondition checks if an os.Getenv call is part of a condition that also checks rr.Recording()
func (a *HttprrAnalyzer) isInRecordingCondition(call *ast.CallExpr) bool {
	// Walk up the AST to find the containing if statement or similar condition
	// This is a simplified implementation that looks for Recording() calls in the same expression tree

	// Find the containing statement by looking at the function body
	if a.currentFunc == nil || a.currentFunc.Body == nil {
		return false
	}

	// Check if this os.Getenv call is in an expression that also contains rr.Recording()
	callLine := a.fset.Position(call.Pos()).Line

	// Walk through the function statements to find the one containing our call
	for _, stmt := range a.currentFunc.Body.List {
		stmtStartLine := a.fset.Position(stmt.Pos()).Line
		stmtEndLine := a.fset.Position(stmt.End()).Line

		if callLine >= stmtStartLine && callLine <= stmtEndLine {
			// This statement contains our call, check if it also contains rr.Recording()
			return a.statementContainsRecordingCall(stmt)
		}
	}

	return false
}

// statementContainsRecordingCall checks if a statement contains a call to rr.Recording() or rr.Replaying()
func (a *HttprrAnalyzer) statementContainsRecordingCall(stmt ast.Stmt) bool {
	contains := false
	ast.Inspect(stmt, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "rr" && (sel.Sel.Name == "Recording" || sel.Sel.Name == "Replaying") {
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

	case *ast.IfStmt:
		// Handle !rr.Recording() patterns that should use rr.Replaying()
		f.fixNotRecordingPattern(n)

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

// fixNotRecordingPattern fixes !rr.Recording() patterns to use rr.Replaying() instead.
func (f *HttprrFixer) fixNotRecordingPattern(ifStmt *ast.IfStmt) {
	// Check if this if statement has a NotRecordingPattern issue
	ifPos := f.fset.Position(ifStmt.Pos())
	var hasIssue bool
	for _, issue := range f.issues {
		if issue.Type == IssueNotRecordingPattern && issue.Line == ifPos.Line {
			hasIssue = true
			break
		}
	}

	if !hasIssue {
		return
	}

	// Transform !rr.Recording() to rr.Replaying()
	if unary, ok := ifStmt.Cond.(*ast.UnaryExpr); ok {
		if unary.Op == token.NOT {
			if call, ok := unary.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok {
						if ident.Name == "rr" && sel.Sel.Name == "Recording" {
							// Replace !rr.Recording() with rr.Replaying()
							sel.Sel.Name = "Replaying"
							ifStmt.Cond = call // Remove the ! operator
						}
					}
				}
			}
		}
	}
}

// checkArchitecturalRules checks architectural rules and patterns in the codebase.
func checkArchitecturalRules(fix bool) error {
	if *flagVerbose {
		log.Println("Checking architectural rules using AST analysis")
	}

	// Collect all Go files in the repository
	var goFiles []string
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

		// Check all Go files including tests
		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	var allIssues []ArchitecturalIssue
	for _, file := range goFiles {
		// Skip certain directories that should be excluded from architectural checks
		if shouldSkipArchitecturalCheck(file) {
			continue
		}

		issues, err := checkArchitecturalRulesInFile(file)
		if err != nil {
			if *flagVerbose {
				log.Printf("Warning: failed to check architectural rules in %s: %v", file, err)
			}
			continue
		}
		allIssues = append(allIssues, issues...)
	}

	if len(allIssues) == 0 {
		if *flagVerbose {
			log.Println("No architectural rule violations found")
		}
		return nil
	}

	// Group and display issues
	issuesByType := make(map[ArchitecturalIssueType][]ArchitecturalIssue)
	for _, issue := range allIssues {
		issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
	}

	log.Printf("checkArchitecturalRules: Found %d architectural rule violations:", len(allIssues))

	// Sort and display issues by type
	for _, issueType := range []ArchitecturalIssueType{
		IssueDirectHttpClientUsage,
		IssueProviderCrossDependency,
		IssueMissingOptionsPattern,
		IssueInterfaceViolation,
		IssueTestPlacement,
		IssueInternalPackageUsage,
		IssueTestHttpClientUsage,
		IssueTestMissingHttprr,
		IssueTestHttpGetUsage,
		IssueTestHttpClientModification,
		IssueTestProviderConstructorWithoutHttprr,
		IssueTestHardcodedApiUrl,
		IssueTestSkipsHttprrCheck,
	} {
		if issues, exists := issuesByType[issueType]; exists {
			for _, issue := range issues {
				log.Printf("  - %s:%d: %s", issue.File, issue.Line, issue.Message)
			}
		}
	}

	// Provide guidance
	log.Println("")
	log.Println("Architectural Guidelines:")
	log.Println("1. HTTP Client Usage: Use httputil.DefaultClient instead of http.DefaultClient")
	log.Println("2. Provider Isolation: Providers should not depend on each other directly")
	log.Println("3. Options Pattern: All constructors should use functional options")
	log.Println("4. Interface Compliance: Types should implement expected interfaces")
	log.Println("5. Test Organization: Tests should be in the same package as the code")
	log.Println("6. Internal Package: Only parent packages should import internal/ packages")
	log.Println("7. Test HTTP Usage: Tests should use httprr.OpenForTest() or httputil.DefaultClient, not http.DefaultClient")
	log.Println("8. Test HTTP Recording: All test functions making HTTP calls should use httprr.OpenForTest()")
	log.Println("9. Test HTTP Methods: Avoid direct http.Get() calls in tests; use httprr pattern instead")
	log.Println("10. Test HTTP Client: Avoid modifying http.DefaultClient.Transport; use httprr client replacement")
	log.Println("11. Test Provider Constructors: HTTP-based provider constructors must use httprr.OpenForTest() for deterministic testing")
	log.Println("12. Test API URLs: Avoid hardcoded API URLs in tests; use httprr.OpenForTest() for HTTP mocking")
	log.Println("13. Test Httprr Bypassing: All HTTP calls should go through httprr; avoid bypassing httprr.SkipIfNoCredentialsAndRecordingMissing")
	log.Println("")

	if fix {
		log.Println("Note: Architectural fixes require manual intervention")
		log.Println("These rules help maintain clean architecture but cannot be auto-fixed")
	}

	// Return error to indicate violations were found
	return fmt.Errorf("found %d architectural rule violations", len(allIssues))
}

// ArchitecturalIssueType represents different types of architectural violations.
type ArchitecturalIssueType int

const (
	IssueDirectHttpClientUsage ArchitecturalIssueType = iota
	IssueProviderCrossDependency
	IssueMissingOptionsPattern
	IssueInterfaceViolation
	IssueTestPlacement
	IssueInternalPackageUsage
	IssueTestHttpClientUsage
	IssueTestMissingHttprr
	IssueTestHttpGetUsage
	IssueTestHttpClientModification
	IssueTestProviderConstructorWithoutHttprr
	IssueTestHardcodedApiUrl
	IssueTestSkipsHttprrCheck
)

// ArchitecturalIssue represents an architectural rule violation.
type ArchitecturalIssue struct {
	Type    ArchitecturalIssueType
	File    string
	Line    int
	Message string
	Node    ast.Node
}

// shouldSkipArchitecturalCheck determines if a file should be skipped for architectural checks.
func shouldSkipArchitecturalCheck(file string) bool {
	skipPaths := []string{
		"testing/",
		"examples/",
		"docs/",
		"scripts/",
		"internal/devtools/",
		".git/",
		"vendor/",
	}

	for _, skipPath := range skipPaths {
		if strings.Contains(file, skipPath) {
			return true
		}
	}
	return false
}

// checkArchitecturalRulesInFile checks architectural rules in a single Go file.
func checkArchitecturalRulesInFile(filePath string) ([]ArchitecturalIssue, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analyzer := &ArchitecturalAnalyzer{
		file:        filePath,
		fset:        fset,
		ast:         node,
		issues:      []ArchitecturalIssue{},
		usesHttprr:  false,
		usesTestify: false,
	}

	// Walk the AST and check for architectural violations
	ast.Walk(analyzer, node)

	return analyzer.issues, nil
}

// ArchitecturalAnalyzer implements ast.Visitor to detect architectural violations.
type ArchitecturalAnalyzer struct {
	file        string
	fset        *token.FileSet
	ast         *ast.File
	issues      []ArchitecturalIssue
	usesHttprr  bool // tracks if this file imports httprr
	usesTestify bool // tracks if this file is a test file using testify
}

// Visit implements ast.Visitor.
func (a *ArchitecturalAnalyzer) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.ImportSpec:
		a.checkImportUsage(n)
	case *ast.SelectorExpr:
		a.checkHttpClientUsage(n)
	case *ast.CallExpr:
		a.checkOptionsPatternUsage(n)
		a.checkTestHttpPatterns(n)
		a.checkProviderConstructorUsage(n)
	case *ast.FuncDecl:
		a.checkConstructorPattern(n)
		a.checkInterfaceCompliance(n)
		a.checkTestFunctionHttprr(n)
		a.checkTestProviderConstructors(n)
	case *ast.TypeSpec:
		a.checkTypeDefinition(n)
	}

	return a
}

// checkImportUsage checks for proper import patterns.
func (a *ArchitecturalAnalyzer) checkImportUsage(imp *ast.ImportSpec) {
	if imp.Path == nil {
		return
	}

	importPath := strings.Trim(imp.Path.Value, `"`)

	// Track imports for test analysis
	if strings.Contains(importPath, "/internal/httprr") {
		a.usesHttprr = true
	}
	if strings.Contains(importPath, "github.com/stretchr/testify") {
		a.usesTestify = true
	}

	// Check for internal package usage violations
	if strings.Contains(importPath, "/internal/") {
		if !a.isValidInternalImport(importPath) {
			pos := a.fset.Position(imp.Pos())
			a.issues = append(a.issues, ArchitecturalIssue{
				Type:    IssueInternalPackageUsage,
				File:    a.file,
				Line:    pos.Line,
				Message: fmt.Sprintf("invalid internal package import: %s", importPath),
				Node:    imp,
			})
		}
	}

	// Check for provider cross-dependencies
	if a.isProviderPackage() && a.isImportingOtherProvider(importPath) {
		pos := a.fset.Position(imp.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueProviderCrossDependency,
			File:    a.file,
			Line:    pos.Line,
			Message: fmt.Sprintf("provider should not import other provider: %s", importPath),
			Node:    imp,
		})
	}
}

// checkHttpClientUsage checks for direct usage of http.DefaultClient.
func (a *ArchitecturalAnalyzer) checkHttpClientUsage(sel *ast.SelectorExpr) {
	// Check for http.DefaultClient usage
	if ident, ok := sel.X.(*ast.Ident); ok {
		if ident.Name == "http" && sel.Sel.Name == "DefaultClient" {
			pos := a.fset.Position(sel.Pos())

			if a.isTestFile() {
				// Special rule for test files - provide specific guidance based on what they're already using
				var message string
				if a.usesHttprr {
					message = "test file already uses httprr - replace http.DefaultClient with rr.Client() from httprr.OpenForTest()"
				} else if a.usesTestify {
					message = "test file should use httprr.OpenForTest() for HTTP mocking instead of http.DefaultClient"
				} else {
					message = "tests should use httprr.OpenForTest() for HTTP mocking or httputil.DefaultClient, not http.DefaultClient"
				}

				a.issues = append(a.issues, ArchitecturalIssue{
					Type:    IssueTestHttpClientUsage,
					File:    a.file,
					Line:    pos.Line,
					Message: message,
					Node:    sel,
				})
			} else {
				// Regular files should use httputil.DefaultClient
				a.issues = append(a.issues, ArchitecturalIssue{
					Type:    IssueDirectHttpClientUsage,
					File:    a.file,
					Line:    pos.Line,
					Message: "use httputil.DefaultClient instead of http.DefaultClient",
					Node:    sel,
				})
			}
		}
	}
}

// checkOptionsPatternUsage checks for proper options pattern usage.
func (a *ArchitecturalAnalyzer) checkOptionsPatternUsage(call *ast.CallExpr) {
	// Check if this is a constructor call that should use options pattern
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if strings.HasPrefix(sel.Sel.Name, "New") {
			// Skip standard library constructors
			if a.isStandardLibraryCall(sel) {
				return
			}

			// This is a New* function call, check if it uses options pattern
			if !a.usesOptionsPattern(call) && a.shouldUseOptionsPattern(sel.Sel.Name) {
				pos := a.fset.Position(call.Pos())
				a.issues = append(a.issues, ArchitecturalIssue{
					Type:    IssueMissingOptionsPattern,
					File:    a.file,
					Line:    pos.Line,
					Message: fmt.Sprintf("constructor %s should use functional options pattern", sel.Sel.Name),
					Node:    call,
				})
			}
		}
	}
}

// checkConstructorPattern checks if constructors follow the expected pattern.
func (a *ArchitecturalAnalyzer) checkConstructorPattern(fn *ast.FuncDecl) {
	if fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "New") {
		return
	}

	// Check if this constructor has the right signature for options pattern
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return
	}

	// Look for variadic options parameter
	lastParam := fn.Type.Params.List[len(fn.Type.Params.List)-1]
	if lastParam.Type != nil {
		if ellipsis, ok := lastParam.Type.(*ast.Ellipsis); ok {
			if ident, ok := ellipsis.Elt.(*ast.Ident); ok {
				if ident.Name == "Option" {
					// This is good - uses options pattern
					return
				}
			}
		}
	}

	// If we get here, this New* function doesn't use options pattern
	if a.shouldUseOptionsPattern(fn.Name.Name) {
		pos := a.fset.Position(fn.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueMissingOptionsPattern,
			File:    a.file,
			Line:    pos.Line,
			Message: fmt.Sprintf("constructor %s should accept variadic ...Option parameter", fn.Name.Name),
			Node:    fn,
		})
	}
}

// checkInterfaceCompliance checks if types implement expected interfaces.
func (a *ArchitecturalAnalyzer) checkInterfaceCompliance(fn *ast.FuncDecl) {
	// This would check if LLM providers implement the Model interface,
	// if chains implement the Chain interface, etc.
	// Implementation would require more sophisticated type analysis
}

// checkTypeDefinition checks type definitions for architectural compliance.
func (a *ArchitecturalAnalyzer) checkTypeDefinition(typeSpec *ast.TypeSpec) {
	// Check if types follow naming conventions and patterns
	if typeSpec.Name == nil {
		return
	}

	typeName := typeSpec.Name.Name

	// Check if this is in a provider package and follows naming conventions
	if a.isProviderPackage() {
		if !a.followsProviderNamingConvention(typeName) {
			pos := a.fset.Position(typeSpec.Pos())
			a.issues = append(a.issues, ArchitecturalIssue{
				Type:    IssueInterfaceViolation,
				File:    a.file,
				Line:    pos.Line,
				Message: fmt.Sprintf("provider type %s should follow naming convention", typeName),
				Node:    typeSpec,
			})
		}
	}
}

// Helper methods for architectural analysis

func (a *ArchitecturalAnalyzer) isValidInternalImport(importPath string) bool {
	// Internal packages can only be imported by their parent package or siblings

	// For our project structure, these internal imports are allowed:
	// - /internal/* packages can be imported by any package at the root level
	// - Provider internal packages (e.g., /llms/openai/internal/*) can only be imported by their parent

	if strings.Contains(importPath, "github.com/tmc/langchaingo/internal/") {
		// Root level internal packages are allowed from anywhere in the project
		return true
	}

	// For provider-specific internal packages, check if we're in the same provider
	importParts := strings.Split(importPath, "/")

	// Find the "internal" part in the import path
	internalIndex := -1
	for i, part := range importParts {
		if part == "internal" {
			internalIndex = i
			break
		}
	}

	if internalIndex == -1 {
		return true // Not an internal import
	}

	// Check if we're in the same provider directory
	if internalIndex >= 2 {
		providerPath := strings.Join(importParts[:internalIndex], "/")
		fileDir := filepath.Dir(a.file)

		// Convert paths for comparison
		providerPath = strings.TrimPrefix(providerPath, "github.com/tmc/langchaingo/")
		fileDir = strings.TrimPrefix(fileDir, "./")

		return strings.HasPrefix(fileDir, providerPath)
	}

	return true
}

func (a *ArchitecturalAnalyzer) isProviderPackage() bool {
	return strings.Contains(a.file, "/llms/") ||
		strings.Contains(a.file, "/embeddings/") ||
		strings.Contains(a.file, "/vectorstores/")
}

func (a *ArchitecturalAnalyzer) isMainPackage() bool {
	return strings.Contains(a.file, "/chains/") ||
		strings.Contains(a.file, "/agents/") ||
		strings.Contains(a.file, "/memory/") ||
		strings.Contains(a.file, "/tools/")
}

func (a *ArchitecturalAnalyzer) isTestFile() bool {
	return strings.HasSuffix(a.file, "_test.go")
}

func (a *ArchitecturalAnalyzer) isImportingOtherProvider(importPath string) bool {
	// Check if this import is for another provider
	providerPaths := []string{"/llms/", "/embeddings/", "/vectorstores/"}

	for _, providerPath := range providerPaths {
		if strings.Contains(importPath, providerPath) {
			// This is importing a provider, check if it's a different one
			currentProviderPath := ""
			for _, path := range providerPaths {
				if strings.Contains(a.file, path) {
					currentProviderPath = path
					break
				}
			}

			if currentProviderPath != "" && strings.Contains(importPath, currentProviderPath) {
				// Extract provider names to compare
				currentProvider := a.extractProviderName(a.file, currentProviderPath)
				importedProvider := a.extractProviderName(importPath, providerPath)

				return currentProvider != importedProvider &&
					importedProvider != "" &&
					currentProvider != ""
			}
		}
	}

	return false
}

func (a *ArchitecturalAnalyzer) extractProviderName(path, providerType string) string {
	parts := strings.Split(path, providerType)
	if len(parts) < 2 {
		return ""
	}

	afterProvider := strings.TrimPrefix(parts[1], "/")
	providerParts := strings.Split(afterProvider, "/")
	if len(providerParts) > 0 {
		return providerParts[0]
	}

	return ""
}

func (a *ArchitecturalAnalyzer) usesOptionsPattern(call *ast.CallExpr) bool {
	// Check if the last argument is a variadic options argument
	if len(call.Args) == 0 {
		return false
	}

	// Look for ...opts or similar patterns in the arguments
	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok {
			if strings.Contains(strings.ToLower(ident.Name), "opt") {
				return true
			}
		}
	}

	return false
}

func (a *ArchitecturalAnalyzer) shouldUseOptionsPattern(functionName string) bool {
	// Only check provider constructors and main package constructors
	if !a.isProviderPackage() && !a.isMainPackage() {
		return false
	}

	// Constructor functions that should use options pattern
	constructorPatterns := []string{
		"NewJina",
		"NewOpenAI",
		"NewAnthropic",
		"NewGoogle",
		"NewHuggingFace",
		"NewBedrock",
		"NewOllama",
		"NewChroma",
		"NewPgVector",
		"NewPinecone",
		"NewQdrant",
		"NewWeaviate",
		"NewMilvus",
		"NewLLMChain",
		"NewConversation",
		"NewAPIChain",
		"NewSequentialChain",
		"NewSimpleSequentialChain",
		"NewRetrievalQA",
		"NewConversationalRetrievalQA",
		"NewSQLDatabaseChain",
		"NewMapReduceDocuments",
		"NewStuffDocuments",
		"NewRefineDocuments",
		"NewLLMMathChain",
		"NewTransform",
	}

	for _, pattern := range constructorPatterns {
		if functionName == pattern {
			return true
		}
	}

	// Only flag simple "New" if it's in a provider package
	if functionName == "New" && a.isProviderPackage() {
		return true
	}

	return false
}

func (a *ArchitecturalAnalyzer) followsProviderNamingConvention(typeName string) bool {
	// Provider types should not have generic names
	genericNames := []string{
		"Client",
		"Service",
		"Provider",
		"Handler",
		"Manager",
	}

	for _, generic := range genericNames {
		if typeName == generic {
			return false
		}
	}

	return true
}

// isStandardLibraryCall checks if a call is to the standard library or third-party packages.
func (a *ArchitecturalAnalyzer) isStandardLibraryCall(sel *ast.SelectorExpr) bool {
	if ident, ok := sel.X.(*ast.Ident); ok {
		// Check for common standard library packages that have New* constructors
		standardLibPackages := []string{
			"errors",   // errors.New()
			"context",  // context.New*()
			"sync",     // sync.New*()
			"http",     // http.New*()
			"fmt",      // fmt.New*()
			"strings",  // strings.New*()
			"bytes",    // bytes.New*()
			"time",     // time.New*()
			"crypto",   // crypto.New*()
			"hash",     // hash.New*()
			"regexp",   // regexp.New*()
			"template", // template.New*()
			"json",     // json.New*()
			"xml",      // xml.New*()
			"sql",      // sql.New*()
			"log",      // log.New*()
			"bufio",    // bufio.New*()
			"io",       // io.New*()
			"os",       // os.New*()
		}

		for _, pkg := range standardLibPackages {
			if ident.Name == pkg {
				return true
			}
		}

		// Check for third-party packages (those with dots in import paths)
		// This is a heuristic - most third-party packages don't need to follow our options pattern
		if strings.Contains(ident.Name, ".") {
			return true
		}
	}

	return false
}

// checkTestHttpPatterns checks for proper HTTP usage patterns in test files.
func (a *ArchitecturalAnalyzer) checkTestHttpPatterns(call *ast.CallExpr) {
	if !a.isTestFile() {
		return
	}

	// Check for http.Get() calls in test files
	if a.isHttpGetCall(call) {
		pos := a.fset.Position(call.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestHttpGetUsage,
			File:    a.file,
			Line:    pos.Line,
			Message: "test files should use httprr.OpenForTest() with rr.Client() instead of direct http.Get() calls",
			Node:    call,
		})
	}

	// Check for http.DefaultClient.Transport modifications
	if a.isHttpClientTransportModification(call) {
		pos := a.fset.Position(call.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestHttpClientModification,
			File:    a.file,
			Line:    pos.Line,
			Message: "avoid modifying http.DefaultClient.Transport; use httprr.OpenForTest() and replace httputil.DefaultClient instead",
			Node:    call,
		})
	}
}

// checkTestFunctionHttprr checks if test functions making HTTP calls properly use httprr.
func (a *ArchitecturalAnalyzer) checkTestFunctionHttprr(fn *ast.FuncDecl) {
	if !a.isTestFile() || fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
		return
	}

	if fn.Body == nil {
		return
	}

	// Check if this test function makes HTTP calls but doesn't use httprr
	hasHttpCalls := a.functionMakesHttpCalls(fn)
	hasHttprrUsage := a.functionUsesHttprr(fn)

	// Skip Pinecone tests - they legitimately skip httprr due to client limitations
	isPineconeTest := strings.Contains(a.file, "/pinecone/pinecone_test.go")

	if hasHttpCalls && !hasHttprrUsage && !isPineconeTest {
		pos := a.fset.Position(fn.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestMissingHttprr,
			File:    a.file,
			Line:    pos.Line,
			Message: fmt.Sprintf("test function %s makes HTTP calls but doesn't use httprr.OpenForTest()", fn.Name.Name),
			Node:    fn,
		})
	}
}

// isHttpGetCall checks if a call is http.Get().
func (a *ArchitecturalAnalyzer) isHttpGetCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name == "http" && sel.Sel.Name == "Get"
		}
	}
	return false
}

// isHttpClientTransportModification checks if a call modifies http.DefaultClient.Transport.
func (a *ArchitecturalAnalyzer) isHttpClientTransportModification(call *ast.CallExpr) bool {
	// This would be detected in assignment statements, but we can check for specific patterns
	// like direct assignments to http.DefaultClient.Transport
	return false // This is a simplified implementation
}

// functionMakesHttpCalls checks if a function makes HTTP calls.
func (a *ArchitecturalAnalyzer) functionMakesHttpCalls(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	// Skip unit tests that are clearly not making HTTP calls
	if strings.Contains(fn.Name.Name, "Unit") || strings.Contains(fn.Name.Name, "_Unit") {
		return false
	}

	// Skip Example tests - they're meant to demonstrate real usage
	if strings.HasPrefix(fn.Name.Name, "Example") || strings.Contains(fn.Name.Name, "_Example") {
		return false
	}

	// Special case: local LLM package tests don't make HTTP calls (they use local binaries)
	if a.isInPackage("local") {
		return false
	}

	// Special case: executor tests with test agents don't make HTTP calls
	if a.isInPackage("agents") && strings.Contains(fn.Name.Name, "TestExecutorWithErrorHandler") {
		return false
	}

	// Skip tests that use httptest.Server (local test server)
	if a.functionUsesHttptest(fn) {
		return false
	}

	hasHttpCalls := false
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			// Check for direct HTTP calls
			if a.isHttpGetCall(call) {
				hasHttpCalls = true
				return false
			}

			// Check for provider calls that likely make HTTP requests
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				methodName := sel.Sel.Name

				// Only flag these if they're called on a real provider instance
				// (not test instances)
				httpMethods := []string{
					"Call", "GenerateContent", "EmbedQuery", "EmbedDocuments",
					"Search", "AddDocuments", "SimilaritySearch",
				}
				for _, method := range httpMethods {
					if methodName == method {
						// Check if this is being called on a test instance
						if a.isCallOnTestInstance(sel, fn) {
							continue
						}
						hasHttpCalls = true
						return false
					}
				}
			}

			// Check for client.Get, client.Post, etc. on real clients
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if strings.Contains(ident.Name, "client") || strings.Contains(ident.Name, "Client") {
						httpMethods := []string{"Get", "Post", "Put", "Delete", "Do"}
						for _, method := range httpMethods {
							if sel.Sel.Name == method {
								hasHttpCalls = true
								return false
							}
						}
					}
				}
			}
		}
		return true
	})

	return hasHttpCalls
}

// functionUsesHttptest checks if a function uses httptest.Server.
func (a *ArchitecturalAnalyzer) functionUsesHttptest(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	usesHttptest := false
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "httptest" {
						usesHttptest = true
						return false
					}
				}
			}
		}
		return true
	})

	return usesHttptest
}

// isCallOnTestInstance checks if a method call is on a test instance (rather than a real provider).
func (a *ArchitecturalAnalyzer) isCallOnTestInstance(sel *ast.SelectorExpr, fn *ast.FuncDecl) bool {
	// Special case: chains.Call() is a package-level function, not an HTTP method call
	if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "chains" && sel.Sel.Name == "Call" {
		return true
	}

	// Check if the receiver identifier suggests it's a test instance
	if ident, ok := sel.X.(*ast.Ident); ok {
		testInstanceNames := []string{
			"testLLM", "testEmbedder", "testClient", "testProvider",
			"mockLLM", "mockEmbedder", "mockClient", "mockProvider", "mockCache",
			"fakeLLM", "fakeEmbedder", "fakeClient", "fakeProvider",
		}

		for _, testName := range testInstanceNames {
			if strings.Contains(strings.ToLower(ident.Name), strings.ToLower(testName)) {
				return true
			}
		}

		// Check for common patterns like "llm" variable created by mock functions
		if a.isVariableCreatedByMockFunction(ident.Name, fn) {
			return true
		}

		// Check if the variable is defined with a struct literal in the same function
		// This is a more sophisticated check that would require tracking variable assignments
		return a.isVariableDefinedAsTestStruct(ident.Name, fn)
	}

	return false
}

// isVariableCreatedByMockFunction checks if a variable is created by calling a mock/new function.
func (a *ArchitecturalAnalyzer) isVariableCreatedByMockFunction(varName string, fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	// Look for assignments like: llm := NewMockLLM(...) or cache := New(mockLLM, mockCache)
	for _, stmt := range fn.Body.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			for i, lhs := range assignStmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == varName {
					if i < len(assignStmt.Rhs) {
						if rhs := assignStmt.Rhs[i]; rhs != nil {
							// Check if it's a call to a function with "mock", "new", "fake" in the name
							if call, ok := rhs.(*ast.CallExpr); ok {
								if ident, ok := call.Fun.(*ast.Ident); ok {
									funcName := strings.ToLower(ident.Name)
									mockPatterns := []string{"mock", "new", "fake", "test"}
									for _, pattern := range mockPatterns {
										if strings.Contains(funcName, pattern) {
											return true
										}
									}
								}
								// Special case: Check for local.New() which creates local LLMs that don't make HTTP calls
								if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
									if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "local" && sel.Sel.Name == "New" {
										return true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

// isVariableDefinedAsTestStruct checks if a variable is defined as a test struct in the function.
func (a *ArchitecturalAnalyzer) isVariableDefinedAsTestStruct(varName string, fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	// Look for assignments like: testVar := &SomeStruct{...}
	for _, stmt := range fn.Body.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			for i, lhs := range assignStmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == varName {
					if i < len(assignStmt.Rhs) {
						if rhs := assignStmt.Rhs[i]; rhs != nil {
							// Check if it's a struct literal or address of struct literal
							switch rhsType := rhs.(type) {
							case *ast.UnaryExpr:
								if rhsType.Op == token.AND {
									if _, ok := rhsType.X.(*ast.CompositeLit); ok {
										return true
									}
								}
							case *ast.CompositeLit:
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

// isInPackage checks if the analyzer is currently analyzing a specific package.
func (a *ArchitecturalAnalyzer) isInPackage(packageName string) bool {
	return strings.Contains(a.file, "/"+packageName+"/") || strings.Contains(a.file, "/"+packageName+"_test.go")
}

// functionUsesHttprr checks if a function uses httprr.OpenForTest or calls helper functions that use httprr.
func (a *ArchitecturalAnalyzer) functionUsesHttprr(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	usesHttprr := false
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			// Direct httprr.OpenForTest usage
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "httprr" && sel.Sel.Name == "OpenForTest" {
						usesHttprr = true
						return false
					}
				}
			}

			// Helper function patterns that typically use httprr
			if ident, ok := call.Fun.(*ast.Ident); ok {
				httrrHelperPatterns := []string{
					"newHTTPRRClient", "newOpenAIEmbedder", "newTestClient",
					"newTestLLM", "newTestEmbedder", "setupHTTPRR",
					"createTestClient", "createHTTPRRClient", "newErnieTestClient",
					"newErnieTestLLM", "newPalmTestLLM",
				}
				for _, pattern := range httrrHelperPatterns {
					if ident.Name == pattern {
						usesHttprr = true
						return false
					}
				}
			}

			// Function calls that contain "httprr" or "HTTPRR" in the name
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if strings.Contains(strings.ToLower(ident.Name), "httprr") {
					usesHttprr = true
					return false
				}
			}
		}
		return true
	})

	return usesHttprr
}

// checkProviderConstructorUsage checks for provider constructor calls without proper httprr setup.
func (a *ArchitecturalAnalyzer) checkProviderConstructorUsage(call *ast.CallExpr) {
	if !a.isTestFile() {
		return
	}

	// Check for provider constructor calls
	if a.isProviderConstructorCall(call) {
		constructorName := a.getConstructorName(call)
		if constructorName == "" {
			return
		}

		// Skip vector store constructors - they typically don't make direct HTTP calls
		// but use embedders that should already have httprr setup
		if strings.Contains(constructorName, "vector") || strings.Contains(constructorName, "store") {
			return
		}

		// Skip Pinecone constructor calls - Pinecone client doesn't support custom HTTP clients
		if strings.Contains(constructorName, "pinecone.New") {
			return
		}

		// Check if this call is inside a function that uses httprr properly
		currentFunc := a.getCurrentFunction(call)
		if currentFunc == nil {
			return
		}

		if !a.functionUsesHttprr(currentFunc) && !a.functionSkipsHttprr(currentFunc) {
			pos := a.fset.Position(call.Pos())
			a.issues = append(a.issues, ArchitecturalIssue{
				Type:    IssueTestProviderConstructorWithoutHttprr,
				File:    a.file,
				Line:    pos.Line,
				Message: fmt.Sprintf("test calls %s without httprr.OpenForTest() setup - HTTP-based providers should use httprr for deterministic testing", constructorName),
				Node:    call,
			})
		}
	}

	// Check for hardcoded API URLs (but skip assert/equality checks and internal client tests)
	if a.containsHardcodedApiUrl(call) && !a.isAssertOrEqualsCall(call) &&
		!(strings.Contains(a.file, "/internal/") && strings.Contains(a.file, "client")) {
		pos := a.fset.Position(call.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestHardcodedApiUrl,
			File:    a.file,
			Line:    pos.Line,
			Message: "test contains hardcoded API URL - should use httprr.OpenForTest() for HTTP mocking",
			Node:    call,
		})
	}
}

// checkTestProviderConstructors checks if test functions properly use httprr with provider constructors.
func (a *ArchitecturalAnalyzer) checkTestProviderConstructors(fn *ast.FuncDecl) {
	if !a.isTestFile() || fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
		return
	}

	if fn.Body == nil {
		return
	}

	// Check if this test function bypasses httprr.SkipIfNoCredentialsAndRecordingMissing
	// Skip internal client tests and certain unit tests that are legitimate
	skipHttprrCheck := strings.Contains(a.file, "/internal/") && strings.Contains(a.file, "client") && strings.Contains(a.file, "_test.go")
	skipHttprrCheck = skipHttprrCheck || strings.Contains(a.file, "memory/token_buffer_test.go") // Unit test for token counting
	skipHttprrCheck = skipHttprrCheck || strings.Contains(a.file, "_unit_test.go")               // Unit tests

	// Skip specific constructor unit tests that don't make HTTP calls
	if fn.Name.Name == "TestNew" && (strings.Contains(a.file, "coherellm_test.go") ||
		strings.Contains(a.file, "huggingfacellm_test.go") || strings.Contains(a.file, "mistralmodel_test.go")) {
		skipHttprrCheck = true
	}

	// Skip HuggingFace provider tests - they're testing provider functionality
	if strings.Contains(a.file, "huggingfacellm_test.go") &&
		(fn.Name.Name == "TestHuggingFaceLLMWithProvider" || fn.Name.Name == "TestHuggingFaceLLMStandardInference") {
		skipHttprrCheck = true
	}

	// Skip Pinecone tests - Pinecone client doesn't support custom HTTP clients
	// so tests properly skip when credentials are not available instead of using httprr
	if strings.Contains(a.file, "/pinecone/pinecone_test.go") {
		skipHttprrCheck = true
	}

	if !skipHttprrCheck && a.functionSkipsHttprrCheck(fn) {
		pos := a.fset.Position(fn.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestSkipsHttprrCheck,
			File:    a.file,
			Line:    pos.Line,
			Message: fmt.Sprintf("test function %s bypasses httprr.SkipIfNoCredentialsAndRecordingMissing - all HTTP calls should go through httprr", fn.Name.Name),
			Node:    fn,
		})
	}

	// Check if function uses provider constructors but doesn't use httprr
	hasProviderConstructors := a.functionUsesProviderConstructors(fn)
	hasHttprrUsage := a.functionUsesHttprr(fn)
	skipsHttprr := a.functionSkipsHttprr(fn)

	// Skip Pinecone tests - they legitimately skip httprr due to client limitations
	isPineconeTest := strings.Contains(a.file, "/pinecone/pinecone_test.go")

	if hasProviderConstructors && !hasHttprrUsage && !skipsHttprr && !isPineconeTest {
		pos := a.fset.Position(fn.Pos())
		a.issues = append(a.issues, ArchitecturalIssue{
			Type:    IssueTestProviderConstructorWithoutHttprr,
			File:    a.file,
			Line:    pos.Line,
			Message: fmt.Sprintf("test function %s uses provider constructors but doesn't use httprr.OpenForTest()", fn.Name.Name),
			Node:    fn,
		})
	}
}

// isProviderConstructorCall checks if a call is a provider constructor.
func (a *ArchitecturalAnalyzer) isProviderConstructorCall(call *ast.CallExpr) bool {
	providerConstructors := []string{
		// LLM providers
		"jina.New", "ernie.New", "googleai.New", "anthropic.New", "openai.New",
		"mistral.New", "cohere.New", "huggingface.New", "ollama.New",
		"bedrock.New", "vertexai.New", "llamafile.New", "maritaca.New",
		"watsonx.New", "cloudflare.New", "perplexity.New", "groq.New",
		"deepseek.New", "nvidia.New",

		// Embedding providers
		"jina.NewEmbeddings", "openai.NewEmbeddings", "huggingface.NewEmbeddings",
		"bedrock.NewEmbeddings", "vertexai.NewEmbeddings", "voyageai.New",
		"googleai.NewEmbeddings", "mistral.NewEmbeddings",

		// Vector stores
		"chroma.New", "pinecone.New", "qdrant.New", "weaviate.New",
		"pgvector.New", "redisvector.New", "opensearch.New",
		"azureaisearch.New", "milvus.New", "mongovector.New",

		// Tool providers that make HTTP calls
		"duckduckgo.New", "serpapi.New", "wikipedia.New", "metaphor.New",
		"perplexity.New", "zapier.New",
	}

	constructorName := a.getConstructorName(call)
	for _, constructor := range providerConstructors {
		if constructorName == constructor {
			return true
		}
	}

	return false
}

// getConstructorName extracts the full constructor name from a call expression.
func (a *ArchitecturalAnalyzer) getConstructorName(call *ast.CallExpr) string {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok {
			return ident.Name + "." + sel.Sel.Name
		}
	}
	return ""
}

// getCurrentFunction finds the function containing the given node.
func (a *ArchitecturalAnalyzer) getCurrentFunction(node ast.Node) *ast.FuncDecl {
	// Walk up the AST to find the containing function
	var currentFunc *ast.FuncDecl
	ast.Inspect(a.ast, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			// Check if the node is within this function
			if fn.Pos() <= node.Pos() && node.Pos() <= fn.End() {
				currentFunc = fn
			}
		}
		return true
	})
	return currentFunc
}

// functionSkipsHttprr checks if a function explicitly skips httprr checks.
func (a *ArchitecturalAnalyzer) functionSkipsHttprr(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	// Skip Example tests - they're meant to demonstrate real usage
	if strings.HasPrefix(fn.Name.Name, "Example") || strings.Contains(fn.Name.Name, "_Example") {
		return true
	}

	// Look for patterns that indicate httprr is intentionally skipped
	skipsHttprr := false
	ast.Inspect(fn, func(node ast.Node) bool {
		// Check for t.Skip() calls
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if (ident.Name == "t" || ident.Name == "b") && sel.Sel.Name == "Skip" {
						skipsHttprr = true
						return false
					}
				}
			}
		}

		// Check for comments indicating skip
		if comment, ok := node.(*ast.Comment); ok {
			if strings.Contains(strings.ToLower(comment.Text), "skip") &&
				strings.Contains(strings.ToLower(comment.Text), "httprr") {
				skipsHttprr = true
				return false
			}
		}

		return true
	})

	return skipsHttprr
}

// functionSkipsHttprrCheck checks if a function bypasses httprr.SkipIfNoCredentialsAndRecordingMissing.
func (a *ArchitecturalAnalyzer) functionSkipsHttprrCheck(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	// Skip Example tests - they're meant to demonstrate real usage
	if strings.HasPrefix(fn.Name.Name, "Example") || strings.Contains(fn.Name.Name, "_Example") {
		return false
	}

	// Skip internal client tests - they test HTTP clients directly and have different patterns
	if strings.Contains(a.file, "/internal/") && strings.Contains(a.file, "client") && strings.Contains(a.file, "_test.go") {
		return false
	}

	// Look for early returns or skips that bypass httprr checks
	// First check if function properly calls httprr.SkipIfNoCredentialsAndRecordingMissing
	usesHttprrSkip := false
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "httprr" && sel.Sel.Name == "SkipIfNoCredentialsAndRecordingMissing" {
						usesHttprrSkip = true
						return false
					}
				}
			}
		}
		return true
	})

	// If function uses httprr.SkipIfNoCredentialsAndRecordingMissing, it's not bypassing
	if usesHttprrSkip {
		return false
	}

	bypassesCheck := false
	ast.Inspect(fn, func(node ast.Node) bool {
		// Check for environment variable checks that bypass httprr
		if call, ok := node.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					// Check for os.Getenv calls that might bypass httprr
					if ident.Name == "os" && sel.Sel.Name == "Getenv" {
						if len(call.Args) > 0 {
							if lit, ok := call.Args[0].(*ast.BasicLit); ok {
								envVar := strings.Trim(lit.Value, `"`)
								// Common API key environment variables
								apiKeyVars := []string{
									"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GOOGLE_API_KEY",
									"JINA_API_KEY", "COHERE_API_KEY", "MISTRAL_API_KEY",
									"HF_TOKEN", "HUGGINGFACEHUB_API_TOKEN", "BEDROCK_ACCESS_KEY",
								}
								for _, apiVar := range apiKeyVars {
									if envVar == apiVar {
										// Check if this is followed by a skip or return
										bypassesCheck = true
										return false
									}
								}
							}
						}
					}
				}
			}
		}

		return true
	})

	return bypassesCheck
}

// functionUsesProviderConstructors checks if a function uses provider constructors.
func (a *ArchitecturalAnalyzer) functionUsesProviderConstructors(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}

	hasConstructors := false
	ast.Inspect(fn, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if a.isProviderConstructorCall(call) {
				hasConstructors = true
				return false
			}
		}
		return true
	})

	return hasConstructors
}

// isAssertOrEqualsCall checks if a call is an assertion or equality check.
func (a *ArchitecturalAnalyzer) isAssertOrEqualsCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		// Check for assert.Equal, require.Equal, etc.
		if sel.Sel.Name == "Equal" || sel.Sel.Name == "Contains" || sel.Sel.Name == "NotEqual" {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "assert" || ident.Name == "require"
			}
		}
	}
	return false
}

// containsHardcodedApiUrl checks if a call contains hardcoded API URLs.
func (a *ArchitecturalAnalyzer) containsHardcodedApiUrl(call *ast.CallExpr) bool {
	hardcodedUrls := []string{
		"api.jina.ai",
		"generativelanguage.googleapis.com",
		"aip.baidubce.com",
		"api.openai.com",
		"api.anthropic.com",
		"api.cohere.ai",
		"api.mistral.ai",
		"api.together.xyz",
		"api.groq.com",
		"api.perplexity.ai",
		"inference-api.huggingface.co",
		"bedrock.amazonaws.com",
		"aiplatform.googleapis.com",
		"api.deepseek.com",
		"integrate.api.nvidia.com",
	}

	for _, arg := range call.Args {
		if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			url := strings.Trim(lit.Value, `"`)
			for _, hardcodedUrl := range hardcodedUrls {
				if strings.Contains(url, hardcodedUrl) {
					return true
				}
			}
		}
	}

	return false
}
