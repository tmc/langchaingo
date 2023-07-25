package chains

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

var errDummy = errors.New("boom")

func TestSimpleSequential(t *testing.T) {
	t.Parallel()

	// Build and execute a simple sequential chain with two LLMChains
	testLLM1 := &testLanguageModel{expResult: "the chicken crossed the road"}
	testLLM2 := &testLanguageModel{expResult: "The chicken made it to the other side"}

	chains := []Chain{
		NewLLMChain(testLLM1, prompts.NewPromptTemplate("{{.input}}", []string{"input"})),
		NewLLMChain(testLLM2, prompts.NewPromptTemplate("What happened after {{.output}}?", []string{"output"})),
	}
	simpleSeqChain, err := NewSimpleSequentialChain(chains)
	require.NoError(t, err)

	res, err := Run(context.Background(), simpleSeqChain, "What did the chicken do?")
	require.NoError(t, err)

	// Assert that the second LLMChain received the output of the first LLMChain
	expPrompt := "What happened after the chicken crossed the road?"
	assert.Equal(t, expPrompt, testLLM2.recordedPrompt[0].String())

	// Assert that the output of the second LLMChain is the output of the entire chain
	assert.Equal(t, "The chicken made it to the other side", res)
}

func TestSimpleSequentialErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		chain   Chain
		initErr error
		execErr error
	}{
		{
			name:    "multiple inputs",
			chain:   &testLLMChain{inputKeys: []string{"input1", "input2"}},
			initErr: ErrInvalidInputNumberInSimpleSeq,
		},
		{
			name:    "multiple outputs",
			chain:   &testLLMChain{inputKeys: []string{"input"}, outputKeys: []string{"output1", "output2"}},
			initErr: ErrInvalidOutputNumberInSimpleSeq,
		},
		{
			name:    "chain execution error",
			chain:   &testLLMChain{err: errDummy, inputKeys: []string{"input"}, outputKeys: []string{"output"}},
			execErr: errDummy,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewSimpleSequentialChain([]Chain{tc.chain})
			if tc.initErr != nil {
				assert.ErrorIs(t, err, tc.initErr)
			} else {
				require.NoError(t, err)
				_, err := Run(context.Background(), c, "Do something")
				assert.ErrorIs(t, err, tc.execErr)
			}
		})
	}
}

func TestSequentialChain(t *testing.T) {
	t.Parallel()

	// Build and execute a sequential chain with three LLMChains
	testLLM1 := &testLanguageModel{expResult: "In the year 3000, chickens have taken over the world"}
	testLLM2 := &testLanguageModel{expResult: "An egg-citing adventure"}
	testLLM3 := &testLanguageModel{expResult: "Vey legit"}

	chain1 := NewLLMChain(
		testLLM1,
		prompts.NewPromptTemplate("Write a story titled {{.title}} set in the year {{.year}}", []string{"title", "year"}),
	)
	chain1.OutputKey = "story"
	chain2 := NewLLMChain(testLLM2, prompts.NewPromptTemplate("Review this story: {{.story}}", []string{"story"}))
	chain2.OutputKey = "review"
	chain3 := NewLLMChain(
		testLLM3,
		prompts.NewPromptTemplate("Tell me if this review is legit: {{.review}}", []string{"review"}),
	)

	chains := []Chain{chain1, chain2, chain3}

	seqChain, err := NewSequentialChain(chains, []string{"title", "year"}, []string{_llmChainDefaultOutputKey})
	require.NoError(t, err)

	res, err := Call(context.Background(), seqChain, map[string]any{"title": "Chicken Takeover", "year": 3000})
	require.NoError(t, err)

	// Assert that the second LLMChain received the output of the first LLMChain
	expPrompt := "Review this story: In the year 3000, chickens have taken over the world"
	assert.Equal(t, expPrompt, testLLM2.recordedPrompt[0].String())

	// Assert that the third LLMChain received the output of the second LLMChain
	expPrompt = "Tell me if this review is legit: An egg-citing adventure"
	assert.Equal(t, expPrompt, testLLM3.recordedPrompt[0].String())

	// Assert that the output of the third LLMChain is the output of the entire chain
	assert.Equal(t, "Vey legit", res[_llmChainDefaultOutputKey])
}

func TestSequentialChainErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		chains       []Chain
		initErr      error
		execErr      error
		seqChainOpts []SequentialChainOption
	}{
		{
			name: "missing input key",
			chains: []Chain{
				&testLLMChain{inputKeys: []string{"input1", "input2"}, outputKeys: []string{"output"}},
				// 2nd chain's input key does not exist
				&testLLMChain{inputKeys: []string{"non-existent-input"}},
			},
			initErr: ErrChainInitialization,
		},
		{
			name: "overlapping output key",
			chains: []Chain{
				&testLLMChain{inputKeys: []string{"input1", "input2"}, outputKeys: []string{"output"}},
				// 2nd chain's output key overlaps with 1st chain's output key
				&testLLMChain{inputKeys: []string{"output"}, outputKeys: []string{"output"}},
			},
			initErr: ErrChainInitialization,
		},
		{
			name: "missing output key",
			// no chains have 'output' key which is expected by the sequential chain
			chains: []Chain{
				&testLLMChain{inputKeys: []string{"input1", "input2"}, outputKeys: []string{"output1"}},
				&testLLMChain{inputKeys: []string{"output1"}, outputKeys: []string{"output2"}},
			},
			initErr: ErrChainInitialization,
		},
		{
			name: "chain execution error",
			chains: []Chain{
				// chain throws an error
				&testLLMChain{inputKeys: []string{"input1", "input2"}, outputKeys: []string{"output"}, err: errDummy},
			},
			execErr: errDummy,
		},
		{
			name: "memory key collides with input key",
			chains: []Chain{
				&testLLMChain{inputKeys: []string{"input1"}, outputKeys: []string{"output"}},
			},
			initErr:      ErrChainInitialization,
			seqChainOpts: []SequentialChainOption{WithSeqChainMemory(memory.NewBuffer(memory.WithMemoryKey("input1")))},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c, err := NewSequentialChain(tc.chains, []string{"input1", "input2"}, []string{"output"}, tc.seqChainOpts...)
			if tc.initErr != nil {
				assert.ErrorIs(t, err, tc.initErr)
			} else {
				require.NoError(t, err)
				_, err := Call(context.Background(), c, map[string]any{"input1": "foo", "input2": "bar"})
				assert.ErrorIs(t, err, tc.execErr)
			}
		})
	}
}

// LLMChain for testing purposes.
type testLLMChain struct {
	err        error
	inputKeys  []string
	outputKeys []string
}

// Call runs the logic of the chain and returns the output. This method should
// not be called directly. Use rather the Call, Run or Predict functions that
// handles the memory and other aspects of the chain.
func (c *testLLMChain) Call(_ context.Context, _ map[string]any, _ ...ChainCallOption) (map[string]any, error) { //nolint:lll
	return nil, c.err
}

// GetMemory gets the memory of the chain.
func (c *testLLMChain) GetMemory() schema.Memory {
	return memory.NewSimple()
}

// InputKeys returns the input keys the chain expects.
func (c *testLLMChain) GetInputKeys() []string {
	return c.inputKeys
}

// OutputKeys returns the output keys the chain returns.
func (c *testLLMChain) GetOutputKeys() []string {
	return c.outputKeys
}
