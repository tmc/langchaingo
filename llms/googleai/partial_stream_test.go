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
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type deliverThenBreak struct{}

func (deliverThenBreak) RoundTrip(r *http.Request) (*http.Response, error) {
	const body = "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"sixty rooms are free\"}]," +
		"\"role\":\"model\"},\"index\":0}]}\r\n\r\n" +
		"data: {not json at all\r\n\r\n"
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Request:    r,
	}, nil
}

func TestAStreamThatBreaksMidAnswerKeepsWhatArrived(t *testing.T) {
	t.Parallel()

	llm, err := New(context.Background(),
		WithAPIKey("unit-test-key"), WithRest(),
		WithDefaultModel("gemini-2.5-flash"),
		WithHTTPClient(&http.Client{Transport: deliverThenBreak{}}))
	require.NoError(t, err)

	delivered := ""
	resp, err := llm.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "how many rooms are free?")},
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			if chunk.Type == streaming.ChunkTypeText {
				delivered += chunk.Content
			}
			return nil
		}))

	require.Error(t, err, "a broken stream must not look like a success")
	require.Equal(t, "sixty rooms are free", delivered, "the first chunk did reach the consumer")
	require.NotNil(t, resp, "what reached the consumer must travel with the error")
	require.NotEmpty(t, resp.Choices)
	assert.Equal(t, "sixty rooms are free", resp.Choices[0].Content)
}
