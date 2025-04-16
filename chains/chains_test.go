package chains

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
)

type testLanguageModel struct {
	// expected result of the language model
	expResult string
	// simulate work by sleeping for this duration
	simulateWork time.Duration
	// record the prompt that was passed to the language model
	recordedPrompt []llms.PromptValue
	mu             sync.Mutex
}

type stringPromptValue struct {
	s string
}

func (spv stringPromptValue) String() string {
	return spv.s
}

func (spv stringPromptValue) Messages() []llms.ChatMessage {
	return nil
}

func (l *testLanguageModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, l, prompt, options...)
}

func (l *testLanguageModel) GenerateContent(_ context.Context, mc []llms.MessageContent, _ ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace
	part0 := mc[0].Parts[0]
	var prompt string
	if tc, ok := part0.(llms.TextContent); ok {
		prompt = tc.Text
	} else {
		return nil, fmt.Errorf("passed non-text part")
	}
	l.mu.Lock()
	l.recordedPrompt = []llms.PromptValue{
		stringPromptValue{s: prompt},
	}
	l.mu.Unlock()

	if l.simulateWork > 0 {
		time.Sleep(l.simulateWork)
	}

	var llmResult string

	if l.expResult != "" {
		llmResult = l.expResult
	} else {
		llmResult = prompt
	}

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: llmResult},
		},
	}, nil
}

var _ llms.Model = &testLanguageModel{}

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

	var applyErr error
	go func() {
		defer wg.Done()
		_, applyErr = Apply(ctx, c, inputs, maxWorkers)
	}()

	cancelFunc()
	wg.Wait()

	if applyErr == nil || applyErr.Error() != "context canceled" {
		t.Fatal("expected context canceled error, got:", applyErr)
	}
}
