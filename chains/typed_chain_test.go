package chains

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms/fake"
	"github.com/tmc/langchaingo/outputparser"
	"github.com/tmc/langchaingo/prompts"
)

func TestTypedChainBasicUsage(t *testing.T) {
	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse("Hello, World!")

	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	chain := NewStringChain(llm, prompt)

	// Test typed call
	result, metadata, err := chain.Call(ctx, map[string]any{
		"input": "test input",
	})

	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", result)
	assert.NotNil(t, metadata)
	assert.Contains(t, metadata, "prompt")
	assert.Contains(t, metadata, "raw_output")
}

func TestTypedChainHelperFunctions(t *testing.T) {
	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse("Test Response")

	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	chain := NewStringChain(llm, prompt)

	// Test TypedRun
	result, err := TypedRun(ctx, chain, "test input")
	require.NoError(t, err)
	assert.Equal(t, "Test Response", result)

	// Test TypedPredict
	prediction, err := TypedPredict(ctx, chain, map[string]any{
		"input": "test input",
	})
	require.NoError(t, err)
	assert.Equal(t, "Test Response", prediction)
}

func TestTypedChainStructuredOutput(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse(`{"name": "test", "value": 42}`)

	parser := outputparser.NewStructured[TestStruct](
		"TestStruct",
		"Test structure",
		map[string]string{
			"name":  "test name",
			"value": "test value",
		},
	)

	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	chain := NewTypedLLMChain(llm, prompt, parser)

	result, metadata, err := chain.Call(ctx, map[string]any{
		"input": "extract test data",
	})

	require.NoError(t, err)
	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
	assert.NotNil(t, metadata)
}

func TestChainAdapter(t *testing.T) {
	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse("Adapted Response")

	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	typedChain := NewStringChain(llm, prompt)

	// Test adapter functionality
	adapter := NewChainAdapter(typedChain)
	
	// Use as legacy chain
	result, err := adapter.Call(ctx, map[string]any{
		"input": "test input",
	})

	require.NoError(t, err)
	assert.Contains(t, result, "result")
	assert.Equal(t, "Adapted Response", result["result"])

	// Test interface methods
	assert.Equal(t, []string{"input"}, adapter.GetInputKeys())
	assert.Equal(t, []string{"result"}, adapter.GetOutputKeys())
	assert.NotNil(t, adapter.GetMemory())
}

func TestTypedChainOutputType(t *testing.T) {
	llm := fake.New()
	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	
	stringChain := NewStringChain(llm, prompt)
	assert.Contains(t, stringChain.OutputType(), "string")

	type CustomType struct {
		Field string
	}
	
	parser := outputparser.NewStructured[CustomType](
		"CustomType", "desc", map[string]string{},
	)
	customChain := NewTypedLLMChain(llm, prompt, parser)
	assert.Contains(t, customChain.OutputType(), "CustomType")
}

func TestTypedChainInputValidation(t *testing.T) {
	ctx := context.Background()
	llm := fake.New()
	prompt := prompts.NewPromptTemplate("{{.required}}", []string{"required"})
	chain := NewStringChain(llm, prompt)

	// Test missing required input
	_, _, err := chain.Call(ctx, map[string]any{
		"wrong_key": "value",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required input key")

	// Test with correct input
	llm.SetResponse("Success")
	_, _, err = chain.Call(ctx, map[string]any{
		"required": "value",
	})
	assert.NoError(t, err)
}

// Benchmark to compare typed vs legacy chains
func BenchmarkTypedChainCall(b *testing.B) {
	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse("Benchmark Response")
	
	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	typedChain := NewStringChain(llm, prompt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := typedChain.Call(ctx, map[string]any{
			"input": "benchmark input",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLegacyChainCall(b *testing.B) {
	ctx := context.Background()
	llm := fake.New()
	llm.SetResponse("Benchmark Response")
	
	prompt := prompts.NewPromptTemplate("{{.input}}", []string{"input"})
	legacyChain := NewLLMChain(llm, prompt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := legacyChain.Call(ctx, map[string]any{
			"input": "benchmark input",
		})
		if err != nil {
			b.Fatal(err)
		}
		// Simulate type assertion that would be needed
		_, ok := result["text"].(string)
		if !ok {
			b.Fatal("type assertion failed")
		}
	}
}