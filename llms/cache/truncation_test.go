package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func TestPartialAnswerTravelsWithTheError(t *testing.T) {
	t.Parallel()

	partial := &llms.ContentResponse{Choices: []*llms.ContentChoice{{
		Content:    "half an answ",
		StopReason: "length",
		Truncated:  true,
	}}}
	llm := newMockLLM(partial, llms.NewError(llms.ErrCodeTruncated, "", "stopped at the output token limit"))
	backend := newMockCache()

	resp, err := New(llm, backend).GenerateContent(t.Context(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	require.Error(t, err)
	require.NotNil(t, resp, "the partial answer must reach the caller alongside the error")
	require.Equal(t, "half an answ", resp.Choices[0].Content)
	require.True(t, resp.Choices[0].Truncated)
	require.Zero(t, backend.puts, "a failed generation must not be cached")
}
