// Package otelcallback provides an OpenTelemetry callback handler for
// LangChainGo. It instruments LLM, chain, tool, and retriever lifecycle
// events as OpenTelemetry traces.
//
// Privacy: prompt and completion contents are NOT captured by default.
// Use WithCaptureInput and WithCaptureOutput to opt in explicitly.
package otelcallback

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	instrumentationName    = "github.com/tmc/langchaingo/callbacks/otel"
	instrumentationVersion = "0.0.1"

	// GenAI semantic convention attribute keys.
	// Defined locally because go.opentelemetry.io/otel/semconv does not yet
	// expose stable GenAI constants. Update this file when they are stable.
	attrGenAIOperationName  = "gen_ai.operation.name"
	attrGenAIUsageInputTok  = "gen_ai.usage.input_tokens"
	attrGenAIUsageOutputTok = "gen_ai.usage.output_tokens"
	attrErrorType           = "error.type"
)

// Compile-time check: Handler must satisfy callbacks.Handler.
var _ callbacks.Handler = (*Handler)(nil)

// Handler is an OpenTelemetry callback handler that creates trace spans for
// LangChainGo lifecycle events.
type Handler struct {
	tracer        trace.Tracer
	captureInput  bool
	captureOutput bool

	// spans stores active trace.Span values keyed by context.Context.
	// This bridges the stateless callback interface (which cannot return
	// an enriched context) with the start/end span lifecycle.
	spans sync.Map
}

// NewHandler creates a new Handler. If no WithTracerProvider option is given,
// the global OpenTelemetry TracerProvider is used (which is a no-op tracer
// until the application configures a real exporter).
func NewHandler(opts ...Option) *Handler {
	cfg := config{
		tracerProvider: otel.GetTracerProvider(),
	}
	for _, o := range opts {
		o(&cfg)
	}
	return &Handler{
		tracer: cfg.tracerProvider.Tracer(
			instrumentationName,
			trace.WithInstrumentationVersion(instrumentationVersion),
		),
		captureInput:  cfg.captureInput,
		captureOutput: cfg.captureOutput,
	}
}

// HandleLLMStart starts a span for the legacy Call-based LLM path.
func (h *Handler) HandleLLMStart(ctx context.Context, prompts []string) {
	attrs := []attribute.KeyValue{
		attribute.String(attrGenAIOperationName, "llm.call"),
		attribute.Int("gen_ai.prompt.count", len(prompts)),
	}
	_, span := h.tracer.Start(ctx, "llm.call",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	h.spans.Store(ctx, span)
}

// HandleLLMGenerateContentStart starts a span for the GenerateContent path.
func (h *Handler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	attrs := []attribute.KeyValue{
		attribute.String(attrGenAIOperationName, "llm.generate_content"),
		attribute.Int("gen_ai.prompt.message_count", len(ms)),
	}
	_, span := h.tracer.Start(ctx, "llm.generate_content",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	h.spans.Store(ctx, span)
}

// HandleLLMGenerateContentEnd ends the GenerateContent span and records token usage.
func (h *Handler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	span := h.loadAndDelete(ctx)
	if span == nil {
		return
	}
	defer span.End()

	if res == nil {
		return
	}
	for _, choice := range res.Choices {
		if choice == nil {
			continue
		}
		if v, ok := toInt(choice.GenerationInfo["InputTokens"]); ok {
			span.SetAttributes(attribute.Int(attrGenAIUsageInputTok, v))
		}
		if v, ok := toInt(choice.GenerationInfo["OutputTokens"]); ok {
			span.SetAttributes(attribute.Int(attrGenAIUsageOutputTok, v))
		}
		if h.captureOutput && choice.Content != "" {
			span.SetAttributes(attribute.String("gen_ai.completion", choice.Content))
		}
		break
	}
}

// HandleLLMError records the error on the active LLM span and ends it.
func (h *Handler) HandleLLMError(ctx context.Context, err error) {
	h.endWithError(ctx, err)
}

// HandleChainStart starts a span for a chain execution.
func (h *Handler) HandleChainStart(ctx context.Context, _ map[string]any) {
	_, span := h.tracer.Start(ctx, "chain",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String(attrGenAIOperationName, "chain")),
	)
	h.spans.Store(ctx, span)
}

// HandleChainEnd ends the chain span.
func (h *Handler) HandleChainEnd(ctx context.Context, _ map[string]any) {
	if span := h.loadAndDelete(ctx); span != nil {
		span.End()
	}
}

// HandleChainError records the error on the chain span and ends it.
func (h *Handler) HandleChainError(ctx context.Context, err error) {
	h.endWithError(ctx, err)
}

// HandleToolStart starts a span for a tool execution.
func (h *Handler) HandleToolStart(ctx context.Context, input string) {
	attrs := []attribute.KeyValue{
		attribute.String(attrGenAIOperationName, "tool"),
	}
	if h.captureInput {
		attrs = append(attrs, attribute.String("tool.input", input))
	}
	_, span := h.tracer.Start(ctx, "tool.call",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
	h.spans.Store(ctx, span)
}

// HandleToolEnd ends the tool span.
func (h *Handler) HandleToolEnd(ctx context.Context, output string) {
	span := h.loadAndDelete(ctx)
	if span == nil {
		return
	}
	defer span.End()
	if h.captureOutput && output != "" {
		span.SetAttributes(attribute.String("tool.output", output))
	}
}

// HandleToolError records the error on the tool span and ends it.
func (h *Handler) HandleToolError(ctx context.Context, err error) {
	h.endWithError(ctx, err)
}

// HandleRetrieverStart starts a span for a retriever query.
func (h *Handler) HandleRetrieverStart(ctx context.Context, query string) {
	attrs := []attribute.KeyValue{
		attribute.String(attrGenAIOperationName, "retriever"),
	}
	if h.captureInput {
		attrs = append(attrs, attribute.String("retriever.query", query))
	}
	_, span := h.tracer.Start(ctx, "retriever.query",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attrs...),
	)
	h.spans.Store(ctx, span)
}

// HandleRetrieverEnd ends the retriever span.
func (h *Handler) HandleRetrieverEnd(ctx context.Context, _ string, documents []schema.Document) {
	span := h.loadAndDelete(ctx)
	if span == nil {
		return
	}
	defer span.End()
	span.SetAttributes(attribute.Int("retriever.document_count", len(documents)))
}

// HandleStreamingFunc is intentionally a no-op for tracing.
// Creating a span per chunk would produce extreme noise.
// Use metrics (counters/histograms) if you need streaming observability.
func (h *Handler) HandleStreamingFunc(_ context.Context, _ []byte) {}

// HandleText is a no-op to satisfy the Handler interface.
func (h *Handler) HandleText(_ context.Context, _ string) {}

// HandleAgentAction is a no-op to satisfy the Handler interface.
func (h *Handler) HandleAgentAction(_ context.Context, _ schema.AgentAction) {}

// HandleAgentFinish is a no-op to satisfy the Handler interface.
func (h *Handler) HandleAgentFinish(_ context.Context, _ schema.AgentFinish) {}

// endWithError is a shared helper: retrieves the active span for ctx,
// records the error, sets span status to Error, and ends the span.
func (h *Handler) endWithError(ctx context.Context, err error) {
	span := h.loadAndDelete(ctx)
	if span == nil {
		return
	}
	defer span.End()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String(attrErrorType, fmt.Sprintf("%T", err)))
	}
}

// loadAndDelete retrieves and removes the span stored under ctx.
// Returns nil safely if no span was found.
func (h *Handler) loadAndDelete(ctx context.Context) trace.Span {
	v, ok := h.spans.LoadAndDelete(ctx)
	if !ok {
		return nil
	}
	span, ok := v.(trace.Span)
	if !ok {
		return nil
	}
	return span
}

// toInt converts common numeric types found in GenerationInfo to int.
func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int32:
		return int(val), true
	case int64:
		return int(val), true
	case float32:
		return int(val), true
	case float64:
		return int(val), true
	}
	return 0, false
}

// intStr converts an int to string for use in attribute key construction.
func intStr(i int) string {
	return strconv.Itoa(i)
}
