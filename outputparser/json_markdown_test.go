package outputparser_test

import (
	"strings"
	"testing"

	"github.com/tmc/langchaingo/outputparser"
)

func TestJSONMarkdownParser(t *testing.T) {
	t.Parallel()

	parser := outputparser.NewJSONMarkdown()

	input := strings.ReplaceAll(`User Query:
Show me all science fiction movies by Christopher Nolan that have a rating of more than 8, released after 2010 and limit them to two results

Structured Request:
"""json
{
    "query": "science fiction",
    "filter": "and(eq('director', 'Christopher Nolan'), gt('rating', 8), gt('year', 2010))",
    "limit": 2
}
"""
`, `"""`, "```")

	_, err := parser.Parse(input)
	if err != nil {
		t.Errorf("error parsing JSON %v", err)
	}
}
