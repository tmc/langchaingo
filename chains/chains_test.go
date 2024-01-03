package chains

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

type testLanguageModel struct {
	// expected result of the language model
	expResult string
	// simulate work by sleeping for this duration
	simulateWork time.Duration
	// record the prompt that was passed to the language model
	recordedPrompt []schema.PromptValue
}

type stringPromptValue struct {
	s string
}

func (spv stringPromptValue) String() string {
	return spv.s
}

func (spv stringPromptValue) Messages() []schema.ChatMessage {
	return nil
}

func (l *testLanguageModel) Call(_ context.Context, prompt string, _ ...llms.CallOption) (string, error) {
	l.recordedPrompt = []schema.PromptValue{
		stringPromptValue{s: prompt},
	}
	if l.simulateWork > 0 {
		time.Sleep(l.simulateWork)
	}

	var llmResult string

	if l.expResult != "" {
		llmResult = l.expResult
	} else {
		llmResult = prompt
	}

	return llmResult, nil
}

func (l *testLanguageModel) Generate(
	ctx context.Context, prompts []string, options ...llms.CallOption,
) ([]*llms.Generation, error) {
	result, err := l.Call(ctx, prompts[0], options...)
	if err != nil {
		return nil, err
	}
	return []*llms.Generation{
		{
			Text: result,
		},
	}, nil
}

var _ llms.LLM = &testLanguageModel{}

func TestApply(t *testing.T) {
	t.Parallel()

	numInputs := 10
	maxWorkers := 5
	inputs := make([]map[string]any, numInputs)
	for i := 0; i < len(inputs); i++ {
		inputs[i] = map[string]any{
			"text": strconv.Itoa(i),
		}
	}

	c := NewLLMChain(&testLanguageModel{}, prompts.NewPromptTemplate("{{.text}}", []string{"text"}))
	results, err := Apply(context.Background(), c, inputs, maxWorkers)
	require.NoError(t, err)
	require.Equal(t, inputs, results, "inputs and results not equal")
}

func TestApplyWithCanceledContext(t *testing.T) {
	t.Parallel()

	numInputs := 10
	maxWorkers := 5
	inputs := make([]map[string]any, numInputs)
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	c := NewLLMChain(&testLanguageModel{simulateWork: time.Second}, prompts.NewPromptTemplate("test", nil))

	go func() {
		defer wg.Done()
		_, err := Apply(ctx, c, inputs, maxWorkers)
		require.Error(t, err)
	}()

	cancelFunc()
	wg.Wait()
}
