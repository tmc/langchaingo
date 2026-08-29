package anthropicclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func legacyStream(t *testing.T, body string) (*CompletionResponsePayload, error) {
	t.Helper()
	resp := &http.Response{Body: io.NopCloser(strings.NewReader(body))}
	return parseStreamingCompletionResponse(context.Background(), resp, &completionPayload{})
}

func TestALegacyLineLongerThanTheDefaultBufferSurvives(t *testing.T) {
	t.Parallel()

	answer := strings.Repeat("x", 200*1024)
	got, err := legacyStream(t, `data: {"completion":"`+answer+`","model":"claude-2"}`+"\n")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, answer, got.Completion, "a line over the 64KiB default must not be truncated")
}

func TestALegacyStreamWithNoEventsIsAnError(t *testing.T) {
	t.Parallel()

	got, err := legacyStream(t, ": keep-alive\n\n")

	require.ErrorIs(t, err, ErrEmptyResponse)
	assert.Nil(t, got)
}

func TestALegacyLineOverTheCeilingReportsTheFailure(t *testing.T) {
	t.Parallel()

	got, err := legacyStream(t, `data: {"completion":"`+strings.Repeat("x", maxStreamLine+1)+`"}`+"\n")

	require.ErrorContains(t, err, "failed to read stream",
		"the read failure must reach the caller by name, not as an empty-response error")
	require.NotErrorIs(t, err, ErrEmptyResponse)
	assert.Nil(t, got)
}
