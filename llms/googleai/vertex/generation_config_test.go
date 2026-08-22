package vertex

import (
	"math"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestGenerationConfigSaturatesInsteadOfWrapping(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		in         int
		wantTokens int32
	}{
		{"ordinary", 16384, 16384},
		{"exactly the ceiling", math.MaxInt32, math.MaxInt32},
		{"above the ceiling", math.MaxInt32 + 1, math.MaxInt32},
		{"far above the ceiling", 3_000_000_000, math.MaxInt32},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			n := tc.in
			model := &genai.GenerativeModel{}
			applyGenerationConfig(model, llms.CallOptions{
				MaxTokens:      &n,
				CandidateCount: &n,
				TopK:           &n,
			})

			if got := *model.MaxOutputTokens; got != tc.wantTokens {
				t.Errorf("MaxOutputTokens = %d, want %d", got, tc.wantTokens)
			}
			if got := *model.CandidateCount; got != tc.wantTokens {
				t.Errorf("CandidateCount = %d, want %d", got, tc.wantTokens)
			}
			if got := *model.TopK; got != tc.wantTokens {
				t.Errorf("TopK = %d, want %d", got, tc.wantTokens)
			}
		})
	}
}
