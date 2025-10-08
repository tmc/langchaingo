package mistral

import (
	"context"
	"math"
	"net/http"
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
		{
			name:     "negative values",
			input:    []float64{-1.5, -2.7, -3.9},
			expected: []float32{-1.5, -2.7, -3.9},
		},
		{
			name:     "large values",
			input:    []float64{1e6, 1e7, 1e8},
			expected: []float32{1e6, 1e7, 1e8},
		},
		{
			name:     "very small values",
			input:    []float64{1e-6, 1e-7, 1e-8},
			expected: []float32{1e-6, 1e-7, 1e-8},
		},
		{
			name:     "special values",
			input:    []float64{math.Inf(1), math.Inf(-1), 0},
			expected: []float32{float32(math.Inf(1)), float32(math.Inf(-1)), 0},
		},
		{
			name:     "nil input handling",
			input:    nil,
			expected: []float32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			output := convertFloat64ToFloat32(tt.input)

			require.Equal(t, len(tt.expected), len(output), "length mismatch")
			for i := range output {
				if math.IsInf(float64(tt.expected[i]), 0) {
					require.True(t, math.IsInf(float64(output[i]), int(math.Copysign(1, float64(tt.expected[i])))),
						"at index %d: expected %v, got %v", i, tt.expected[i], output[i])
				} else {
					require.Equal(t, tt.expected[i], output[i], "at index %d", i)
				}
			}
		})
	}
}

func TestErrEmptyEmbeddings(t *testing.T) {
	// Test the error constant
	require.NotNil(t, ErrEmptyEmbeddings)
	require.Equal(t, "empty embeddings", ErrEmptyEmbeddings.Error())
}

func TestMistralEmbed(t *testing.T) {
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "MISTRAL_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	ctx := context.Background()

	// Skip test during replay mode since Mistral SDK doesn't support HTTP client injection
	if !rr.Recording() {
		t.Skip("Mistral SDK doesn't support HTTP client injection - skipping replay test")
	}

	model, err := New()
	require.NoError(t, err)

	e, err := embeddings.NewEmbedder(model)
	require.NoError(t, err)

	_, err = e.EmbedDocuments(ctx, []string{"Hello world"})
	require.NoError(t, err)

	_, err = e.EmbedQuery(ctx, "Hello world")
	require.NoError(t, err)
}
