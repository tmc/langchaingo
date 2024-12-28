package mistral_test

import (
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"testing"

	"github.com/tmc/langchaingo/embeddings"
	sdk "github.com/tmc/langchaingo/llms/mistral"
)

func TestConvertFloat64ToFloat32(t *testing.T) {

	// Test case 1: Empty slice
	input := []float64{}
	output := sdk.ConvertFloat64ToFloat32(input)
	if len(output) != 0 {
		t.Errorf("Expected output length to be 0, but got %d", len(output))
	}

	// Test case 2: Single element slice
	input = []float64{3.14}
	output = sdk.ConvertFloat64ToFloat32(input)
	if len(output) != 1 || output[0] != float32(3.14) {
		t.Errorf("Expected output to be [3.14], but got %v", output)
	}

	// Test case 3: Multiple element slice
	input = []float64{1.23, 4.56, 7.89}
	output = sdk.ConvertFloat64ToFloat32(input)
	if len(output) != 3 || output[0] != float32(1.23) || output[1] != float32(4.56) || output[2] != float32(7.89) {
		t.Errorf("Expected output to be [1.23 4.56 7.89], but got %v", output)
	}

	// Test case 4: Large random slice
	input = make([]float64, 1_000_000)
	r, err := rand.Int(rand.Reader, big.NewInt(42))
	if err != nil {
		panic(err)
	}
	for i := range input {
		input[i] = float64(r.Int64())
	}
	output = sdk.ConvertFloat64ToFloat32(input)
	if len(output) != len(input) {
		t.Errorf("Expected output length to be %d, but got %d", len(input), len(output))
	}
	for i, v := range input {
		if float32(v) != output[i] {
			t.Errorf("Expected output[%d] to be %f, but got %f", i, float32(v), output[i])
		}
	}
}

func TestMistralEmbed(t *testing.T) {
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

	aaa, err := e.EmbedDocuments(context.Background(), []string{"Hello world"})
	if err != nil {
		panic(err)
	}
	t.Logf("aaa EQ: %v\n", aaa)

	bbb, err := e.EmbedQuery(context.Background(), "Hello world")
	if err != nil {
		panic(err)
	}

	t.Logf("bbb EQ: %v\n", bbb)
}
