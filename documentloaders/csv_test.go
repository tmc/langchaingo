package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVLoader(t *testing.T) {
	t.Parallel()
	file, err := os.Open("./testdata/test.csv")
	assert.NoError(t, err)

	loader := NewCSV(file)

	docs, err := loader.Load(context.Background())
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
	assert.NoError(t, err)

	loader := NewCSV(file, "city")

	docs, err := loader.Load(context.Background())
	require.NoError(t, err)
	require.Len(t, docs, 20)

	expected1 := "city: New York"
	assert.Equal(t, docs[0].PageContent, expected1)

	expected2 := "city: London"
	assert.Equal(t, docs[1].PageContent, expected2)
}
