package ollama

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestSamplingReachesTheWireInTheDoorsOwnShape(t *testing.T) {
	t.Parallel()

	body := captureChatRequest(t,
		llms.WithTemperature(0.3),
		llms.WithTopP(0.7),
		llms.WithTopK(11),
		llms.WithMaxTokens(555),
		llms.WithSeed(7),
		llms.WithStopWords([]string{"STOP"}),
		llms.WithRepetitionPenalty(1.3),
		llms.WithFrequencyPenalty(0.4),
		llms.WithPresencePenalty(0.6),
	)

	options, ok := body["options"].(map[string]any)
	require.True(t, ok, "no options in the request; body=%v", body)

	for field, want := range map[string]float64{
		"temperature":       0.3,
		"top_p":             0.7,
		"top_k":             11,
		"num_predict":       555,
		"seed":              7,
		"repeat_penalty":    1.3,
		"frequency_penalty": 0.4,
		"presence_penalty":  0.6,
	} {
		got, present := options[field]
		require.Truef(t, present, "%s never reached the wire; options=%v", field, options)
		require.InDeltaf(t, want, got, 1e-6, "%s = %v, want %v", field, got, want)
	}

	require.Equal(t, []any{"STOP"}, options["stop"], "stop words did not reach the wire")
}
