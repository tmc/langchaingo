package otelcallback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	otelcallback "github.com/tmc/langchaingo/callbacks/otel"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// newTestProvider returns an in-memory TracerProvider and its SpanRecorder.
func newTestProvider() (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	return tp, recorder
}

// TestHandlerImplementsInterface is a compile-time + runtime check.
func TestHandlerImplementsInterface(t *testing.T) {
	t.Parallel()
	tp, _ := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))
	// runtime interface check
	var _ interface{ HandleText(context.Context, string) } = h
	assert.NotNil(t, h)
}

// TestLLMGenerateContentStartCreatesSpan checks that a span is started.
func TestLLMGenerateContentStartCreatesSpan(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleLLMGenerateContentStart(ctx, []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{llms.TextContent{Text: "hello"}}},
	})

	// End the span by calling End
	h.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: "world", GenerationInfo: map[string]any{"InputTokens": 5, "OutputTokens": 3}},
		},
	})

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "llm.generate_content", spans[0].Name())
	assert.Equal(t, codes.Unset, spans[0].Status().Code)

	attrs := spanAttrMap(spans[0])
	assert.Equal(t, "llm.generate_content", attrs["gen_ai.operation.name"])
	assert.Equal(t, int64(5), attrs["gen_ai.usage.input_tokens"])
	assert.Equal(t, int64(3), attrs["gen_ai.usage.output_tokens"])
}

// TestLLMGenerateContentEndDoesNotCaptureOutputByDefault checks privacy default.
func TestLLMGenerateContentEndDoesNotCaptureOutputByDefault(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleLLMGenerateContentStart(ctx, nil)
	h.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "secret output"}},
	})

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	attrs := spanAttrMap(spans[0])
	_, captured := attrs["gen_ai.completion"]
	assert.False(t, captured, "completion must not be captured by default")
}

// TestLLMGenerateContentCapturesOutputWhenEnabled checks the opt-in capture.
func TestLLMGenerateContentCapturesOutputWhenEnabled(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(
		otelcallback.WithTracerProvider(tp),
		otelcallback.WithCaptureOutput(true),
	)

	ctx := context.Background()
	h.HandleLLMGenerateContentStart(ctx, nil)
	h.HandleLLMGenerateContentEnd(ctx, &llms.ContentResponse{
		Choices: []*llms.ContentChoice{{Content: "explicit output"}},
	})

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	attrs := spanAttrMap(spans[0])
	assert.Equal(t, "explicit output", attrs["gen_ai.completion"])
}

// TestLLMErrorSetsSpanStatusError checks error recording.
func TestLLMErrorSetsSpanStatusError(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleLLMGenerateContentStart(ctx, nil)
	h.HandleLLMError(ctx, errors.New("llm failed"))

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
	assert.Equal(t, "llm failed", spans[0].Status().Description)
	assert.NotEmpty(t, spans[0].Events(), "error event should be recorded")
}

// TestToolErrorSetsSpanStatusError checks tool error recording.
func TestToolErrorSetsSpanStatusError(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleToolStart(ctx, "some input")
	h.HandleToolError(ctx, errors.New("tool exploded"))

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "tool.call", spans[0].Name())
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

// TestRetrieverEndRecordsDocumentCount checks retriever span attributes.
func TestRetrieverEndRecordsDocumentCount(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleRetrieverStart(ctx, "find docs")
	h.HandleRetrieverEnd(ctx, "find docs", []schema.Document{{PageContent: "doc1"}, {PageContent: "doc2"}})

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "retriever.query", spans[0].Name())
	attrs := spanAttrMap(spans[0])
	assert.Equal(t, int64(2), attrs["retriever.document_count"])
}

// TestNilInputsDoNotPanic verifies the handler is nil-safe.
func TestNilInputsDoNotPanic(t *testing.T) {
	t.Parallel()
	tp, _ := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))
	ctx := context.Background()

	assert.NotPanics(t, func() {
		h.HandleLLMStart(ctx, nil)
		h.HandleLLMError(ctx, nil)
	})
	assert.NotPanics(t, func() {
		h.HandleLLMGenerateContentStart(ctx, nil)
		h.HandleLLMGenerateContentEnd(ctx, nil)
	})
	assert.NotPanics(t, func() {
		h.HandleChainStart(ctx, nil)
		h.HandleChainEnd(ctx, nil)
	})
	assert.NotPanics(t, func() {
		h.HandleToolStart(ctx, "")
		h.HandleToolEnd(ctx, "")
	})
	assert.NotPanics(t, func() {
		h.HandleRetrieverStart(ctx, "")
		h.HandleRetrieverEnd(ctx, "", nil)
	})
	assert.NotPanics(t, func() {
		h.HandleStreamingFunc(ctx, nil)
	})
}

// TestEndCalledWithoutStartDoesNotPanic checks orphan End calls are safe.
func TestEndCalledWithoutStartDoesNotPanic(t *testing.T) {
	t.Parallel()
	tp, _ := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))
	ctx := context.Background()

	assert.NotPanics(t, func() { h.HandleLLMGenerateContentEnd(ctx, nil) })
	assert.NotPanics(t, func() { h.HandleLLMError(ctx, errors.New("orphan")) })
	assert.NotPanics(t, func() { h.HandleChainEnd(ctx, nil) })
	assert.NotPanics(t, func() { h.HandleChainError(ctx, errors.New("orphan")) })
	assert.NotPanics(t, func() { h.HandleToolEnd(ctx, "") })
	assert.NotPanics(t, func() { h.HandleToolError(ctx, errors.New("orphan")) })
	assert.NotPanics(t, func() { h.HandleRetrieverEnd(ctx, "", nil) })
}

// TestCustomTracerProviderIsUsed verifies the option is respected.
func TestCustomTracerProviderIsUsed(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))

	ctx := context.Background()
	h.HandleChainStart(ctx, nil)
	h.HandleChainEnd(ctx, nil)

	assert.Len(t, recorder.Ended(), 1, "span should appear in the custom provider's recorder")
}

// TestStreamingDoesNotPanic verifies streaming chunks are silently ignored.
func TestStreamingDoesNotPanic(t *testing.T) {
	t.Parallel()
	tp, recorder := newTestProvider()
	h := otelcallback.NewHandler(otelcallback.WithTracerProvider(tp))
	ctx := context.Background()

	assert.NotPanics(t, func() {
		for i := 0; i < 10; i++ {
			h.HandleStreamingFunc(ctx, []byte("chunk"))
		}
	})
	assert.Empty(t, recorder.Ended(), "streaming must not create spans")
}

// spanAttrMap converts a recorded span's attributes to a map for easy assertion.
func spanAttrMap(span sdktrace.ReadOnlySpan) map[string]interface{} {
	m := make(map[string]interface{})
	for _, kv := range span.Attributes() {
		switch kv.Value.Type() {
		case attribute.STRING:
			m[string(kv.Key)] = kv.Value.AsString()
		case attribute.INT64:
			m[string(kv.Key)] = kv.Value.AsInt64()
		case attribute.BOOL:
			m[string(kv.Key)] = kv.Value.AsBool()
		case attribute.FLOAT64:
			m[string(kv.Key)] = kv.Value.AsFloat64()
		}
	}
	return m
}

// ensure trace import is used (for SpanKind constant visibility in test file)
var _ = trace.SpanKindClient
