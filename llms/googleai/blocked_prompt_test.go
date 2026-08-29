package googleai

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
)

type blockedPrompt struct{}

func (blockedPrompt) RoundTrip(r *http.Request) (*http.Response, error) {
	const body = `{"candidates":[],"promptFeedback":{"blockReason":"SAFETY",` +
		`"blockReasonMessage":"the prompt was rejected by the safety filter"}}`
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Request:    r,
	}, nil
}

func TestABlockedPromptIsNamedOnTheWholeAnswerPathToo(t *testing.T) {
	t.Parallel()

	llm, err := New(context.Background(),
		WithAPIKey("unit-test-key"), WithRest(),
		WithDefaultModel("gemini-2.5-flash"),
		WithHTTPClient(&http.Client{Transport: blockedPrompt{}}))
	require.NoError(t, err)

	_, err = llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")})

	var apiErr *llms.Error
	require.ErrorAs(t, err, &apiErr, "a blocked prompt must not read as an empty response")
	assert.Equal(t, llms.ErrCodeContentFilter, apiErr.Code)
	assert.Contains(t, apiErr.Message, "SAFETY")
}
