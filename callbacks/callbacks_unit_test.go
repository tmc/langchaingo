package callbacks

import (
	"context"
	"testing"

	"github.com/0xDezzy/langchaingo/llms"
	"github.com/0xDezzy/langchaingo/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Unit tests that don't require external dependencies

type testHandlerHaver struct {
	handler Handler
}

func (t *testHandlerHaver) GetCallbackHandler() Handler {
	return t.handler
}

type mockHandler struct {
	mock.Mock
}

func (m *mockHandler) HandleText(ctx context.Context, text string) {
	m.Called(ctx, text)
}

func (m *mockHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	m.Called(ctx, prompts)
}

func (m *mockHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	m.Called(ctx, ms)
}

func (m *mockHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	m.Called(ctx, res)
}

func (m *mockHandler) HandleLLMError(ctx context.Context, err error) {
	m.Called(ctx, err)
}

func (m *mockHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	m.Called(ctx, inputs)
}

func (m *mockHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	m.Called(ctx, outputs)
}

func (m *mockHandler) HandleChainError(ctx context.Context, err error) {
	m.Called(ctx, err)
}

func (m *mockHandler) HandleToolStart(ctx context.Context, input string) {
	m.Called(ctx, input)
}

func (m *mockHandler) HandleToolEnd(ctx context.Context, output string) {
	m.Called(ctx, output)
}

func (m *mockHandler) HandleToolError(ctx context.Context, err error) {
	m.Called(ctx, err)
}

func (m *mockHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	m.Called(ctx, action)
}

func (m *mockHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	m.Called(ctx, finish)
}

func (m *mockHandler) HandleRetrieverStart(ctx context.Context, query string) {
	m.Called(ctx, query)
}

func (m *mockHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	m.Called(ctx, query, documents)
}

func (m *mockHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	m.Called(ctx, chunk)
}

func TestSimpleHandler(t *testing.T) {
	t.Parallel()

	// Test that SimpleHandler implements Handler interface
	var _ Handler = SimpleHandler{}

	ctx := context.Background()
	handler := SimpleHandler{}

	// Test all methods run without error (they're all no-ops)
	handler.HandleText(ctx, "test")
	handler.HandleLLMStart(ctx, []string{"prompt"})
	handler.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{})
	handler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{})
	handler.HandleLLMError(ctx, assert.AnError)
	handler.HandleChainStart(ctx, map[string]any{"input": "value"})
	handler.HandleChainEnd(ctx, map[string]any{"output": "value"})
	handler.HandleChainError(ctx, assert.AnError)
	handler.HandleToolStart(ctx, "tool input")
	handler.HandleToolEnd(ctx, "tool output")
	handler.HandleToolError(ctx, assert.AnError)
	handler.HandleAgentAction(ctx, schema.AgentAction{})
	handler.HandleAgentFinish(ctx, schema.AgentFinish{})
	handler.HandleRetrieverStart(ctx, "query")
	handler.HandleRetrieverEnd(ctx, "query", []schema.Document{})
	handler.HandleStreamingFunc(ctx, []byte("chunk"))

	// No assertions needed - if we get here, all methods executed without panic
}

func TestCombiningHandler(t *testing.T) {
	t.Parallel()

	// Test that CombiningHandler implements Handler interface
	var _ Handler = CombiningHandler{}

	ctx := context.Background()

	t.Run("empty callbacks", func(t *testing.T) {
		handler := CombiningHandler{Callbacks: []Handler{}}

		// All methods should work with empty callbacks
		handler.HandleText(ctx, "test")
		handler.HandleLLMStart(ctx, []string{"prompt"})
		handler.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{})
		handler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{})
		handler.HandleLLMError(ctx, assert.AnError)
		handler.HandleChainStart(ctx, map[string]any{"input": "value"})
		handler.HandleChainEnd(ctx, map[string]any{"output": "value"})
		handler.HandleChainError(ctx, assert.AnError)
		handler.HandleToolStart(ctx, "tool input")
		handler.HandleToolEnd(ctx, "tool output")
		handler.HandleToolError(ctx, assert.AnError)
		handler.HandleAgentAction(ctx, schema.AgentAction{})
		handler.HandleAgentFinish(ctx, schema.AgentFinish{})
		handler.HandleRetrieverStart(ctx, "query")
		handler.HandleRetrieverEnd(ctx, "query", []schema.Document{})
		handler.HandleStreamingFunc(ctx, []byte("chunk"))
	})

	t.Run("single callback", func(t *testing.T) {
		mock1 := &mockHandler{}
		handler := CombiningHandler{Callbacks: []Handler{mock1}}

		// Set up expectations
		mock1.On("HandleText", ctx, "test")
		mock1.On("HandleLLMStart", ctx, []string{"prompt"})
		mock1.On("HandleLLMGenerateContentStart", ctx, []llms.MessageContent{})
		mock1.On("HandleLLMGenerateContentEnd", ctx, &llms.ContentResponse{})
		mock1.On("HandleLLMError", ctx, assert.AnError)
		mock1.On("HandleChainStart", ctx, map[string]any{"input": "value"})
		mock1.On("HandleChainEnd", ctx, map[string]any{"output": "value"})
		mock1.On("HandleChainError", ctx, assert.AnError)
		mock1.On("HandleToolStart", ctx, "tool input")
		mock1.On("HandleToolEnd", ctx, "tool output")
		mock1.On("HandleToolError", ctx, assert.AnError)
		mock1.On("HandleAgentAction", ctx, schema.AgentAction{})
		mock1.On("HandleAgentFinish", ctx, schema.AgentFinish{})
		mock1.On("HandleRetrieverStart", ctx, "query")
		mock1.On("HandleRetrieverEnd", ctx, "query", []schema.Document{})
		mock1.On("HandleStreamingFunc", ctx, []byte("chunk"))

		// Call all methods
		handler.HandleText(ctx, "test")
		handler.HandleLLMStart(ctx, []string{"prompt"})
		handler.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{})
		handler.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{})
		handler.HandleLLMError(ctx, assert.AnError)
		handler.HandleChainStart(ctx, map[string]any{"input": "value"})
		handler.HandleChainEnd(ctx, map[string]any{"output": "value"})
		handler.HandleChainError(ctx, assert.AnError)
		handler.HandleToolStart(ctx, "tool input")
		handler.HandleToolEnd(ctx, "tool output")
		handler.HandleToolError(ctx, assert.AnError)
		handler.HandleAgentAction(ctx, schema.AgentAction{})
		handler.HandleAgentFinish(ctx, schema.AgentFinish{})
		handler.HandleRetrieverStart(ctx, "query")
		handler.HandleRetrieverEnd(ctx, "query", []schema.Document{})
		handler.HandleStreamingFunc(ctx, []byte("chunk"))

		// Verify all expectations were met
		mock1.AssertExpectations(t)
	})

	t.Run("multiple callbacks", func(t *testing.T) {
		mock1 := &mockHandler{}
		mock2 := &mockHandler{}
		handler := CombiningHandler{Callbacks: []Handler{mock1, mock2}}

		// Set up expectations for both mocks
		for _, m := range []*mockHandler{mock1, mock2} {
			m.On("HandleText", ctx, "test")
			m.On("HandleChainStart", ctx, map[string]any{"input": "value"})
			m.On("HandleToolError", ctx, assert.AnError)
		}

		// Call methods
		handler.HandleText(ctx, "test")
		handler.HandleChainStart(ctx, map[string]any{"input": "value"})
		handler.HandleToolError(ctx, assert.AnError)

		// Verify all expectations were met
		mock1.AssertExpectations(t)
		mock2.AssertExpectations(t)
	})
}

func TestCombiningHandlerStructure(t *testing.T) {
	t.Parallel()

	// Test struct initialization
	handler := CombiningHandler{
		Callbacks: []Handler{
			SimpleHandler{},
			SimpleHandler{},
		},
	}

	assert.Len(t, handler.Callbacks, 2)

	// Test that we can access the callbacks
	for _, callback := range handler.Callbacks {
		assert.NotNil(t, callback)
		// Verify each callback implements Handler interface
		assert.Implements(t, (*Handler)(nil), callback)
	}
}

func TestHandlerInterfaceCompleteness(t *testing.T) {
	t.Parallel()

	// Verify that our mock handler implements all methods of the Handler interface
	var _ Handler = &mockHandler{}

	// Verify that SimpleHandler implements all methods of the Handler interface
	var _ Handler = SimpleHandler{}

	// Verify that CombiningHandler implements all methods of the Handler interface
	var _ Handler = CombiningHandler{}
}

func TestHandlerHaverInterface(t *testing.T) {
	t.Parallel()

	// Verify interface implementation
	var _ HandlerHaver = &testHandlerHaver{}

	handler := SimpleHandler{}
	haver := &testHandlerHaver{handler: handler}

	retrieved := haver.GetCallbackHandler()
	assert.Equal(t, handler, retrieved)
}

func TestComplexCombiningScenario(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create a complex scenario with nested combining handlers
	mock1 := &mockHandler{}
	mock2 := &mockHandler{}

	innerCombining := CombiningHandler{Callbacks: []Handler{mock1}}
	outerCombining := CombiningHandler{Callbacks: []Handler{innerCombining, mock2}}

	// Set up expectations
	mock1.On("HandleText", ctx, "nested test")
	mock2.On("HandleText", ctx, "nested test")

	// Call method
	outerCombining.HandleText(ctx, "nested test")

	// Verify expectations
	mock1.AssertExpectations(t)
	mock2.AssertExpectations(t)
}

func TestCombiningHandlerWithMixedTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Test combining different handler types
	mock := &mockHandler{}
	simple := SimpleHandler{}

	handler := CombiningHandler{
		Callbacks: []Handler{mock, simple},
	}

	mock.On("HandleLLMError", ctx, assert.AnError)

	// This should call the mock and the simple handler (which is a no-op)
	handler.HandleLLMError(ctx, assert.AnError)

	mock.AssertExpectations(t)
}

func TestCallbackTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name     string
		callFunc func(Handler)
	}{
		{
			name: "HandleText",
			callFunc: func(h Handler) {
				h.HandleText(ctx, "sample text")
			},
		},
		{
			name: "HandleLLMStart",
			callFunc: func(h Handler) {
				h.HandleLLMStart(ctx, []string{"prompt1", "prompt2"})
			},
		},
		{
			name: "HandleLLMGenerateContentStart",
			callFunc: func(h Handler) {
				h.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{
					{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextPart("test")}},
				})
			},
		},
		{
			name: "HandleLLMGenerateContentEnd",
			callFunc: func(h Handler) {
				h.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
					Choices: []*llms.ContentChoice{{Content: "response"}},
				})
			},
		},
		{
			name: "HandleAgentAction",
			callFunc: func(h Handler) {
				h.HandleAgentAction(ctx, schema.AgentAction{
					Tool:      "calculator",
					ToolInput: "2+2",
					Log:       "Using calculator",
				})
			},
		},
		{
			name: "HandleAgentFinish",
			callFunc: func(h Handler) {
				h.HandleAgentFinish(ctx, schema.AgentFinish{
					ReturnValues: map[string]any{"result": "4"},
					Log:          "Calculation complete",
				})
			},
		},
		{
			name: "HandleRetrieverEnd",
			callFunc: func(h Handler) {
				h.HandleRetrieverEnd(ctx, "search query", []schema.Document{
					{PageContent: "document content", Metadata: map[string]any{"source": "test"}},
				})
			},
		},
		{
			name: "HandleStreamingFunc",
			callFunc: func(h Handler) {
				h.HandleStreamingFunc(ctx, []byte("streaming chunk"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with SimpleHandler (should not panic)
			simple := SimpleHandler{}
			tt.callFunc(simple)

			// Test with CombiningHandler with empty callbacks
			combining := CombiningHandler{Callbacks: []Handler{}}
			tt.callFunc(combining)

			// Test with CombiningHandler with SimpleHandler
			combiningWithSimple := CombiningHandler{Callbacks: []Handler{simple}}
			tt.callFunc(combiningWithSimple)
		})
	}
}
