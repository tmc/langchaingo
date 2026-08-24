package vertex

import (
	"math"
	"testing"

	"cloud.google.com/go/vertexai/genai"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
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

func TestResolveCallOptionsFillsTheTemperatureTheCallerLeftUnset(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		model string
		opts  []llms.CallOption
		want  float64
	}{
		{"a thinking-level model runs at 1.0", "gemini-3-pro-preview", nil, 1.0},
		{"an explicit value survives", "gemini-3-pro-preview", []llms.CallOption{llms.WithTemperature(0.2)}, 0.2},
		{"other families keep the client default", "gemini-2.5-flash", nil, 0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := &Vertex{opts: googleai.Options{DefaultModel: tc.model, DefaultTemperature: 0.5}}

			got := g.resolveCallOptions(tc.opts)
			if got.Temperature == nil {
				t.Fatal("temperature must be resolved, not left to the model")
			}
			if *got.Temperature != tc.want {
				t.Errorf("temperature = %v, want %v", *got.Temperature, tc.want)
			}
		})
	}
}

func TestRefusesReasoningOffOnlyWhereOmittingWouldNotDisable(t *testing.T) {
	t.Parallel()

	off := llms.WithReasoningDisabled()
	for _, tc := range []struct {
		model string
		want  bool
	}{
		{"gemini-2.0-flash", false},
		{"gemini-1.5-flash", false},
		{"gemini-2.5-flash-lite", false},
		{"gemini-2.5-flash", true},
		{"gemini-2.5-pro", true},
		{"", true},
	} {
		opts := llms.CallOptions{Model: &tc.model}
		off(&opts)
		if got := refusesReasoningOff(opts); got != tc.want {
			t.Errorf("refusesReasoningOff(%q) = %v, want %v", tc.model, got, tc.want)
		}
	}

	none := llms.CallOptions{}
	if refusesReasoningOff(none) {
		t.Error("a call that never asked to disable must not be refused")
	}
}
