package outputparser_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/outputparser"
)

func TestDefinedJSONFieldNames(t *testing.T) {
	t.Parallel()
	type Item struct {
		Title string `json:"title,omitempty"`
		Count int    `json:",omitempty"`
		Local string `json:"-" describe:"local application state"`
	}
	parser, err := outputparser.NewDefined(Item{})
	require.NoError(t, err)
	instructions := parser.GetFormatInstructions()
	require.Contains(t, instructions, "\ttitle: string;")
	require.Contains(t, instructions, "\tCount: int;")
	require.NotContains(t, instructions, "omitempty")
	require.NotContains(t, instructions, "local application state")

	result, err := parser.Parse("```json\n{\"title\":\"A book\",\"Count\":2}\n```")
	require.NoError(t, err)
	require.Equal(t, Item{Title: "A book", Count: 2}, result)
}

func TestDefinedNestedJSONFieldNames(t *testing.T) {
	t.Parallel()
	type Item struct {
		Name string `json:"name,omitempty"`
		Skip string `json:"-"`
	}
	parser, err := outputparser.NewDefined(struct {
		Items []Item `json:"items,omitempty"`
	}{})
	require.NoError(t, err)
	instructions := parser.GetFormatInstructions()
	require.Contains(t, instructions, "\titems: Items[];")
	require.Contains(t, instructions, "\tname: string;")
	require.NotContains(t, instructions, "omitempty")
	require.NotContains(t, instructions, "\t-:")
}
