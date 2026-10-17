package documentloaders

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVLoader(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.csv")
	require.NoError(t, err)

	loader := NewCSV(file)

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 20)

	expected1 := "name: John Doe\nage: 25\ncity: New York\ncountry: United States"
	assert.Equal(t, docs[0].PageContent, expected1)

	expected2 := `name: Jane Smith
age: 32
city: London
country: United Kingdom`
	assert.Equal(t, docs[1].PageContent, expected2)
}

func TestCSVLoaderWithFilteringColumns(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.csv")
	require.NoError(t, err)

	loader := NewCSV(file, "city")

	ctx := context.Background()
	docs, err := loader.Load(ctx)
	require.NoError(t, err)
	require.Len(t, docs, 20)

	expected1 := "city: New York"
	assert.Equal(t, docs[0].PageContent, expected1)

	expected2 := "city: London"
	assert.Equal(t, docs[1].PageContent, expected2)
}

func TestCSVLoaderUnevenRows(t *testing.T) {
	t.Parallel()
	// A short row and a long row used to fail the whole file with
	// "wrong number of fields".
	loader := NewCSV(strings.NewReader("name,age,city\nJohn,25\nJane,32,London,UK\n"))

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Equal(t, "name: John\nage: 25", docs[0].PageContent)
	assert.Equal(t, "name: Jane\nage: 32\ncity: London\nUK", docs[1].PageContent)
}

func TestCSVLoaderUnevenRowsWithFilteringColumns(t *testing.T) {
	t.Parallel()
	loader := NewCSV(strings.NewReader("name,age,city\nJohn,25\nJane,32,London,UK\n"), "city")

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Empty(t, docs[0].PageContent)
	assert.Equal(t, "city: London", docs[1].PageContent)
}

func TestCSVLoaderByteOrderMark(t *testing.T) {
	t.Parallel()
	loader := NewCSV(strings.NewReader("\ufeffname,age\nJohn,25\n"), "name")

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 1)
	assert.Equal(t, "name: John", docs[0].PageContent)
}
