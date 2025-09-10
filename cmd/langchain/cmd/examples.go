package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Example struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	HasReadme   bool     `json:"has_readme"`
}

var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Browse and manage LangChain Go examples",
	Long: `Browse, list, and run the extensive collection of LangChain Go examples.
	
The examples cover various categories:
  • LLM Providers: OpenAI, Anthropic, Cohere, Groq, Google AI, etc.
  • Vector Stores: Chroma, Pinecone, Weaviate, MongoDB, etc.
  • Chains & Agents: Conversation, document Q&A, tool calling
  • Memory Systems: SQLite, Redis, in-memory storage
  • Integrations: AWS Bedrock, Google Cloud, Azure`,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available examples",
	Long: `List all available LangChain Go examples with descriptions and categories.
	
Use filters to narrow down results:
  langchain examples list --category llm
  langchain examples list --tag vision
  langchain examples list --format json`,
	RunE: listExamples,
}

var runCmd = &cobra.Command{
	Use:   "run <example-name>",
	Short: "Run a specific example",
	Long: `Run a specific example by name. The example will be executed with proper
environment setup and dependency resolution.

Example:
  langchain examples run openai-completion
  langchain examples run anthropic-tool-call`,
	Args: cobra.ExactArgs(1),
	RunE: runExample,
}

var (
	categoryFilter string
	tagFilter      string
	outputFormat   string
)

func init() {
	rootCmd.AddCommand(examplesCmd)
	examplesCmd.AddCommand(listCmd)
	examplesCmd.AddCommand(runCmd)

	listCmd.Flags().StringVar(&categoryFilter, "category", "", "Filter by category (llm, vectorstore, chain, etc.)")
	listCmd.Flags().StringVar(&tagFilter, "tag", "", "Filter by tag")
	listCmd.Flags().StringVar(&outputFormat, "format", "table", "Output format (table, json)")
}

func listExamples(cmd *cobra.Command, args []string) error {
	examples, err := discoverExamples()
	if err != nil {
		return fmt.Errorf("failed to discover examples: %w", err)
	}

	// Apply filters
	if categoryFilter != "" {
		filtered := make([]Example, 0)
		for _, ex := range examples {
			if strings.Contains(strings.ToLower(ex.Category), strings.ToLower(categoryFilter)) {
				filtered = append(filtered, ex)
			}
		}
		examples = filtered
	}

	if tagFilter != "" {
		filtered := make([]Example, 0)
		for _, ex := range examples {
			for _, tag := range ex.Tags {
				if strings.Contains(strings.ToLower(tag), strings.ToLower(tagFilter)) {
					filtered = append(filtered, ex)
					break
				}
			}
		}
		examples = filtered
	}

	// Output results
	switch outputFormat {
	case "json":
		return outputJSON(examples)
	default:
		return outputTable(examples)
	}
}

func runExample(cmd *cobra.Command, args []string) error {
	exampleName := args[0]

	examples, err := discoverExamples()
	if err != nil {
		return fmt.Errorf("failed to discover examples: %w", err)
	}

	var targetExample *Example
	for _, ex := range examples {
		if ex.Name == exampleName {
			targetExample = &ex
			break
		}
	}

	if targetExample == nil {
		return fmt.Errorf("example '%s' not found. Use 'langchain examples list' to see available examples", exampleName)
	}

	fmt.Printf("Running example: %s\n", targetExample.Name)
	fmt.Printf("Description: %s\n", targetExample.Description)
	fmt.Printf("Path: %s\n\n", targetExample.Path)

	// Change to example directory and run
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	err = os.Chdir(targetExample.Path)
	if err != nil {
		return fmt.Errorf("failed to change to example directory: %w", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			fmt.Printf("Warning: failed to restore original directory: %v\n", chdirErr)
		}
	}()

	// Check if there's a main.go or specific go file
	var runFile string
	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read example directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			runFile = entry.Name()
			break
		}
	}

	if runFile == "" {
		return fmt.Errorf("no Go file found in example directory")
	}

	// Check dependencies first
	if err := checkExampleDependencies(targetExample.Path); err != nil {
		fmt.Printf("Warning: Failed to check/download dependencies: %v\n", err)
	}

	// Execute the example
	return executeExample(targetExample.Path, runFile)
}

func discoverExamples() ([]Example, error) {
	var examples []Example
	examplesDir := "examples"

	err := filepath.WalkDir(examplesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != examplesDir {
			// Check if this directory contains Go files
			hasGoFile := false
			dirEntries, err := os.ReadDir(path)
			if err != nil {
				return nil // Skip directories we can't read
			}

			var hasReadme bool
			for _, entry := range dirEntries {
				if !entry.IsDir() {
					if strings.HasSuffix(entry.Name(), ".go") {
						hasGoFile = true
					}
					if strings.ToLower(entry.Name()) == "readme.md" {
						hasReadme = true
					}
				}
			}

			if hasGoFile {
				example := Example{
					Name:      filepath.Base(path),
					Path:      path,
					HasReadme: hasReadme,
				}

				// Categorize based on name patterns
				example.Category = categorizeExample(example.Name)
				example.Tags = extractTags(example.Name)
				example.Description = generateDescription(example.Name, example.Category)

				// Try to read description from README if available
				if hasReadme {
					if desc := readDescriptionFromReadme(path); desc != "" {
						example.Description = desc
					}
				}

				examples = append(examples, example)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort examples by name
	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Name < examples[j].Name
	})

	return examples, nil
}

func categorizeExample(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "openai") || strings.Contains(name, "anthropic") ||
		strings.Contains(name, "cohere") || strings.Contains(name, "groq") ||
		strings.Contains(name, "googleai") || strings.Contains(name, "ernie") ||
		strings.Contains(name, "deepseek") || strings.Contains(name, "huggingface") ||
		strings.Contains(name, "bedrock") || strings.Contains(name, "completion"):
		return "llm"
	case strings.Contains(name, "chroma") || strings.Contains(name, "vectorstore") ||
		strings.Contains(name, "pinecone") || strings.Contains(name, "weaviate") ||
		strings.Contains(name, "alloydb") || strings.Contains(name, "cloudsql"):
		return "vectorstore"
	case strings.Contains(name, "chain") || strings.Contains(name, "agent"):
		return "chain"
	case strings.Contains(name, "memory") || strings.Contains(name, "sqlite") ||
		strings.Contains(name, "zep"):
		return "memory"
	case strings.Contains(name, "embedding"):
		return "embedding"
	case strings.Contains(name, "tool") || strings.Contains(name, "zapier"):
		return "tool"
	case strings.Contains(name, "document") || strings.Contains(name, "qa"):
		return "document"
	default:
		return "other"
	}
}

func extractTags(name string) []string {
	var tags []string
	name = strings.ToLower(name)

	tagMap := map[string]string{
		"vision":     "vision",
		"tool":       "tool-calling",
		"stream":     "streaming",
		"cache":      "caching",
		"qa":         "question-answering",
		"chat":       "chat",
		"completion": "completion",
	}

	for keyword, tag := range tagMap {
		if strings.Contains(name, keyword) {
			tags = append(tags, tag)
		}
	}

	return tags
}

func generateDescription(name, category string) string {
	descriptions := map[string]string{
		"llm":         "Language model integration example",
		"vectorstore": "Vector database integration example",
		"chain":       "Chain or agent workflow example",
		"memory":      "Memory management example",
		"embedding":   "Text embedding example",
		"tool":        "External tool integration example",
		"document":    "Document processing example",
	}

	if desc, ok := descriptions[category]; ok {
		return desc
	}
	return "LangChain Go example"
}

func readDescriptionFromReadme(examplePath string) string {
	readmePath := filepath.Join(examplePath, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "```") {
			// Return first substantial non-header line as description
			if len(line) < 200 {
				return line
			}
		}
	}
	return ""
}

func outputJSON(examples []Example) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(examples)
}

func outputTable(examples []Example) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCATEGORY\tTAGS\tDESCRIPTION")
	fmt.Fprintln(w, "----\t--------\t----\t-----------")

	for _, ex := range examples {
		tags := strings.Join(ex.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ex.Name, ex.Category, tags, ex.Description)
	}

	return w.Flush()
}
