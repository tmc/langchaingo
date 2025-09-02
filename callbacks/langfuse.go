package callbacks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

// LangfuseHandler is a callback handler that sends traces to Langfuse.
type LangfuseHandler struct {
	baseURL    string
	publicKey  string
	secretKey  string
	httpClient *http.Client
	traceID    string
	sessionID  string
	userID     string
	metadata   map[string]interface{}
	mu         sync.RWMutex
	spans      map[string]*LangfuseSpan
	traces     map[string]*LangfuseTrace
}

// LangfuseOptions holds configuration options for the Langfuse handler.
type LangfuseOptions struct {
	BaseURL   string
	PublicKey string
	SecretKey string
	TraceID   string
	SessionID string
	UserID    string
	Metadata  map[string]interface{}
}

// LangfuseTrace represents a trace in Langfuse.
type LangfuseTrace struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name,omitempty"`
	UserID    string                 `json:"userId,omitempty"`
	SessionID string                 `json:"sessionId,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Input     interface{}            `json:"input,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Release   string                 `json:"release,omitempty"`
	Version   string                 `json:"version,omitempty"`
}

// LangfuseSpan represents a span in Langfuse.
type LangfuseSpan struct {
	ID            string                 `json:"id"`
	TraceID       string                 `json:"traceId"`
	ParentSpanID  string                 `json:"parentObservationId,omitempty"`
	Name          string                 `json:"name"`
	StartTime     time.Time              `json:"startTime"`
	EndTime       *time.Time             `json:"endTime,omitempty"`
	Level         string                 `json:"level,omitempty"`
	StatusMessage string                 `json:"statusMessage,omitempty"`
	Input         interface{}            `json:"input,omitempty"`
	Output        interface{}            `json:"output,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Model         string                 `json:"model,omitempty"`
	ModelParams   map[string]interface{} `json:"modelParameters,omitempty"`
	Usage         *LangfuseUsage         `json:"usage,omitempty"`
}

// LangfuseUsage represents token usage information.
type LangfuseUsage struct {
	Input  int `json:"input,omitempty"`
	Output int `json:"output,omitempty"`
	Total  int `json:"total,omitempty"`
}

// NewLangfuseHandler creates a new LangfuseHandler with the provided options.
func NewLangfuseHandler(opts LangfuseOptions) (*LangfuseHandler, error) {
	if opts.BaseURL == "" {
		opts.BaseURL = "https://cloud.langfuse.com"
	}
	if opts.PublicKey == "" || opts.SecretKey == "" {
		return nil, fmt.Errorf("langfuse public key and secret key are required")
	}
	if opts.TraceID == "" {
		opts.TraceID = uuid.New().String()
	}

	handler := &LangfuseHandler{
		baseURL:    opts.BaseURL,
		publicKey:  opts.PublicKey,
		secretKey:  opts.SecretKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		traceID:    opts.TraceID,
		sessionID:  opts.SessionID,
		userID:     opts.UserID,
		metadata:   opts.Metadata,
		spans:      make(map[string]*LangfuseSpan),
		traces:     make(map[string]*LangfuseTrace),
	}

	// Create initial trace
	trace := &LangfuseTrace{
		ID:        handler.traceID,
		UserID:    handler.userID,
		SessionID: handler.sessionID,
		Metadata:  handler.metadata,
		Timestamp: time.Now(),
	}
	handler.traces[handler.traceID] = trace

	return handler, nil
}

var _ Handler = (*LangfuseHandler)(nil)

// HandleText implements the Handler interface.
func (h *LangfuseHandler) HandleText(ctx context.Context, text string) {
	// Text events are typically logged as events rather than spans
}

// HandleLLMStart implements the Handler interface.
func (h *LangfuseHandler) HandleLLMStart(ctx context.Context, prompts []string) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "llm",
		StartTime: time.Now(),
		Input:     prompts,
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	// Store span ID in context for later use
	h.storeSpanID(ctx, spanID)
}

// HandleLLMGenerateContentStart implements the Handler interface.
func (h *LangfuseHandler) HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent) {
	spanID := uuid.New().String()
	
	// Convert messages to a more readable format
	input := make([]map[string]interface{}, len(ms))
	for i, msg := range ms {
		parts := make([]string, len(msg.Parts))
		for j, part := range msg.Parts {
			if textPart, ok := part.(llms.TextContent); ok {
				parts[j] = textPart.Text
			} else {
				parts[j] = fmt.Sprintf("%v", part)
			}
		}
		input[i] = map[string]interface{}{
			"role":  string(msg.Role),
			"parts": parts,
		}
	}

	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "llm-generation",
		StartTime: time.Now(),
		Input:     input,
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	h.storeSpanID(ctx, spanID)
}

// HandleLLMGenerateContentEnd implements the Handler interface.
func (h *LangfuseHandler) HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	// Extract output and usage information
	output := make([]map[string]interface{}, len(res.Choices))
	totalUsage := &LangfuseUsage{}
	
	for i, choice := range res.Choices {
		output[i] = map[string]interface{}{
			"content":     choice.Content,
			"stopReason":  choice.StopReason,
			"funcCall":    choice.FuncCall,
			"toolCalls":   choice.ToolCalls,
		}
	}

	// Extract usage information if available
	// Note: Usage information may be available in GenerationInfo
	if len(res.Choices) > 0 && len(res.Choices[0].GenerationInfo) > 0 {
		if promptTokens, ok := res.Choices[0].GenerationInfo["prompt_tokens"].(int); ok {
			totalUsage.Input = promptTokens
		}
		if completionTokens, ok := res.Choices[0].GenerationInfo["completion_tokens"].(int); ok {
			totalUsage.Output = completionTokens
		}
		if totalTokens, ok := res.Choices[0].GenerationInfo["total_tokens"].(int); ok {
			totalUsage.Total = totalTokens
		}
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Output = output
	span.Usage = totalUsage

	// Extract model information from generation info
	if len(res.Choices) > 0 && len(res.Choices[0].GenerationInfo) > 0 {
		if model, ok := res.Choices[0].GenerationInfo["model"].(string); ok {
			span.Model = model
		}
		span.ModelParams = res.Choices[0].GenerationInfo
	}

	h.mu.Unlock()

	// Send span to Langfuse
	h.sendSpan(span)
}

// HandleLLMError implements the Handler interface.
func (h *LangfuseHandler) HandleLLMError(ctx context.Context, err error) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Level = "ERROR"
	span.StatusMessage = err.Error()
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleChainStart implements the Handler interface.
func (h *LangfuseHandler) HandleChainStart(ctx context.Context, inputs map[string]any) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "chain",
		StartTime: time.Now(),
		Input:     inputs,
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	h.storeSpanID(ctx, spanID)
}

// HandleChainEnd implements the Handler interface.
func (h *LangfuseHandler) HandleChainEnd(ctx context.Context, outputs map[string]any) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Output = outputs
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleChainError implements the Handler interface.
func (h *LangfuseHandler) HandleChainError(ctx context.Context, err error) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Level = "ERROR"
	span.StatusMessage = err.Error()
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleToolStart implements the Handler interface.
func (h *LangfuseHandler) HandleToolStart(ctx context.Context, input string) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "tool",
		StartTime: time.Now(),
		Input:     input,
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	h.storeSpanID(ctx, spanID)
}

// HandleToolEnd implements the Handler interface.
func (h *LangfuseHandler) HandleToolEnd(ctx context.Context, output string) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Output = output
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleToolError implements the Handler interface.
func (h *LangfuseHandler) HandleToolError(ctx context.Context, err error) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Level = "ERROR"
	span.StatusMessage = err.Error()
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleAgentAction implements the Handler interface.
func (h *LangfuseHandler) HandleAgentAction(ctx context.Context, action schema.AgentAction) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "agent-action",
		StartTime: time.Now(),
		Input: map[string]interface{}{
			"tool":      action.Tool,
			"toolInput": action.ToolInput,
			"log":       action.Log,
		},
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	// Agent actions are typically immediately completed
	endTime := time.Now()
	span.EndTime = &endTime
	h.sendSpan(span)
}

// HandleAgentFinish implements the Handler interface.
func (h *LangfuseHandler) HandleAgentFinish(ctx context.Context, finish schema.AgentFinish) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "agent-finish",
		StartTime: time.Now(),
		Input:     finish.ReturnValues,
		Output:    finish.Log,
	}

	endTime := time.Now()
	span.EndTime = &endTime

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleRetrieverStart implements the Handler interface.
func (h *LangfuseHandler) HandleRetrieverStart(ctx context.Context, query string) {
	spanID := uuid.New().String()
	span := &LangfuseSpan{
		ID:        spanID,
		TraceID:   h.traceID,
		Name:      "retriever",
		StartTime: time.Now(),
		Input:     query,
	}

	h.mu.Lock()
	h.spans[spanID] = span
	h.mu.Unlock()

	h.storeSpanID(ctx, spanID)
}

// HandleRetrieverEnd implements the Handler interface.
func (h *LangfuseHandler) HandleRetrieverEnd(ctx context.Context, query string, documents []schema.Document) {
	spanID := h.getSpanID(ctx)
	if spanID == "" {
		return
	}

	h.mu.Lock()
	span, exists := h.spans[spanID]
	if !exists {
		h.mu.Unlock()
		return
	}

	endTime := time.Now()
	span.EndTime = &endTime
	span.Output = map[string]interface{}{
		"documents": documents,
		"count":     len(documents),
	}
	h.mu.Unlock()

	h.sendSpan(span)
}

// HandleStreamingFunc implements the Handler interface.
func (h *LangfuseHandler) HandleStreamingFunc(ctx context.Context, chunk []byte) {
	// Streaming chunks are typically not sent as individual spans
	// but could be accumulated and sent at the end of a stream
}

// Helper methods

func (h *LangfuseHandler) storeSpanID(ctx context.Context, spanID string) {
	// In a real implementation, you might use context.WithValue
	// For now, we'll store the latest span ID as a simple approach
	h.mu.Lock()
	h.spans["_current"] = &LangfuseSpan{ID: spanID}
	h.mu.Unlock()
}

func (h *LangfuseHandler) getSpanID(ctx context.Context) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if span, exists := h.spans["_current"]; exists {
		return span.ID
	}
	return ""
}

func (h *LangfuseHandler) sendSpan(span *LangfuseSpan) {
	go func() {
		payload := map[string]interface{}{
			"id":       span.ID,
			"type":     "SPAN",
			"body":     span,
			"metadata": map[string]interface{}{
				"sdk": map[string]interface{}{
					"name":    "langchaingo",
					"version": "1.0.0",
				},
			},
		}

		if err := h.sendToLangfuse(payload); err != nil {
			// In a production environment, you might want to use a proper logger
			fmt.Printf("Failed to send span to Langfuse: %v\n", err)
		}
	}()
}

func (h *LangfuseHandler) sendToLangfuse(payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", h.baseURL+"/api/public/ingestion", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(h.publicKey, h.secretKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("langfuse API returned status: %d", resp.StatusCode)
	}

	return nil
}

// Flush sends any pending traces to Langfuse.
func (h *LangfuseHandler) Flush() error {
	h.mu.RLock()
	traces := make([]*LangfuseTrace, 0, len(h.traces))
	for _, trace := range h.traces {
		traces = append(traces, trace)
	}
	h.mu.RUnlock()

	for _, trace := range traces {
		payload := map[string]interface{}{
			"id":   trace.ID,
			"type": "TRACE_CREATE",
			"body": trace,
			"metadata": map[string]interface{}{
				"sdk": map[string]interface{}{
					"name":    "langchaingo",
					"version": "1.0.0",
				},
			},
		}

		if err := h.sendToLangfuse(payload); err != nil {
			return fmt.Errorf("failed to send trace %s: %w", trace.ID, err)
		}
	}

	return nil
}

// SetTraceMetadata updates the metadata for the current trace.
func (h *LangfuseHandler) SetTraceMetadata(metadata map[string]interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if trace, exists := h.traces[h.traceID]; exists {
		if trace.Metadata == nil {
			trace.Metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			trace.Metadata[k] = v
		}
	}
}

// GetTraceID returns the current trace ID.
func (h *LangfuseHandler) GetTraceID() string {
	return h.traceID
}