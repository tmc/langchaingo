package mistral_test

import (
	"context"
	"os"
	"testing"

	"github.com/tmc/langchaingo/embeddings"
	sdk "github.com/tmc/langchaingo/llms/mistral"
)

// TestConvertFloat64ToFloat32 tests the ConvertFloat64ToFloat32 function using table-driven tests.
func TestConvertFloat64ToFloat32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []float64
		expected []float32
	}{
		{
			name:     "empty slice",
			input:    []float64{},
			expected: []float32{},
		},
		{
			name:     "single element",
			input:    []float64{3.14},
			expected: []float32{3.14},
		},
		{
			name:     "multiple elements",
			input:    []float64{1.23, 4.56, 7.89},
			expected: []float32{1.23, 4.56, 7.89},
		},
		{
			name:     "zero values",
			input:    []float64{0.0, 0.0, 0.0},
			expected: []float32{0.0, 0.0, 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := sdk.ConvertFloat64ToFloat32(tt.input)
			if len(output) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(output), len(tt.expected))
				return
			}
			for i := range output {
				if output[i] != tt.expected[i] {
					t.Errorf("at index %d: got %f, want %f", i, output[i], tt.expected[i])
				}
			}
		})
	}
}

func TestMistralEmbed(t *testing.T) {
	t.Parallel()
	envVar := "MISTRAL_API_KEY"

	// Get the value of the environment variable
	value := os.Getenv(envVar)

	// Check if it is set (non-empty)
	if value == "" {
		t.Logf("Environment variable %s is not set, so skipping the test", envVar)
		return
	}

	model, err := sdk.New()
	if err != nil {
		panic(err)
	}

	e, err := embeddings.NewEmbedder(model)
	if err != nil {
		panic(err)
	}

	docEmbeddings, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	if err != nil {
		panic(err)
	}
	t.Logf("Document embeddings: %v\n", docEmbeddings)

	queryEmbedding, err := e.EmbedQuery(context.Background(), "Hello world")
	if err != nil {
		panic(err)
	}
	t.Logf("Query embedding: %v\n", queryEmbedding)
}
