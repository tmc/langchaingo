package chains

import (
	"context"
	"errors"
	"testing"

	"github.com/0xDezzy/langchaingo/callbacks"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Unit tests that don't require external dependencies

type mockChain struct {
	mock.Mock
	inputKeys  []string
	outputKeys []string
	memory     schema.Memory
}

func (m *mockChain) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
	args := m.Called(ctx, inputs, options)
	if outputs := args.Get(0); outputs != nil {
		return outputs.(map[string]any), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockChain) GetMemory() schema.Memory {
	if m.memory != nil {
		return m.memory
	}
	return &mockMemory{}
}

func (m *mockChain) GetInputKeys() []string {
	return m.inputKeys
}

func (m *mockChain) GetOutputKeys() []string {
	return m.outputKeys
}

type mockMemory struct {
	mock.Mock
}

func (m *mockMemory) MemoryVariables(ctx context.Context) []string {
	args := m.Called(ctx)
	return args.Get(0).([]string)
}

func (m *mockMemory) LoadMemoryVariables(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	args := m.Called(ctx, inputs)
	if vars := args.Get(0); vars != nil {
		return vars.(map[string]any), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockMemory) SaveContext(ctx context.Context, inputs, outputs map[string]any) error {
	args := m.Called(ctx, inputs, outputs)
	return args.Error(0)
}

func (m *mockMemory) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockMemory) GetMemoryKey(ctx context.Context) string {
	args := m.Called(ctx)
	return args.String(0)
}

func TestValidateInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputKeys   []string
		inputVals   map[string]any
		expectErr   bool
		errContains string
	}{
		{
			name:      "valid inputs",
			inputKeys: []string{"key1", "key2"},
			inputVals: map[string]any{"key1": "value1", "key2": "value2"},
			expectErr: false,
		},
		{
			name:      "empty inputs and keys",
			inputKeys: []string{},
			inputVals: map[string]any{},
			expectErr: false,
		},
		{
			name:        "missing input key",
			inputKeys:   []string{"key1", "key2"},
			inputVals:   map[string]any{"key1": "value1"},
			expectErr:   true,
			errContains: "missing key in input values",
		},
		{
			name:        "no input values",
			inputKeys:   []string{"key1"},
			inputVals:   map[string]any{},
			expectErr:   true,
			errContains: "missing key in input values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &mockChain{inputKeys: tt.inputKeys}
			err := validateInputs(chain, tt.inputVals)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		outputKeys  []string
		outputVals  map[string]any
		expectErr   bool
		errContains string
	}{
		{
			name:       "valid outputs",
			outputKeys: []string{"result1", "result2"},
			outputVals: map[string]any{"result1": "value1", "result2": "value2"},
			expectErr:  false,
		},
		{
			name:       "empty outputs and keys",
			outputKeys: []string{},
			outputVals: map[string]any{},
			expectErr:  false,
		},
		{
			name:        "missing output key",
			outputKeys:  []string{"result1", "result2"},
			outputVals:  map[string]any{"result1": "value1"},
			expectErr:   true,
			errContains: "missing key in output values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &mockChain{outputKeys: tt.outputKeys}
			err := validateOutputs(chain, tt.outputVals)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetChainCallbackHandler(t *testing.T) {
	t.Parallel()

	t.Run("chain without callback handler", func(t *testing.T) {
		chain := &mockChain{}
		handler := getChainCallbackHandler(chain)
		assert.Nil(t, handler)
	})

	t.Run("chain with callback handler", func(t *testing.T) {
		mockHandler := &callbacks.SimpleHandler{}

		// Mock the GetCallbackHandler method
		handlerHaver := &mockHandlerHaver{handler: mockHandler}

		handler := getChainCallbackHandler(handlerHaver)
		assert.Equal(t, mockHandler, handler)
	})
}

type mockHandlerHaver struct {
	handler callbacks.Handler
}

func (m *mockHandlerHaver) GetCallbackHandler() callbacks.Handler {
	return m.handler
}

func (m *mockHandlerHaver) Call(ctx context.Context, inputs map[string]any, options ...ChainCallOption) (map[string]any, error) {
	return nil, nil
}

func (m *mockHandlerHaver) GetMemory() schema.Memory {
	return &mockMemory{}
}

func (m *mockHandlerHaver) GetInputKeys() []string {
	return []string{}
}

func (m *mockHandlerHaver) GetOutputKeys() []string {
	return []string{}
}

func TestSendApplyInputJobs(t *testing.T) {
	t.Parallel()

	inputValues := []map[string]any{
		{"key1": "value1"},
		{"key2": "value2"},
		{"key3": "value3"},
	}

	inputJobs := make(chan applyInputJob, len(inputValues))

	sendApplyInputJobs(inputJobs, inputValues)

	// Collect all jobs from the channel
	var receivedJobs []applyInputJob
	for job := range inputJobs {
		receivedJobs = append(receivedJobs, job)
	}

	// Verify we received the expected number of jobs
	assert.Len(t, receivedJobs, len(inputValues))

	// Verify job contents
	for i, job := range receivedJobs {
		assert.Equal(t, inputValues[i], job.input)
		assert.Equal(t, i, job.i)
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "intermediateSteps", _intermediateStepsOutputKey)
	assert.Equal(t, 5, _defaultApplyMaxNumberWorkers)
}

func TestErrorConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrInvalidInputValues",
			err:         ErrInvalidInputValues,
			expectedMsg: "invalid input values",
		},
		{
			name:        "ErrMissingInputValues",
			err:         ErrMissingInputValues,
			expectedMsg: "missing key in input values",
		},
		{
			name:        "ErrInputValuesWrongType",
			err:         ErrInputValuesWrongType,
			expectedMsg: "input key is of wrong type",
		},
		{
			name:        "ErrMissingMemoryKeyValues",
			err:         ErrMissingMemoryKeyValues,
			expectedMsg: "missing memory key in input values",
		},
		{
			name:        "ErrMemoryValuesWrongType",
			err:         ErrMemoryValuesWrongType,
			expectedMsg: "memory key is of wrong type",
		},
		{
			name:        "ErrInvalidOutputValues",
			err:         ErrInvalidOutputValues,
			expectedMsg: "missing key in output values",
		},
		{
			name:        "ErrMultipleInputsInRun",
			err:         ErrMultipleInputsInRun,
			expectedMsg: "run not supported in chain with more then one expected input",
		},
		{
			name:        "ErrMultipleOutputsInRun",
			err:         ErrMultipleOutputsInRun,
			expectedMsg: "run not supported in chain with more then one expected output",
		},
		{
			name:        "ErrWrongOutputTypeInRun",
			err:         ErrWrongOutputTypeInRun,
			expectedMsg: "run not supported in chain that returns value that is not string",
		},
		{
			name:        "ErrOutputNotStringInPredict",
			err:         ErrOutputNotStringInPredict,
			expectedMsg: "predict is not supported with a chain that does not return a string",
		},
		{
			name:        "ErrMultipleOutputsInPredict",
			err:         ErrMultipleOutputsInPredict,
			expectedMsg: "predict is not supported with a chain that returns multiple values",
		},
		{
			name:        "ErrChainInitialization",
			err:         ErrChainInitialization,
			expectedMsg: "error initializing chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedMsg, tt.err.Error())
		})
	}
}

func TestChainInterface(t *testing.T) {
	t.Parallel()

	// Verify mockChain implements the Chain interface
	var _ Chain = &mockChain{}
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("multiple inputs error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input1", "input2"},
			outputKeys: []string{"output"},
		}

		memory := &mockMemory{}
		memory.On("MemoryVariables", ctx).Return([]string{})
		chain.memory = memory

		result, err := Run(ctx, chain, "test")
		assert.Error(t, err)
		assert.Equal(t, ErrMultipleInputsInRun, err)
		assert.Empty(t, result)
	})

	t.Run("multiple outputs error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"output1", "output2"},
		}

		memory := &mockMemory{}
		memory.On("MemoryVariables", ctx).Return([]string{})
		chain.memory = memory

		result, err := Run(ctx, chain, "test")
		assert.Error(t, err)
		assert.Equal(t, ErrMultipleOutputsInRun, err)
		assert.Empty(t, result)
	})

	t.Run("wrong output type error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"output"},
		}

		memory := &mockMemory{}
		memory.On("MemoryVariables", ctx).Return([]string{})
		memory.On("LoadMemoryVariables", ctx, mock.Anything).Return(map[string]any{}, nil)
		memory.On("SaveContext", ctx, mock.Anything, mock.Anything).Return(nil)
		chain.memory = memory

		// Chain returns non-string output
		chain.On("Call", ctx, mock.Anything, mock.Anything).Return(map[string]any{
			"output": 123, // non-string
		}, nil)

		result, err := Run(ctx, chain, "test")
		assert.Error(t, err)
		assert.Equal(t, ErrWrongOutputTypeInRun, err)
		assert.Empty(t, result)
	})
}

func TestPredictErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("multiple outputs error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"output1", "output2"},
		}

		memory := &mockMemory{}
		memory.On("MemoryVariables", ctx).Return([]string{})
		memory.On("LoadMemoryVariables", ctx, mock.Anything).Return(map[string]any{}, nil)
		memory.On("SaveContext", ctx, mock.Anything, mock.Anything).Return(nil)
		chain.memory = memory

		chain.On("Call", ctx, mock.Anything, mock.Anything).Return(map[string]any{
			"output1": "value1",
			"output2": "value2",
		}, nil)

		result, err := Predict(ctx, chain, map[string]any{"input": "test"})
		assert.Error(t, err)
		assert.Equal(t, ErrMultipleOutputsInPredict, err)
		assert.Empty(t, result)
	})

	t.Run("non-string output error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"output"},
		}

		memory := &mockMemory{}
		memory.On("MemoryVariables", ctx).Return([]string{})
		memory.On("LoadMemoryVariables", ctx, mock.Anything).Return(map[string]any{}, nil)
		memory.On("SaveContext", ctx, mock.Anything, mock.Anything).Return(nil)
		chain.memory = memory

		chain.On("Call", ctx, mock.Anything, mock.Anything).Return(map[string]any{
			"output": 123, // non-string
		}, nil)

		result, err := Predict(ctx, chain, map[string]any{"input": "test"})
		assert.Error(t, err)
		assert.Equal(t, ErrOutputNotStringInPredict, err)
		assert.Empty(t, result)
	})
}

func TestCallChainWithValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("input validation error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"required_input"},
			outputKeys: []string{"output"},
		}

		// Missing required input
		inputs := map[string]any{"wrong_key": "value"}

		result, err := callChain(ctx, chain, inputs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing key in input values")
		assert.Nil(t, result)
	})

	t.Run("output validation error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"required_output"},
		}

		inputs := map[string]any{"input": "value"}

		// Chain returns output without required key
		chain.On("Call", ctx, inputs, mock.Anything).Return(map[string]any{
			"wrong_output": "value",
		}, nil)

		result, err := callChain(ctx, chain, inputs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing key in output values")
		assert.NotNil(t, result) // callChain returns the output even on validation error
	})

	t.Run("chain call error", func(t *testing.T) {
		chain := &mockChain{
			inputKeys:  []string{"input"},
			outputKeys: []string{"output"},
		}

		inputs := map[string]any{"input": "value"}

		// Chain returns an error
		chainErr := errors.New("chain execution failed")
		chain.On("Call", ctx, inputs, mock.Anything).Return(nil, chainErr)

		result, err := callChain(ctx, chain, inputs)
		assert.Error(t, err)
		assert.Equal(t, chainErr, err)
		assert.Nil(t, result)
	})
}

func TestApplyStructs(t *testing.T) {
	t.Parallel()

	t.Run("applyInputJob", func(t *testing.T) {
		job := applyInputJob{
			input: map[string]any{"key": "value"},
			i:     42,
		}
		assert.Equal(t, map[string]any{"key": "value"}, job.input)
		assert.Equal(t, 42, job.i)
	})

	t.Run("applyResult", func(t *testing.T) {
		result := applyResult{
			result: map[string]any{"result": "value"},
			err:    errors.New("test error"),
			i:      24,
		}
		assert.Equal(t, map[string]any{"result": "value"}, result.result)
		assert.Equal(t, "test error", result.err.Error())
		assert.Equal(t, 24, result.i)
	})
}

func TestApplyWithZeroMaxWorkers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	inputs := []map[string]any{
		{"input": "test"},
	}

	chain := &mockChain{
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
	}

	memory := &mockMemory{}
	memory.On("MemoryVariables", ctx).Return([]string{})
	memory.On("LoadMemoryVariables", ctx, mock.Anything).Return(map[string]any{}, nil)
	memory.On("SaveContext", ctx, mock.Anything, mock.Anything).Return(nil)
	chain.memory = memory

	chain.On("Call", ctx, mock.Anything, mock.Anything).Return(map[string]any{
		"output": "result",
	}, nil)

	// Test with maxWorkers = 0, should default to _defaultApplyMaxNumberWorkers
	results, err := Apply(ctx, chain, inputs, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestRunWithMemoryKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	chain := &mockChain{
		inputKeys:  []string{"input", "memory_key"},
		outputKeys: []string{"output"},
	}

	// Memory provides one of the required inputs
	memory := &mockMemory{}
	memory.On("MemoryVariables", ctx).Return([]string{"memory_key"})
	memory.On("LoadMemoryVariables", ctx, mock.Anything).Return(map[string]any{"memory_key": "memory_value"}, nil)
	memory.On("SaveContext", ctx, mock.Anything, mock.Anything).Return(nil)
	chain.memory = memory

	chain.On("Call", ctx, mock.Anything, mock.Anything).Return(map[string]any{
		"output": "result",
	}, nil)

	result, err := Run(ctx, chain, "test_input")
	require.NoError(t, err)
	assert.Equal(t, "result", result)

	// Verify the chain was called with both the input and memory values
	chain.AssertCalled(t, "Call", ctx, map[string]any{
		"input":      "test_input",
		"memory_key": "memory_value",
	}, mock.Anything)
}
