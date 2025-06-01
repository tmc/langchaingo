package documentloaders

import (
	"os"
	"testing"

	aai "github.com/AssemblyAI/assemblyai-go-sdk"
	"github.com/stretchr/testify/require"
)

func TestAssemblyAIAudioTranscriptLoader_Load(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	var apiKey string
	if apiKey = os.Getenv("ASSEMBLYAI_API_KEY"); apiKey == "" {
		t.Skip("ASSEMBLYAI_API_KEY not set")
	}

	audioURL := "https://github.com/AssemblyAI-Examples/audio-examples/raw/main/20230607_me_canadian_wildfires.mp3"

	loader := NewAssemblyAIAudioTranscript(
		apiKey,
		WithAudioURL(audioURL),
		WithTranscriptFormat(TranscriptFormatText),
		WithTranscriptParams(&aai.TranscriptOptionalParams{
			RedactPII:         aai.Bool(true),
			RedactPIIPolicies: []aai.PIIPolicy{"person_name"},
		}),
	)

	docs, err := loader.Load(ctx)
	require.NoError(t, err)

	require.Len(t, docs, 1)

	require.NotEmpty(t, docs[0].PageContent)

	redactPII, ok := docs[0].Metadata["redact_pii"].(bool)

	require.True(t, ok)
	require.True(t, redactPII)
}

func TestAssemblyAIAudioTranscriptLoader_toMetadata(t *testing.T) {
	t.Parallel()

	metadata, err := toMetadata(aai.TranscriptSentence{
		Speaker: aai.String("1"),
		Text:    aai.String("This is a test sentence."),
	})
	require.NoError(t, err)

	require.Equal(t, "1", metadata["speaker"])
	require.Nil(t, metadata["text"])
}
