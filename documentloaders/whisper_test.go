package documentloaders

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranscription(t *testing.T) {
	t.Parallel()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
	t.Run("Test with local file", func(t *testing.T) {
		t.Parallel()
		audioFilePath := "./sample.mp3"
		_, err := os.Stat(audioFilePath)
		require.NoError(t, err)
		opts := []WhisperOpenAIOption{
			WithAudioPath(audioFilePath),
		}
		whisper := NewWhisperOpenAI(os.Getenv("OPENAI_API_KEY"), opts...)

		rsp, err := whisper.Load(context.Background())
		require.NoError(t, err)

		assert.NotEmpty(t, rsp)
	})

	t.Run("Test from url", func(t *testing.T) {
		t.Parallel()
		audioURL := "https://github.com/AssemblyAI-Examples/audio-examples/raw/main/20230607_me_canadian_wildfires.mp3"

		opts := []WhisperOpenAIOption{
			WithAudioPath(audioURL),
		}
		whisper := NewWhisperOpenAI(os.Getenv("OPENAI_API_KEY"), opts...)

		rsp, err := whisper.Load(context.Background())
		require.NoError(t, err)

		assert.NotEmpty(t, rsp)
	})
}
