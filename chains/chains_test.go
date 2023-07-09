package chains

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// testLanguageModel is a struct that implement the language model interface
// and returns the prompt value as a string.
type testLanguageModel struct{}

func (l testLanguageModel) GeneratePrompt(_ context.Context, promptValue []schema.PromptValue, _ ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.LLMResult{
		Generations: [][]*llms.Generation{{&llms.Generation{
			Text: promptValue[0].String(),
		}}},
	}, nil
}

func (l testLanguageModel) GetNumTokens(text string) int {
	return len(text)
}

var _ llms.LanguageModel = testLanguageModel{}

func TestApply(t *testing.T) {
	t.Parallel()

	numInputs := 10
	maxWorkers := 5
	inputs := make([]map[string]any, numInputs)
	for i := 0; i < len(inputs); i++ {
		inputs[i] = map[string]any{
			"text": fmt.Sprint(i),
		}
	}

	c := NewLLMChain(testLanguageModel{}, prompts.NewPromptTemplate("{{.text}}", []string{"text"}))
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
	c := NewLLMChain(testLanguageModel{}, prompts.NewPromptTemplate("test", nil))

	go func() {
		defer wg.Done()
		_, err := Apply(ctx, c, inputs, maxWorkers)
		require.Error(t, err)
	}()

	cancelFunc()
	wg.Wait()
}
