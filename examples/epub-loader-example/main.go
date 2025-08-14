package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xDezzy/langchaingo/documentloaders"
)

func main() {
	// Example 1: Load EPUB from file path in "single" mode (default)
	fmt.Println("=== Single Mode Example ===")
	singleLoader := documentloaders.NewEPUB("example.epub")
	singleDocs, err := singleLoader.Load(context.Background())
	if err != nil {
		log.Printf("Error loading EPUB in single mode: %v", err)
	} else {
		fmt.Printf("Loaded %d document(s) in single mode\n", len(singleDocs))
		for i, doc := range singleDocs {
			fmt.Printf("Document %d:\n", i+1)
			fmt.Printf("  Title: %v\n", doc.Metadata["title"])
			fmt.Printf("  Author: %v\n", doc.Metadata["author"])
			fmt.Printf("  Chapters: %v\n", doc.Metadata["chapters"])
			fmt.Printf("  Content length: %d characters\n", len(doc.PageContent))
			fmt.Printf("  Content preview: %.100s...\n\n", doc.PageContent)
		}
	}

	// Example 2: Load EPUB in "elements" mode (one document per chapter)
	fmt.Println("=== Elements Mode Example ===")
	elementsLoader := documentloaders.NewEPUB("example.epub", documentloaders.WithMode("elements"))
	elementDocs, err := elementsLoader.Load(context.Background())
	if err != nil {
		log.Printf("Error loading EPUB in elements mode: %v", err)
	} else {
		fmt.Printf("Loaded %d document(s) in elements mode\n", len(elementDocs))
		for i, doc := range elementDocs {
			fmt.Printf("Chapter %d:\n", i+1)
			fmt.Printf("  Chapter Title: %v\n", doc.Metadata["chapter_title"])
			fmt.Printf("  Chapter: %v of %v\n", doc.Metadata["chapter"], doc.Metadata["total_chapters"])
			fmt.Printf("  Content length: %d characters\n", len(doc.PageContent))
			fmt.Printf("  Content preview: %.100s...\n\n", doc.PageContent)
		}
	}

	// Example 3: Load from bytes
	fmt.Println("=== Load from Bytes Example ===")
	// In a real application, you would read EPUB file data into a byte slice
	// epubData, err := os.ReadFile("example.epub")
	// if err != nil {
	//     log.Fatal(err)
	// }
	// bytesLoader := documentloaders.NewEPUBFromBytes(epubData)
	// docs, err := bytesLoader.Load(context.Background())

	fmt.Println("Usage examples completed. Note: Actual EPUB file required for successful execution.")
}
