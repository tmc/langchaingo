package textsplitter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func TestMarkdownHeaderTextSplitter_CodeBlockHeaders(t *testing.T) {
	t.Parallel()

	markdown := `# Real Header 1

Content here.

` + "```python" + `
# This is NOT a header - it's a comment
def foo():
    pass
` + "```" + `

## Real Header 2

More content.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	assert.Equal(t, "/", docs[0].Metadata["header_path"])
	assert.Equal(t, "/Real Header 1/", docs[1].Metadata["header_path"])
	assert.Contains(t, docs[0].PageContent, "```python")
	assert.Contains(t, docs[0].PageContent, "# This is NOT a header")
}

func TestMarkdownHeaderTextSplitter_HeaderPath(t *testing.T) {
	t.Parallel()

	markdown := `# Chapter 1
## Section 1.1
### Subsection 1.1.1

Content here.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 3, len(docs))
	assert.Equal(t, "/", docs[0].Metadata["header_path"])
	assert.Equal(t, "/Chapter 1/", docs[1].Metadata["header_path"])
	assert.Equal(t, "/Chapter 1/Section 1.1/", docs[2].Metadata["header_path"])
	assert.Equal(t, 1, len(docs[2].Metadata))
}

func TestMarkdownHeaderTextSplitter_NonSequentialHeaders(t *testing.T) {
	t.Parallel()

	markdown := `# Header 1
### Header 3

Content under H3.

## Header 2

Content under H2.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 3, len(docs))
	assert.Equal(t, "/Header 1/", docs[1].Metadata["header_path"])
	assert.Equal(t, "/Header 1/", docs[2].Metadata["header_path"])
}

func TestMarkdownHeaderTextSplitter_EmptySections(t *testing.T) {
	t.Parallel()

	markdown := `# Header 1
## Header 2
### Header 3

Only content under H3.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 3, len(docs))
	lastDoc := docs[len(docs)-1]
	assert.Equal(t, "/Header 1/Header 2/", lastDoc.Metadata["header_path"])
	assert.Contains(t, lastDoc.PageContent, "Only content under H3")
}

func TestMarkdownHeaderTextSplitter_MultipleCodeBlocks(t *testing.T) {
	t.Parallel()

	markdown := `# Documentation

## Installation

` + "```bash" + `
# Install the package
npm install package
` + "```" + `

## Usage

` + "```javascript" + `
// # This is a comment
const x = 1;
` + "```" + `

Done.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 3, len(docs))
	assert.Contains(t, docs[1].PageContent, "npm install")
	assert.Contains(t, docs[2].PageContent, "const x = 1")
}

func TestMarkdownHeaderTextSplitter_SplitDocuments(t *testing.T) {
	t.Parallel()

	originalDoc := schema.Document{
		PageContent: `# Header 1
Content here.

## Header 2
More content.`,
		Metadata: map[string]any{
			"source":       "test.md",
			"author":       "test",
			"custom_field": "value",
		},
		Score: 0.95,
	}

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitDocuments([]schema.Document{originalDoc})
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	for _, doc := range docs {
		assert.Equal(t, "test.md", doc.Metadata["source"])
		assert.Equal(t, "test", doc.Metadata["author"])
		assert.Equal(t, "value", doc.Metadata["custom_field"])
		assert.Equal(t, float32(0.95), doc.Score)
		assert.Contains(t, doc.Metadata, "header_path")
	}
}

func TestMarkdownHeaderTextSplitter_SplitTextCompatibility(t *testing.T) {
	t.Parallel()

	markdown := `# Header 1
Content 1.

## Header 2
Content 2.`

	splitter := NewMarkdownHeaderTextSplitter()
	chunks, err := splitter.SplitText(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(chunks))
	assert.Contains(t, chunks[0], "# Header 1")
	assert.Contains(t, chunks[1], "## Header 2")
}

func TestMarkdownHeaderTextSplitter_RealWorldExample(t *testing.T) {
	t.Parallel()

	markdown := `# Getting Started

Welcome to our documentation.

## Prerequisites

You need the following:

- Go 1.21+
- Git

` + "```bash" + `
# Check your Go version
go version
` + "```" + `

## Installation

### Using Go Install

Run the following command:

` + "```bash" + `
go install github.com/example/tool@latest
` + "```" + `

### Using Docker

Alternatively, use Docker:

` + "```dockerfile" + `
FROM golang:1.21
# Build the application
RUN go build
` + "```" + `

## Configuration

Edit the config file.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Greater(t, len(docs), 5)

	var goInstallDoc *schema.Document
	for i := range docs {
		if strings.Contains(docs[i].PageContent, "### Using Go Install") {
			goInstallDoc = &docs[i]
			break
		}
	}

	require.NotNil(t, goInstallDoc)
	assert.Equal(t, "/Getting Started/Installation/", goInstallDoc.Metadata["header_path"])
	assert.Contains(t, goInstallDoc.PageContent, "go install")
	assert.Equal(t, 1, len(goInstallDoc.Metadata))
}

func TestMarkdownHeaderTextSplitter_HeaderPathSeparator(t *testing.T) {
	t.Parallel()

	markdown := `# Chapter 1
## Section 1.1

Content.`

	splitter := &MarkdownHeaderTextSplitter{
		HeaderPathSeparator: " > ",
		IncludeMetadata:     true,
	}

	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	assert.Equal(t, " > Chapter 1 > ", docs[1].Metadata["header_path"])
}

func TestMarkdownHeaderTextSplitter_NoMetadata(t *testing.T) {
	t.Parallel()

	markdown := `# Header 1
Content.`

	splitter := &MarkdownHeaderTextSplitter{
		HeaderPathSeparator: "/",
		IncludeMetadata:     false,
	}

	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 1, len(docs))
	assert.Empty(t, docs[0].Metadata)
}

func TestMarkdownHeaderTextSplitter_SeparatorCollision(t *testing.T) {
	t.Parallel()

	markdown := `# Client / Server Architecture
## API / REST Design

Content here.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	assert.Equal(t, "/Client - Server Architecture/", docs[1].Metadata["header_path"])
	assert.Contains(t, docs[1].PageContent, "## API / REST Design")
}

func TestMarkdownHeaderTextSplitter_TildeFence(t *testing.T) {
	t.Parallel()

	markdown := `# Documentation

Some text.

~~~python
# This is NOT a header
def foo():
    pass
~~~

## Next Section

More text.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	assert.Contains(t, docs[0].PageContent, "~~~python")
	assert.Contains(t, docs[0].PageContent, "# This is NOT a header")
	assert.Contains(t, docs[0].PageContent, "~~~")
}

func TestMarkdownHeaderTextSplitter_MixedCodeFences(t *testing.T) {
	t.Parallel()

	markdown := `# Mixed Fences

` + "```bash" + `
# Bash comment
echo "test"
` + "```" + `

~~~python
# Python comment
print("test")
~~~

## Done

Text.`

	splitter := NewMarkdownHeaderTextSplitter()
	docs, err := splitter.SplitTextToDocuments(markdown)
	require.NoError(t, err)

	assert.Equal(t, 2, len(docs))
	assert.Contains(t, docs[0].PageContent, "```bash")
	assert.Contains(t, docs[0].PageContent, "~~~python")
	assert.Contains(t, docs[0].PageContent, "# Bash comment")
	assert.Contains(t, docs[0].PageContent, "# Python comment")
}
