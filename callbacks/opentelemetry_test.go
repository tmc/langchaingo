package callbacks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestOpenTelemetryCallbacksHandler_HandleAgentAction(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	action := schema.AgentAction{}

	handler.HandleAgentAction(ctx, action)
}

func TestOpenTelemetryCallbacksHandler_HandleAgentFinish(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	finish := schema.AgentFinish{}

	handler.HandleAgentFinish(ctx, finish)
}

func TestOpenTelemetryCallbacksHandler_HandleChain(t *testing.T) {
	ctx := context.Background()

	tracer := noop.NewTracerProvider().Tracer("test")
	handler, err := NewOpenTelemetryCallbacksHandler(tracer)
	assert.NoError(t, err)

	inputs := make(map[string]interface{})
	info := schema.ChainInfo{}
	next := func(ctx context.Context) (map[string]interface{}, error) {
		spanCtx := trace.SpanContextFromContext(ctx)
		assert.NotEmpty(t, spanCtx.SpanID().String())
		assert.NotEmpty(t, spanCtx.TraceID().String())
		return map[string]interface{}{
			"done": "right",
		}, nil
	}
	res, err := handler.HandleChain(ctx, inputs, info, next)

	assert.NoError(t, err)
	assert.Equal(t, res["done"], "right")
}

func TestOpenTelemetryCallbacksHandler_HandleChainEnd(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	outputs := make(map[string]interface{})

	handler.HandleChainEnd(ctx, outputs)
}

func TestOpenTelemetryCallbacksHandler_HandleChainError(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	err := assert.AnError

	handler.HandleChainError(ctx, err)
}

func TestOpenTelemetryCallbacksHandler_HandleChainStart(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	inputs := make(map[string]interface{})

	handler.HandleChainStart(ctx, inputs)
}

func TestOpenTelemetryCallbacksHandler_HandleLLM(t *testing.T) {
	ctx := context.Background()

	tracer := noop.NewTracerProvider().Tracer("test")
	handler, err := NewOpenTelemetryCallbacksHandler(tracer)
	assert.NoError(t, err)

	messages := []llms.MessageContent{}
	options := llms.CallOptions{}
	next := func(ctx context.Context) (*llms.ContentResponse, error) {
		spanCtx := trace.SpanContextFromContext(ctx)
		assert.NotEmpty(t, spanCtx.SpanID().String())
		assert.NotEmpty(t, spanCtx.TraceID().String())
		return &llms.ContentResponse{
			Choices: []*llms.ContentChoice{},
			Usage: &llms.Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
				TotalTokens:      2,
			},
		}, nil
	}
	res, err := handler.HandleLLM(ctx, messages, options, next)

	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestOpenTelemetryCallbacksHandler_HandleLLMError(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	err := assert.AnError

	handler.HandleLLMError(ctx, err)
}

func TestOpenTelemetryCallbacksHandler_HandleLLMGenerateContentEnd(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	res := &llms.ContentResponse{}

	handler.HandleLLMGenerateContentEnd(ctx, res)
}

func TestOpenTelemetryCallbacksHandler_HandleLLMGenerateContentStart(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	ms := []llms.MessageContent{}

	handler.HandleLLMGenerateContentStart(ctx, ms)
}

func TestOpenTelemetryCallbacksHandler_HandleLLMStart(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	prompts := []string{}

	handler.HandleLLMStart(ctx, prompts)
}

func TestOpenTelemetryCallbacksHandler_HandleRetrieverEnd(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	query := ""

	handler.HandleRetrieverEnd(ctx, query, nil)
}

func TestOpenTelemetryCallbacksHandler_HandleRetrieverStart(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	query := ""

	handler.HandleRetrieverStart(ctx, query)
}

func TestOpenTelemetryCallbacksHandler_HandleStreamingFunc(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	chunk := []byte{}

	handler.HandleStreamingFunc(ctx, chunk)
}

func TestOpenTelemetryCallbacksHandler_HandleText(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	text := ""

	handler.HandleText(ctx, text)
}

func TestOpenTelemetryCallbacksHandler_HandleToolEnd(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	output := ""

	handler.HandleToolEnd(ctx, output)
}

func TestOpenTelemetryCallbacksHandler_HandleToolError(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	err := assert.AnError

	handler.HandleToolError(ctx, err)
}

func TestOpenTelemetryCallbacksHandler_HandleToolStart(t *testing.T) {
	handler := &OpenTelemetryCallbacksHandler{}
	ctx := context.Background()
	input := ""

	handler.HandleToolStart(ctx, input)
}

func TestNewOpenTelemetryCallbacksHandler(t *testing.T) {
	tracer := noop.NewTracerProvider().Tracer("test")
	opts := []openTelemetryCallbacksHandlerOption{}

	handler, err := NewOpenTelemetryCallbacksHandler(tracer, opts...)

	assert.NoError(t, err)
	assert.NotNil(t, handler)
}
