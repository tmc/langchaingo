package documentLoaders

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTextLoader(t *testing.T) {
	loader := NewTextLoaderFromFile("./testdata/test.txt")

	docs, err := loader.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(docs) != 1 {
		t.Errorf("Number of docs from text loader expected to be 1")
	}

	expectedPageContent := "Foo Bar Baz"
	if docs[0].PageContent != expectedPageContent {
		t.Errorf("Page content form text loader not the same as expected. Got:\n %s\nExpect:\n%s", docs[0].PageContent, expectedPageContent)
	}

	expectedMetadata := map[string]any{
		"source": "./testdata/test.txt",
	}

	if !cmp.Equal(docs[0].Metadata, expectedMetadata) {
		t.Errorf("Meta data form text loader not the same as expected. Got:\n %s\nExpect:%s\n", docs[0].Metadata, expectedMetadata)
	}
}
