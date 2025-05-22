package mistral

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/internal/httprr"
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

func newTestClient(t *testing.T, opts ...Option) *Mistral {
	t.Helper()
	
	// Check if we need an API key (only for recording mode)
	if httprr.GetTestMode() == httprr.TestModeRecord {
		if mistralKey := os.Getenv("MISTRAL_API_KEY"); mistralKey == "" {
			t.Skip("MISTRAL_API_KEY not set")
			return nil
		}
	} else {
		// For replay mode, provide a fake API key if none is set
		if os.Getenv("MISTRAL_API_KEY") == "" {
			opts = append([]Option{WithAPIKey("fake-api-key-for-testing")}, opts...)
		}
	}
	
	// Create HTTP client with httprr support
	httpClient := httprr.TestClient(t, "mistral_"+t.Name())
	opts = append([]Option{WithHTTPClient(httpClient)}, opts...)

	model, err := New(opts...)
	require.NoError(t, err)
	return model
}

func TestMistralEmbed(t *testing.T) {
	t.Parallel()
	model := newTestClient(t)

	e, err := embeddings.NewEmbedder(model)
	require.NoError(t, err)

	_, err = e.EmbedDocuments(context.Background(), []string{"Hello world"})
	require.NoError(t, err)

	_, err = e.EmbedQuery(context.Background(), "Hello world")
	require.NoError(t, err)
}
