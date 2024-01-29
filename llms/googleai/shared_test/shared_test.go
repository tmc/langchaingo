package shared_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
	"github.com/tmc/langchaingo/schema"
)

func newGoogleAIClient(t *testing.T) *googleai.GoogleAI {
	t.Helper()

	genaiKey := os.Getenv("GENAI_API_KEY")
	if genaiKey == "" {
		t.Skip("GENAI_API_KEY not set")
		return nil
	}
	llm, err := googleai.New(context.Background(), googleai.WithAPIKey(genaiKey))
	require.NoError(t, err)
	return llm
}

func newVertexClient(t *testing.T) *vertex.Vertex {
	t.Helper()

	project := os.Getenv("VERTEX_PROJECT")
	if project == "" {
		t.Skip("VERTEX_PROJECT not set")
		return nil
	}
	location := os.Getenv("VERTEX_LOCATION")
	if location == "" {
		location = "us-central1"
	}

	llm, err := vertex.New(
		context.Background(),
		vertex.WithCloudProject(project),
		vertex.WithCloudLocation(location))
	require.NoError(t, err)
	return llm
}

// funcName obtains the name of the given function value, without a package
// prefix.
func funcName(f any) string {
	fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

func testMultiContentText(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)dog|canid|canine", c1.Content)
}

type testFunc func(*testing.T, llms.Model)

var testFuncs = []testFunc{
	testMultiContentText,
}

func TestShared(t *testing.T) {
	for _, f := range testFuncs {
		t.Run(fmt.Sprintf("%s-googleai", funcName(f)), func(t *testing.T) {
			llm := newGoogleAIClient(t)
			f(t, llm)
		})
		t.Run(fmt.Sprintf("%s-vertex", funcName(f)), func(t *testing.T) {
			llm := newVertexClient(t)
			f(t, llm)
		})
	}
}
