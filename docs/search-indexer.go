package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmparser "github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	meta "github.com/yuin/goldmark-meta"
)

// SearchEntry represents a single searchable item
type SearchEntry struct {
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Content     string            `json:"content"`
	Type        string            `json:"type"` // "doc", "function", "type", "package", etc.
	Package     string            `json:"package,omitempty"`
	Signature   string            `json:"signature,omitempty"`
	External    bool              `json:"external,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// SearchIndex represents the complete search index
type SearchIndex struct {
	Entries []SearchEntry `json:"entries"`
	Meta    IndexMeta     `json:"meta"`
}

// IndexMeta contains metadata about the search index
type IndexMeta struct {
	Generated   string `json:"generated"`
	DocCount    int    `json:"docCount"`
	SymbolCount int    `json:"symbolCount"`
	Version     string `json:"version"`
}

// Config holds configuration for the indexer
type Config struct {
	DocsDir     string
	SourceDir   string
	OutputFile  string
	BaseURL     string
	PkgGoDevURL string
	ModulePath  string
	Debug       bool
}

var (
	// Regex patterns for content cleaning
	codeBlockRegex = regexp.MustCompile("```[\\s\\S]*?```")
	headerRegex    = regexp.MustCompile("#{1,6}\\s+")
	linkRegex      = regexp.MustCompile("\\[([^\\]]+)\\]\\([^)]+\\)")
	whitespaceRegex = regexp.MustCompile("\\s+")
)

func main() {
	config := parseFlags()
	
	if config.Debug {
		log.Println("Starting search index generation...")
		log.Printf("Config: %+v", config)
	}

	indexer := NewIndexer(config)
	
	// Parse documentation files
	if err := indexer.ParseDocs(); err != nil {
		log.Fatalf("Error parsing docs: %v", err)
	}
	
	// Parse Go source code
	if err := indexer.ParseGoSource(); err != nil {
		log.Fatalf("Error parsing Go source: %v", err)
	}
	
	// Generate and write index
	if err := indexer.WriteIndex(); err != nil {
		log.Fatalf("Error writing index: %v", err)
	}
	
	log.Printf("Search index generated successfully: %s", config.OutputFile)
	log.Printf("Indexed %d documentation entries and %d Go symbols", 
		indexer.docCount, indexer.symbolCount)
}

func parseFlags() Config {
	config := Config{}
	
	flag.StringVar(&config.DocsDir, "docs", "./docs", "Path to documentation directory")
	flag.StringVar(&config.SourceDir, "source", "../", "Path to Go source code directory")
	flag.StringVar(&config.OutputFile, "output", "./build/search-index.json", "Output file path")
	flag.StringVar(&config.BaseURL, "base-url", "/langchaingo", "Base URL for documentation links")
	flag.StringVar(&config.PkgGoDevURL, "pkg-url", "https://pkg.go.dev/github.com/tmc/langchaingo", "pkg.go.dev base URL")
	flag.StringVar(&config.ModulePath, "module", "github.com/tmc/langchaingo", "Go module path")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug logging")
	
	flag.Parse()
	return config
}

// Indexer handles the search index generation
type Indexer struct {
	config      Config
	entries     []SearchEntry
	docCount    int
	symbolCount int
}

func NewIndexer(config Config) *Indexer {
	return &Indexer{
		config:  config,
		entries: make([]SearchEntry, 0),
	}
}

// ParseDocs processes all markdown documentation files
func (idx *Indexer) ParseDocs() error {
	if idx.config.Debug {
		log.Println("Parsing documentation files...")
	}

	// Initialize goldmark with frontmatter support
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
	)

	return filepath.WalkDir(idx.config.DocsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || (!strings.HasSuffix(path, ".md") && !strings.HasSuffix(path, ".mdx")) {
			return nil
		}

		if idx.config.Debug {
			log.Printf("Processing doc file: %s", path)
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading file %s: %w", path, err)
		}

		entry, err := idx.parseMarkdownFile(md, path, content)
		if err != nil {
			log.Printf("Warning: error parsing %s: %v", path, err)
			return nil // Continue processing other files
		}

		if entry != nil {
			idx.entries = append(idx.entries, *entry)
			idx.docCount++
		}

		return nil
	})
}

// parseMarkdownFile processes a single markdown file
func (idx *Indexer) parseMarkdownFile(md goldmark.Markdown, filePath string, content []byte) (*SearchEntry, error) {
	// Parse the markdown
	context := gmparser.NewContext()
	_ = md.Parser().Parse(text.NewReader(content), gmparser.WithContext(context))
	
	// Extract frontmatter metadata
	metaData := meta.Get(context)
	
	// Generate URL from file path
	relPath, err := filepath.Rel(idx.config.DocsDir, filePath)
	if err != nil {
		return nil, err
	}
	
	url := idx.generateDocURL(relPath)
	
	// Extract title from frontmatter or filename
	title := idx.extractTitle(metaData, relPath)
	
	// Clean and extract content
	cleanContent := idx.cleanMarkdownContent(string(content))
	
	// Extract keywords from content
	keywords := idx.extractKeywords(cleanContent)
	
	entry := &SearchEntry{
		Title:    title,
		URL:      url,
		Content:  cleanContent,
		Type:     "doc",
		Keywords: keywords,
		Metadata: convertMetadata(metaData),
	}
	
	return entry, nil
}

// generateDocURL creates a documentation URL from a file path
func (idx *Indexer) generateDocURL(relPath string) string {
	// Remove file extension
	urlPath := strings.TrimSuffix(relPath, filepath.Ext(relPath))
	
	// Handle index files
	if strings.HasSuffix(urlPath, "/index") {
		urlPath = strings.TrimSuffix(urlPath, "/index")
	}
	
	// Ensure it starts with /docs
	if !strings.HasPrefix(urlPath, "docs/") && urlPath != "docs" {
		urlPath = "docs/" + urlPath
	}
	
	return idx.config.BaseURL + "/" + urlPath
}

// extractTitle gets the title from frontmatter or generates from filename
func (idx *Indexer) extractTitle(metaData map[string]interface{}, relPath string) string {
	// Try frontmatter title
	if title, ok := metaData["title"].(string); ok && title != "" {
		return title
	}
	
	// Try sidebar_label
	if label, ok := metaData["sidebar_label"].(string); ok && label != "" {
		return label
	}
	
	// Generate from filename
	filename := filepath.Base(relPath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	
	if filename == "index" {
		// Use parent directory name
		dir := filepath.Dir(relPath)
		if dir != "." {
			filename = filepath.Base(dir)
		}
	}
	
	// Convert kebab-case to Title Case
	parts := strings.Split(filename, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	
	return strings.Join(parts, " ")
}

// cleanMarkdownContent removes markdown syntax and cleans content for search
func (idx *Indexer) cleanMarkdownContent(content string) string {
	// Remove frontmatter
	if strings.HasPrefix(content, "---") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) >= 3 {
			content = parts[2]
		}
	}
	
	// Remove code blocks
	content = codeBlockRegex.ReplaceAllString(content, "")
	
	// Remove headers markup
	content = headerRegex.ReplaceAllString(content, "")
	
	// Convert links to just text
	content = linkRegex.ReplaceAllString(content, "$1")
	
	// Normalize whitespace
	content = whitespaceRegex.ReplaceAllString(content, " ")
	
	// Trim and limit length
	content = strings.TrimSpace(content)
	if len(content) > 500 {
		content = content[:500] + "..."
	}
	
	return content
}

// extractKeywords extracts relevant keywords from content
func (idx *Indexer) extractKeywords(content string) []string {
	words := strings.Fields(strings.ToLower(content))
	keywordMap := make(map[string]int)
	
	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,!?;:")
		
		// Skip short words and common words
		if len(word) < 3 || isCommonWord(word) {
			continue
		}
		
		keywordMap[word]++
	}
	
	// Sort by frequency and take top keywords
	type wordFreq struct {
		word string
		freq int
	}
	
	var wordFreqs []wordFreq
	for word, freq := range keywordMap {
		wordFreqs = append(wordFreqs, wordFreq{word, freq})
	}
	
	sort.Slice(wordFreqs, func(i, j int) bool {
		return wordFreqs[i].freq > wordFreqs[j].freq
	})
	
	// Take top 10 keywords
	keywords := make([]string, 0, 10)
	for i, wf := range wordFreqs {
		if i >= 10 {
			break
		}
		keywords = append(keywords, wf.word)
	}
	
	return keywords
}

// ParseGoSource processes Go source code to extract symbols
func (idx *Indexer) ParseGoSource() error {
	if idx.config.Debug {
		log.Printf("Parsing Go source code from: %s", idx.config.SourceDir)
		log.Printf("Source directory absolute path: %s", filepath.Join(idx.config.SourceDir))
	}

	// Verify source directory exists
	if _, err := os.Stat(idx.config.SourceDir); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", idx.config.SourceDir)
	}

	fileSet := token.NewFileSet()
	goFileCount := 0
	processedFileCount := 0
	skippedDirs := make(map[string]int)
	
	err := filepath.WalkDir(idx.config.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if idx.config.Debug {
				log.Printf("Error accessing path %s: %v", path, err)
			}
			return err
		}

		// Skip non-Go files and certain directories
		if d.IsDir() {
			name := d.Name()
			// Skip specific directories, but allow ".." (parent directory)
			if name == ".git" || name == "vendor" || name == "node_modules" || name == "docs" || (strings.HasPrefix(name, ".") && name != "..") {
				if idx.config.Debug {
					log.Printf("Skipping directory: %s (reason: %s)", path, name)
				}
				skippedDirs[name]++
				return filepath.SkipDir
			}
			if idx.config.Debug {
				log.Printf("Entering directory: %s", path)
			}
			return nil
		}

		// Count Go files
		if strings.HasSuffix(path, ".go") {
			goFileCount++
			if strings.HasSuffix(path, "_test.go") {
				if idx.config.Debug {
					log.Printf("Skipping test file: %s", path)
				}
				return nil
			}
		} else {
			// Not a Go file
			return nil
		}

		if idx.config.Debug {
			log.Printf("Processing Go file: %s", path)
		}
		processedFileCount++

		if err := idx.parseGoFile(fileSet, path); err != nil {
			log.Printf("Error parsing Go file %s: %v", path, err)
			// Don't fail the entire process for one file
			return nil
		}

		return nil
	})

	if idx.config.Debug {
		log.Printf("Source parsing complete:")
		log.Printf("  Total Go files found: %d", goFileCount)
		log.Printf("  Go files processed: %d", processedFileCount)
		log.Printf("  Symbols extracted: %d", idx.symbolCount)
		log.Printf("  Skipped directories: %v", skippedDirs)
	}

	return err
}

// parseGoFile processes a single Go file
func (idx *Indexer) parseGoFile(fileSet *token.FileSet, filePath string) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Parse the Go file
	file, err := parser.ParseFile(fileSet, filePath, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", filePath, err)
	}

	// Create doc package
	pkg := &ast.Package{
		Name:  file.Name.Name,
		Files: map[string]*ast.File{filePath: file},
	}
	
	docPkg := doc.New(pkg, "", doc.AllDecls)
	
	// Generate package URL
	relPath, _ := filepath.Rel(idx.config.SourceDir, filepath.Dir(filePath))
	packagePath := idx.config.ModulePath
	if relPath != "." {
		packagePath = filepath.Join(packagePath, relPath)
	}
	
	// Index package
	if docPkg.Doc != "" {
		entry := SearchEntry{
			Title:     fmt.Sprintf("package %s", docPkg.Name),
			URL:       fmt.Sprintf("%s/%s", idx.config.PkgGoDevURL, strings.ReplaceAll(relPath, "\\", "/")),
			Content:   docPkg.Doc,
			Type:      "package",
			Package:   docPkg.Name,
			External:  true,
			Keywords:  []string{"package", docPkg.Name},
		}
		idx.entries = append(idx.entries, entry)
		idx.symbolCount++
	}
	
	// Index functions
	for _, fn := range docPkg.Funcs {
		entry := idx.createFunctionEntry(fn, packagePath, docPkg.Name)
		idx.entries = append(idx.entries, entry)
		idx.symbolCount++
	}
	
	// Index types
	for _, typ := range docPkg.Types {
		entry := idx.createTypeEntry(typ, packagePath, docPkg.Name)
		idx.entries = append(idx.entries, entry)
		idx.symbolCount++
		
		// Index type methods
		for _, method := range typ.Methods {
			entry := idx.createMethodEntry(method, typ.Name, packagePath, docPkg.Name)
			idx.entries = append(idx.entries, entry)
			idx.symbolCount++
		}
	}
	
	return nil
}

// createFunctionEntry creates a search entry for a function
func (idx *Indexer) createFunctionEntry(fn *doc.Func, packagePath, packageName string) SearchEntry {
	pkgURL := strings.TrimPrefix(packagePath, idx.config.ModulePath)
	if pkgURL != "" && !strings.HasPrefix(pkgURL, "/") {
		pkgURL = "/" + pkgURL
	}
	
	return SearchEntry{
		Title:     fn.Name,
		URL:       fmt.Sprintf("%s%s#%s", idx.config.PkgGoDevURL, pkgURL, fn.Name),
		Content:   fn.Doc,
		Type:      "function",
		Package:   packageName,
		Signature: formatFuncSignature(fn),
		External:  true,
		Keywords:  []string{"function", fn.Name, packageName},
	}
}

// createTypeEntry creates a search entry for a type
func (idx *Indexer) createTypeEntry(typ *doc.Type, packagePath, packageName string) SearchEntry {
	pkgURL := strings.TrimPrefix(packagePath, idx.config.ModulePath)
	if pkgURL != "" && !strings.HasPrefix(pkgURL, "/") {
		pkgURL = "/" + pkgURL
	}
	
	return SearchEntry{
		Title:     typ.Name,
		URL:       fmt.Sprintf("%s%s#%s", idx.config.PkgGoDevURL, pkgURL, typ.Name),
		Content:   typ.Doc,
		Type:      "type",
		Package:   packageName,
		Signature: formatTypeSignature(typ),
		External:  true,
		Keywords:  []string{"type", typ.Name, packageName},
	}
}

// createMethodEntry creates a search entry for a method
func (idx *Indexer) createMethodEntry(method *doc.Func, typeName, packagePath, packageName string) SearchEntry {
	pkgURL := strings.TrimPrefix(packagePath, idx.config.ModulePath)
	if pkgURL != "" && !strings.HasPrefix(pkgURL, "/") {
		pkgURL = "/" + pkgURL
	}
	
	return SearchEntry{
		Title:     fmt.Sprintf("%s.%s", typeName, method.Name),
		URL:       fmt.Sprintf("%s%s#%s.%s", idx.config.PkgGoDevURL, pkgURL, typeName, method.Name),
		Content:   method.Doc,
		Type:      "method",
		Package:   packageName,
		Signature: formatFuncSignature(method),
		External:  true,
		Keywords:  []string{"method", method.Name, typeName, packageName},
	}
}

// WriteIndex generates and writes the search index to file
func (idx *Indexer) WriteIndex() error {
	// For the search component, we need a flat array, not wrapped in an object
	// But we'll generate both formats for different use cases
	
	// Create metadata
	meta := IndexMeta{
		Generated:   fmt.Sprintf("%d", os.Getpid()), // Simple timestamp
		DocCount:    idx.docCount,
		SymbolCount: idx.symbolCount,
		Version:     "1.0",
	}
	
	// Full index structure for API/debugging
	searchIndex := SearchIndex{
		Entries: idx.entries,
		Meta:    meta,
	}
	
	// Ensure output directory exists
	outputDir := filepath.Dir(idx.config.OutputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	
	// Write full index file (for API/debugging)
	file, err := os.Create(idx.config.OutputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(searchIndex); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	
	// Write flat array version for search component
	searchFile := filepath.Join(outputDir, "search-index.json")
	flatFile, err := os.Create(searchFile)
	if err != nil {
		return fmt.Errorf("creating search file: %w", err)
	}
	defer flatFile.Close()
	
	flatEncoder := json.NewEncoder(flatFile)
	flatEncoder.SetIndent("", "  ")
	
	if err := flatEncoder.Encode(idx.entries); err != nil {
		return fmt.Errorf("encoding flat search JSON: %w", err)
	}
	
	return nil
}

// Helper functions

func convertMetadata(metaData map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range metaData {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

func formatFuncSignature(fn *doc.Func) string {
	if fn.Decl == nil || fn.Decl.Type == nil {
		return ""
	}
	
	// This is a simplified signature formatter
	// In a real implementation, you'd want more sophisticated formatting
	return fn.Name + "()"
}

func formatTypeSignature(typ *doc.Type) string {
	if len(typ.Decl.Specs) == 0 {
		return ""
	}
	
	// Simplified type signature
	return fmt.Sprintf("type %s", typ.Name)
}

func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "you": true, "that": true, "was": true,
		"for": true, "are": true, "with": true, "his": true, "they": true,
		"this": true, "have": true, "from": true, "not": true, "been": true,
		"can": true, "will": true, "use": true, "one": true, "all": true,
	}
	return commonWords[word]
}