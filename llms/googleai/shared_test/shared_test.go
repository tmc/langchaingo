// nolint
package shared_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/embeddings"
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

type testFunc func(*testing.T, llms.Model)

// testFuncs is a list of all test functions in this file to run with both
// client types.
var testFuncs = []testFunc{
	testMultiContentText,
	testGenerateFromSinglePrompt,
	testMultiContentTextChatSequence,
	testMultiContentImageLink,
	testMultiContentImageBinary,
	testEmbeddings,
	testCandidateCountSetting,
	testMaxTokensSetting,
	testWithStreaming,
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

func testMultiContentTextUsingTextParts(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	content := llms.TextParts(
		schema.ChatMessageTypeHuman,
		"I'm a pomeranian",
		"What kind of mammal am I?",
	)

	rsp, err := llm.GenerateContent(context.Background(), []llms.MessageContent{content})
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)dog|canid|canine", c1.Content)
}

func testGenerateFromSinglePrompt(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	prompt := "name all the planets in the solar system"
	rsp, err := llms.GenerateFromSinglePrompt(context.Background(), llm, prompt)
	require.NoError(t, err)

	assert.Regexp(t, "(?i)jupiter", rsp)
}

func testMultiContentTextChatSequence(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name some countries")},
		},
		{
			Role:  schema.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Spain and Lesotho")},
		},
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Which if these is larger?")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-pro"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)spain.*larger", c1.Content)
}

func testMultiContentImageLink(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	parts := []llms.ContentPart{
		llms.ImageURLPart("https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"),
		llms.TextPart("describe this image in detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-pro-vision"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)parrot", c1.Content)
}

func testMultiContentImageBinary(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	b, err := os.ReadFile(filepath.Join("testdata", "parrot-icon.png"))
	if err != nil {
		t.Fatal(err)
	}

	parts := []llms.ContentPart{
		llms.BinaryPart("image/png", b),
		llms.TextPart("what does this image show? please use detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithModel("gemini-pro-vision"))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "(?i)parrot", c1.Content)
}

func testEmbeddings(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	texts := []string{"foo", "parrot", "foo"}
	emb := llm.(embeddings.EmbedderClient)
	res, err := emb.CreateEmbedding(context.Background(), texts)
	require.NoError(t, err)

	assert.Equal(t, len(texts), len(res))
	assert.NotEmpty(t, res[0])
	assert.NotEmpty(t, res[1])
	assert.Equal(t, res[0], res[2])
}

func testCandidateCountSetting(t *testing.T, llm llms.Model) {
	t.Helper()

	parts := []llms.ContentPart{
		llms.TextPart("Name five countries in Africa"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	{
		rsp, err := llm.GenerateContent(context.Background(), content,
			llms.WithCandidateCount(1), llms.WithTemperature(1))
		require.NoError(t, err)

		assert.Len(t, rsp.Choices, 1)
	}

	// TODO: test multiple candidates when the backend supports it
}

func testWithStreaming(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	content := llms.TextParts(
		schema.ChatMessageTypeHuman,
		"I'm a pomeranian",
		"Tell me more about my taxonomy",
	)

	var sb strings.Builder
	rsp, err := llm.GenerateContent(
		context.Background(),
		[]llms.MessageContent{content},
		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}))

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
	assert.Regexp(t, "dog|canid", strings.ToLower(sb.String()))
}

func testMaxTokensSetting(t *testing.T, llm llms.Model) {
	t.Helper()
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("Describe my taxonomy, health and care"),
	}
	content := []llms.MessageContent{
		{
			Role:  schema.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	// First, try this with a very low MaxTokens setting for such a query; expect
	// a stop reason that max of tokens was reached.
	{
		rsp, err := llm.GenerateContent(context.Background(), content,
			llms.WithMaxTokens(16))
		require.NoError(t, err)

		assert.NotEmpty(t, rsp.Choices)
		c1 := rsp.Choices[0]
		assert.Regexp(t, "(?i)MaxTokens", c1.StopReason)
	}

	// Now, try it again with a much larger MaxTokens setting and expect to
	// finish successfully and generate a response.
	{
		rsp, err := llm.GenerateContent(context.Background(), content,
			llms.WithMaxTokens(2048))
		require.NoError(t, err)

		assert.NotEmpty(t, rsp.Choices)
		c1 := rsp.Choices[0]
		assert.Regexp(t, "(?i)stop", c1.StopReason)
		assert.Regexp(t, "(?i)dog|breed|canid|canine", c1.Content)
	}
}
