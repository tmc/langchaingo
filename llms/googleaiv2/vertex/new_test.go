package vertex

import (
	"net/http"
	"testing"

	"github.com/vendasta/langchaingo/llms/googleaiv2"
	"google.golang.org/api/option"
)

func TestNewWithOptions(t *testing.T) {
	// Since we can't control the genai.NewClient behavior in tests,
	// we'll skip these tests that require actual API interaction
	t.Skip("Skipping New() tests that require actual API credentials or mocked clients")
}

func TestDefaultOptions(t *testing.T) {
	opts := googleaiv2.DefaultOptions()

	// Test default values
	if opts.DefaultModel != "gemini-2.0-flash" {
		t.Errorf("expected default model 'gemini-2.0-flash', got %q", opts.DefaultModel)
	}
	if opts.DefaultEmbeddingModel != "embedding-001" {
		t.Errorf("expected default embedding model 'embedding-001', got %q", opts.DefaultEmbeddingModel)
	}
	if opts.DefaultCandidateCount != 1 {
		t.Errorf("expected default candidate count 1, got %d", opts.DefaultCandidateCount)
	}
	if opts.DefaultMaxTokens != 2048 {
		t.Errorf("expected default max tokens 2048, got %d", opts.DefaultMaxTokens)
	}
	if opts.DefaultTemperature != 0.5 {
		t.Errorf("expected default temperature 0.5, got %f", opts.DefaultTemperature)
	}
	if opts.DefaultTopK != 3 {
		t.Errorf("expected default TopK 3, got %d", opts.DefaultTopK)
	}
	if opts.DefaultTopP != 0.95 {
		t.Errorf("expected default TopP 0.95, got %f", opts.DefaultTopP)
	}
	if opts.HarmThreshold != googleaiv2.HarmBlockOnlyHigh {
		t.Errorf("expected default harm threshold HarmBlockOnlyHigh, got %v", opts.HarmThreshold)
	}
}

func TestOptionsApplication(t *testing.T) { //nolint:funlen // comprehensive test //nolint:funlen // comprehensive test
	// Test that options modify the default correctly
	defaultOpts := googleaiv2.DefaultOptions()

	// Apply options
	googleaiv2.WithDefaultModel("custom-model")(&defaultOpts)
	if defaultOpts.DefaultModel != "custom-model" {
		t.Errorf("WithDefaultModel did not update model")
	}

	googleaiv2.WithDefaultEmbeddingModel("custom-embedding")(&defaultOpts)
	if defaultOpts.DefaultEmbeddingModel != "custom-embedding" {
		t.Errorf("WithDefaultEmbeddingModel did not update embedding model")
	}

	googleaiv2.WithDefaultCandidateCount(3)(&defaultOpts)
	if defaultOpts.DefaultCandidateCount != 3 {
		t.Errorf("WithDefaultCandidateCount did not update candidate count")
	}

	googleaiv2.WithDefaultMaxTokens(4096)(&defaultOpts)
	if defaultOpts.DefaultMaxTokens != 4096 {
		t.Errorf("WithDefaultMaxTokens did not update max tokens")
	}

	googleaiv2.WithDefaultTemperature(0.9)(&defaultOpts)
	if defaultOpts.DefaultTemperature != 0.9 {
		t.Errorf("WithDefaultTemperature did not update temperature")
	}

	googleaiv2.WithDefaultTopK(5)(&defaultOpts)
	if defaultOpts.DefaultTopK != 5 {
		t.Errorf("WithDefaultTopK did not update TopK")
	}

	googleaiv2.WithDefaultTopP(0.99)(&defaultOpts)
	if defaultOpts.DefaultTopP != 0.99 {
		t.Errorf("WithDefaultTopP did not update TopP")
	}

	googleaiv2.WithHarmThreshold(googleaiv2.HarmBlockNone)(&defaultOpts)
	if defaultOpts.HarmThreshold != googleaiv2.HarmBlockNone {
		t.Errorf("WithHarmThreshold did not update harm threshold")
	}

	googleaiv2.WithCloudProject("my-project")(&defaultOpts)
	if defaultOpts.CloudProject != "my-project" {
		t.Errorf("WithCloudProject did not update project")
	}

	googleaiv2.WithCloudLocation("europe-west1")(&defaultOpts)
	if defaultOpts.CloudLocation != "europe-west1" {
		t.Errorf("WithCloudLocation did not update location")
	}

	// Test client options
	googleaiv2.WithAPIKey("test-key")(&defaultOpts)
	if len(defaultOpts.ClientOptions) == 0 {
		t.Error("WithAPIKey did not add client option")
	}

	googleaiv2.WithCredentialsFile("creds.json")(&defaultOpts)
	found := false
	for _, opt := range defaultOpts.ClientOptions {
		if opt != nil {
			found = true
			break
		}
	}
	if !found {
		t.Error("WithCredentialsFile did not add client option")
	}

	googleaiv2.WithCredentialsJSON([]byte("{}"))(&defaultOpts)
	if len(defaultOpts.ClientOptions) < 2 {
		t.Error("WithCredentialsJSON did not add client option")
	}

	// Test empty credential options
	emptyOpts := googleaiv2.DefaultOptions()
	googleaiv2.WithCredentialsFile("")(&emptyOpts)
	googleaiv2.WithCredentialsJSON([]byte{})(&emptyOpts)
	if len(emptyOpts.ClientOptions) != 0 {
		t.Error("Empty credential options should not add client options")
	}
}

func TestHarmThresholdValues(t *testing.T) {
	// Test that harm threshold constants have expected values
	tests := []struct {
		name      string
		threshold googleaiv2.HarmBlockThreshold
		expected  int32
	}{
		{"HarmBlockUnspecified", googleaiv2.HarmBlockUnspecified, 0},
		{"HarmBlockLowAndAbove", googleaiv2.HarmBlockLowAndAbove, 1},
		{"HarmBlockMediumAndAbove", googleaiv2.HarmBlockMediumAndAbove, 2},
		{"HarmBlockOnlyHigh", googleaiv2.HarmBlockOnlyHigh, 3},
		{"HarmBlockNone", googleaiv2.HarmBlockNone, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int32(tt.threshold) != tt.expected {
				t.Errorf("expected %s to be %d, got %d", tt.name, tt.expected, int32(tt.threshold))
			}
		})
	}
}

func TestVertexStructure(t *testing.T) {
	// Test that Vertex has the expected fields
	// This is mainly for documentation and to catch breaking changes
	v := &Vertex{}

	// Check that fields exist by assignment (will fail to compile if missing)
	v.CallbacksHandler = nil
	v.client = nil
	v.opts = googleaiv2.Options{}
	v.palmClient = nil
}

func TestWithHTTPClientOption(t *testing.T) {
	opts := googleaiv2.DefaultOptions()

	// Create a custom HTTP client
	httpClient := &http.Client{}

	// Apply the HTTP client option
	googleaiv2.WithHTTPClient(httpClient)(&opts)

	// We can't directly check the client options, but we can verify
	// that the option was added
	if len(opts.ClientOptions) == 0 {
		t.Error("WithHTTPClient did not add a client option")
	}
}

func TestOptionsEnsureAuthPresent(t *testing.T) {
	tests := []struct {
		name           string
		opts           googleaiv2.Options
		envAPIKey      string
		expectAddition bool
	}{
		{
			name: "with existing auth options",
			opts: googleaiv2.Options{
				ClientOptions: []option.ClientOption{
					option.WithAPIKey("existing-key"),
				},
			},
			envAPIKey:      "env-key",
			expectAddition: false,
		},
		{
			name:           "without auth options but with env key",
			opts:           googleaiv2.Options{},
			envAPIKey:      "env-key",
			expectAddition: true,
		},
		{
			name:           "without auth options and no env key",
			opts:           googleaiv2.Options{},
			envAPIKey:      "",
			expectAddition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var first
			t.Setenv("GOOGLE_API_KEY", "")

			// Set env var for test if needed
			if tt.envAPIKey != "" {
				t.Setenv("GOOGLE_API_KEY", tt.envAPIKey)
			}

			// Make a copy of opts to avoid modifying the original
			opts := tt.opts
			initialLen := len(opts.ClientOptions)
			opts.EnsureAuthPresent()

			if tt.expectAddition {
				if len(opts.ClientOptions) <= initialLen {
					t.Error("expected auth option to be added")
				}
			} else {
				if len(opts.ClientOptions) > initialLen {
					t.Error("expected no auth option to be added")
				}
			}
		})
	}
}
