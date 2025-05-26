// Command lint runs various linters on the codebase.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

var (
	flagFix             = flag.Bool("fix", false, "fix issues found by linters")
	flagPrepush         = flag.Bool("prepush", false, "run additional linters that need to pass before pushing to GitHub")
	flagArchitecture    = flag.Bool("architecture", false, "run linters that verify the package architecture and dependency graph")
	flagNoChdir         = flag.Bool("no-chdir", false, "don't automatically change to repository root directory")
	flagVerbose         = flag.Bool("v", false, "enable verbose output")
	flagVeryVerbose     = flag.Bool("vv", false, "enable very verbose output")
	flagStrict          = flag.Bool("strict", false, "treat all architecture issues as errors, including known issues")
	flagKnownIssue      = flag.String("known-issue", "", "add a known issue to the list of issues that are only warnings (format: 'pkg imports otherpkg')")
	flagKnownIssueFile  = flag.String("known-issue-file", "", "path to file containing known issues, one per line")
	flagSaveKnownIssues = flag.String("save-known-issues", "", "path to file where to save known issues")
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

	// Load default known issues file if it exists
	defaultKnownIssuesFile := filepath.Join("internal", "devtools", "lint", "known_issues.txt")
	if _, err := os.Stat(defaultKnownIssuesFile); err == nil {
		if err := loadKnownIssuesFile(defaultKnownIssuesFile); err != nil {
			log.Printf("Warning: Failed to load default known issues file: %v", err)
		} else if *flagVerbose {
			log.Printf("Loaded default known issues from %s", defaultKnownIssuesFile)
		}
	}

	// Process additional known issues
	if *flagKnownIssue != "" {
		addKnownIssue(*flagKnownIssue)
		if *flagVerbose {
			log.Printf("Added known issue: %s", *flagKnownIssue)
		}
	}

	// Process specified known issues file (overrides default)
	if *flagKnownIssueFile != "" {
		// Clear previous known issues if loading a new file
		clearKnownIssues()
		if err := loadKnownIssuesFile(*flagKnownIssueFile); err != nil {
			log.Fatalf("Error loading known issues file: %v", err)
		}
	}

	// Save known issues if requested (do this before running checks)
	if *flagSaveKnownIssues != "" {
		if err := saveKnownIssuesToFile(*flagSaveKnownIssues); err != nil {
			log.Fatalf("Error saving known issues: %v", err)
		}
		if *flagVerbose {
			log.Printf("Saved known issues to %s", *flagSaveKnownIssues)
		}
		// If we're just saving known issues, exit successfully
		if !*flagArchitecture && !*flagPrepush && !*flagFix {
			return
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
	// For architecture mode, run checks related to code architecture
	if *flagArchitecture {
		log.Println("running in architecture mode, checking package dependencies and architecture")
		if *flagVeryVerbose {
			log.Println("running with very verbose debugging")
		}

		// Skip running tests when in architecture mode
		// Tests can be run separately with `go test`

		// Run the architecture checks
		// if err := checkArchitecture(fix); err != nil {
		return fmt.Errorf("checkArchitecture: not on this branch")
		// }
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
	for _, dir := range slices.Sorted(maps.Keys(allDirs)) {
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
	if err := os.WriteFile(gomodPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod in %s: %w", dir, err)
	}

	if *flagVerbose {
		log.Printf("removed replace directives from %s", gomodPath)
	}

	return nil
}

// Known issues tracking functions
var knownIssues = make(map[string]bool)

// loadKnownIssuesFile loads known issues from a file
func loadKnownIssuesFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			knownIssues[line] = true
		}
	}

	return nil
}

// addKnownIssue adds a known issue to the list
func addKnownIssue(issue string) {
	knownIssues[issue] = true
}

// clearKnownIssues clears the known issues list
func clearKnownIssues() {
	knownIssues = make(map[string]bool)
}

// saveKnownIssuesToFile saves known issues to a file
func saveKnownIssuesToFile(filename string) error {
	var issues []string
	for issue := range knownIssues {
		issues = append(issues, issue)
	}
	slices.Sort(issues)

	content := strings.Join(issues, "\n") + "\n"
	return os.WriteFile(filename, []byte(content), 0644)
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
			// Skip files in cmd/httprr-convert as those are for testing
			if !strings.Contains(path, "cmd/httprr-convert") {
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
			cmd := exec.Command("go", "run", "./cmd/httprr-convert", "-compress", "-dir", dir)
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
	errorLines = append(errorLines, "  ./internal/devtools/httprr-pack pack")

	return fmt.Errorf("%s", strings.Join(errorLines, "\n"))
}
