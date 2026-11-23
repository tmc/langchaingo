package textsplitter_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/textsplitter"
)

func TestMarkdownNestedCodeBlocks(t *testing.T) {
	t.Parallel()

	for _, code := range []string{
		"```go\nfmt.Println(42)\n```",
		"    fmt.Println(42)",
	} {
		for depth := range 3 {
			for _, enabled := range []bool{false, true} {
				t.Run(fmt.Sprintf("%q/depth=%d/enabled=%t", code, depth, enabled), func(t *testing.T) {
					t.Parallel()

					prefix := strings.Repeat("> ", depth)
					input := prefix + strings.ReplaceAll(code, "\n", "\n"+prefix)
					splitter := textsplitter.NewMarkdownTextSplitter(textsplitter.WithCodeBlocks(enabled))
					docs, err := textsplitter.CreateDocuments(
						splitter, []string{input}, []map[string]any{{"source": "guide.md"}},
					)
					require.NoError(t, err)

					var content strings.Builder
					for _, doc := range docs {
						content.WriteString(doc.PageContent)
						assert.Equal(t, "guide.md", doc.Metadata["source"])
					}
					if enabled {
						assert.Contains(t, content.String(), "fmt.Println(42)")
					} else {
						assert.NotContains(t, content.String(), "fmt.Println(42)")
					}
				})
			}
		}
	}
}

func TestMarkdownNestedReferenceLinks(t *testing.T) {
	t.Parallel()

	for _, prefix := range []string{"", "> ", "> > ", "- ", "1. "} {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("%q/enabled=%t", prefix, enabled), func(t *testing.T) {
				t.Parallel()

				input := prefix + "_keep_ [API][ref]\n\n[ref]: https://example.com/api"
				splitter := textsplitter.NewMarkdownTextSplitter(textsplitter.WithReferenceLinks(enabled))
				chunks, err := splitter.SplitText(input)
				require.NoError(t, err)

				content := strings.Join(chunks, "\n")
				if enabled {
					assert.Contains(t, content, "*keep* [API](https://example.com/api)")
				} else {
					assert.Contains(t, content, "_keep_ [API][ref]")
					assert.NotContains(t, content, "https://example.com/api")
				}
			})
		}
	}
}

func TestMarkdownNestedJoinTableRows(t *testing.T) {
	t.Parallel()

	const table = "| A | B |\n|---|---|\n|x|1|\n|y|2|"
	for depth := range 3 {
		for _, enabled := range []bool{false, true} {
			t.Run(fmt.Sprintf("depth=%d/enabled=%t", depth, enabled), func(t *testing.T) {
				t.Parallel()

				prefix := strings.Repeat("> ", depth)
				input := prefix + strings.ReplaceAll(table, "\n", "\n"+prefix)
				splitter := textsplitter.NewMarkdownTextSplitter(textsplitter.WithJoinTableRows(enabled))
				chunks, err := splitter.SplitText(input)
				require.NoError(t, err)

				content := strings.Join(chunks, "\n")
				assert.Contains(t, content, "| x | 1 |")
				assert.Contains(t, content, "| y | 2 |")
				expectedHeaders := 2
				if enabled {
					expectedHeaders = 1
				}
				assert.Equal(t, expectedHeaders, strings.Count(content, "| A | B |"))
			})
		}
	}
}
