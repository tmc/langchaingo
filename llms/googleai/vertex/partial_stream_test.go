package vertex

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/googleai"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type deliverThenBreak struct{}

func (deliverThenBreak) RoundTrip(r *http.Request) (*http.Response, error) {
	const body = `[{"candidates":[{"content":{"parts":[{"text":"sixty rooms are free"}],` +
		`"role":"model"},"index":0}]},` + `{not json at all`
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Request:    r,
	}, nil
}

func TestAStreamThatBreaksMidAnswerKeepsWhatArrived(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "unit-test-key")

	v, err := New(context.Background(),
		googleai.WithRest(), googleai.WithAPIKey("unit-test-key"),
		googleai.WithCloudProject("unit"), googleai.WithCloudLocation("us-central1"),
		googleai.WithDefaultModel("gemini-2.5-flash"),
		googleai.WithHTTPClient(&http.Client{Transport: deliverThenBreak{}}))
	require.NoError(t, err)
	t.Cleanup(func() { _ = v.Close() })

	delivered := ""
	resp, err := v.GenerateContent(context.Background(),
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
