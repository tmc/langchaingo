package mistral

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/embeddings"
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
			output := convertFloat64ToFloat32(tt.input)

			require.Equal(t, len(tt.expected), len(output), "length mismatch")
			for i := range output {
				require.Equal(t, tt.expected[i], output[i], "at index %d", i)
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
		t.Skipf("Environment variable %s is not set, so skipping the test", envVar)
		return
	}

	model, err := New()
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(model)
	require.NoError(t, err)

	_, err = e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world")
	require.NoError(t, err)
}
