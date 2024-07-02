package callbacks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type OpenTelemetryCallbacksHandler struct {
	tracer trace.Tracer
	opts   OpenTelemetryCallbacksHandlerOptions
}

// Span attributes should be named per semantic conventions where possible https://opentelemetry.io/docs/specs/semconv/gen-ai/llm-spans/
const (
	llmModel                        = attribute.Key("gen_ai.request.model")
	llmVendor                       = attribute.Key("gen_ai.system")
	llmMode                         = attribute.Key("gen_ai.mode")
	llmTemperature                  = attribute.Key("gen_ai.request.temperature")
	llmTopP                         = attribute.Key("gen_ai.top_p")
	llmTopK                         = attribute.Key("gen_ai.top_k")
	llmStopSequences                = attribute.Key("gen_ai.stop_sequences")
	llmFrequencyPenalty             = attribute.Key("gen_ai.frequency_penalty")
	llmPresencePenalty              = attribute.Key("gen_ai.presence_penalty")
	llmMaxTokens                    = attribute.Key("gen_ai.request.max_tokens")
	llmMessageIndex                 = attribute.Key("gen_ai.message.index")
	llmMessageRole                  = attribute.Key("gen_ai.message.role")
	llmMessageContent               = attribute.Key("gen_ai.message.content")
	llmCompletionsIndex             = attribute.Key("gen_ai.completion.index")
	llmCompletionsContent           = attribute.Key("gen_ai.completion.content")
	llmCompletionsStopReason        = attribute.Key("gen_ai.completion.finish_reasons")
	llmCompletionsGenerationInfo    = attribute.Key("gen_ai.completion.generation_info")
	llmCompletionsFuncCallName      = attribute.Key("gen_ai.completion.func_call.name")
	llmCompletionsFuncCallArguments = attribute.Key("gen_ai.completion.func_call.arguments")
	llmCompletionsToolCalls         = attribute.Key("gen_ai.completion.tool_calls")
	llmUsageTotalTokens             = attribute.Key("gen_ai.usage.total_tokens")
	llmUsagePromptTokens            = attribute.Key("gen_ai.usage.prompt_tokens")
	llmUsageCompletionTokens        = attribute.Key("gen_ai.usage.completion_tokens")
)

// semconvGenAIMessage is a struct that represents a message to/from a Gen-AI model in accordance with the semantic conventions for Gen-AI LLM spans.
type semconvGenAIMessage struct {
	Index   int
	Role    string
	Content string
}

// OpenTelemetryCallbacksHandlerOptions contains the options for the OpenTelemetryCallbacksHandler.
type OpenTelemetryCallbacksHandlerOptions struct {
	logPrompts     bool
	logCompletions bool
}

// OpenTelemetryCallbacksHandlerOption is a function that can be used to modify the behavior of the OpenTelemetryCallbacksHandler. An arbitrary number of such functions can be passed to NewOpenTelemetryCallbacksHandler.
type OpenTelemetryCallbacksHandlerOption func(*OpenTelemetryCallbacksHandlerOptions) *OpenTelemetryCallbacksHandlerOptions

// WithLogPrompts is an option to enable logging of prompts in the OpenTelemetryCallbacksHandler.
func WithLogPrompts(logPrompts bool) OpenTelemetryCallbacksHandlerOption {
	return func(opts *OpenTelemetryCallbacksHandlerOptions) *OpenTelemetryCallbacksHandlerOptions {
		opts.logPrompts = logPrompts
		return opts
	}
}

// WithLogCompletions is an option to enable logging of completions in the OpenTelemetryCallbacksHandler.
func WithLogCompletions(logCompletions bool) OpenTelemetryCallbacksHandlerOption {
	return func(opts *OpenTelemetryCallbacksHandlerOptions) *OpenTelemetryCallbacksHandlerOptions {
		opts.logCompletions = logCompletions
		return opts
	}
}

var _ Handler = &OpenTelemetryCallbacksHandler{}

// HandleAgentAction implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleAgentAction(_ context.Context, _ schema.AgentAction) {
}

// HandleAgentFinish implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleAgentFinish(_ context.Context, _ schema.AgentFinish) {
}

// HandleChain implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleChain(ctx context.Context, _ map[string]any, info schema.ChainInfo, next func(ctx context.Context) (map[string]any, error)) (map[string]any, error) {
	ctx, span := o.tracer.Start(ctx, "langchaingo.chain."+info.Name)
	defer span.End()
	return next(ctx)
}

// HandleChainEnd implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleChainEnd(_ context.Context, _ map[string]any) {
}

// HandleChainError implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleChainError(_ context.Context, _ error) {
}

// HandleChainStart implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleChainStart(_ context.Context, _ map[string]any) {
}

// HandleLLM implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleLLM(ctx context.Context, messages []llms.MessageContent, options llms.CallOptions, next func(ctx context.Context) (*llms.ContentResponse, error)) (*llms.ContentResponse, error) {
	ctx, span := o.tracer.Start(ctx, "gen_ai.model."+options.Model+".completion")
	defer span.End()
	response, err := next(ctx)
	if err != nil {
		span.RecordError(err)
	}

	o.processLLMRequest(span, messages, options)
	o.processLLMResponse(span, response, options)

	return response, err
}

// processLLMRequest records information about the request to the span.
func (o *OpenTelemetryCallbacksHandler) processLLMRequest(span trace.Span, messages []llms.MessageContent, options llms.CallOptions) {
	o.recordRequestAttributes(options, span)

	if o.opts.logPrompts {
		o.logPrompts(messages, span)
	}
}

// recordRequestAttributes records the request attributes to the span.
func (*OpenTelemetryCallbacksHandler) recordRequestAttributes(options llms.CallOptions, childSpan trace.Span) {
	llmModeValue := "text"
	if options.JSONMode {
		llmModeValue = "json"
	}

	requestAttributes := []attribute.KeyValue{
		llmModel.String(options.Model),
		llmVendor.String(options.Vendor),
		llmMode.String(llmModeValue),
		llmTemperature.Float64(float64(options.Temperature)),
		llmTopP.Float64(float64(options.TopP)),
		llmTopK.Float64(float64(options.TopK)),
		llmStopSequences.StringSlice(options.StopWords),
		llmFrequencyPenalty.Float64(options.FrequencyPenalty),
		llmPresencePenalty.Float64(float64(options.PresencePenalty)),
		llmMaxTokens.Int(options.MaxTokens),
	}
	childSpan.SetAttributes(requestAttributes...)
}

// logPrompts logs the prompts to the span as events.
func (o *OpenTelemetryCallbacksHandler) logPrompts(messages []llms.MessageContent, span trace.Span) {
	requestInfoMessages := o.transformPromptMessages(messages)

	for _, message := range requestInfoMessages {
		span.AddEvent("gen_ai.prompt", trace.WithAttributes(
			llmMessageIndex.Int(message.Index),
			llmMessageRole.String(message.Role),
			llmMessageContent.String(message.Content),
		))
	}
}

// transformPromptMessages transforms the Langchain-Go messages to a format that can be logged as events in the span, in accordance with the semantic conventions for GenAI LLM spans.
func (o *OpenTelemetryCallbacksHandler) transformPromptMessages(messages []llms.MessageContent) []semconvGenAIMessage {
	semconvGenAIMessages := []semconvGenAIMessage{}
	for idx, message := range messages {
		msg := &semconvGenAIMessage{
			Index:   idx,
			Role:    string(message.Role),
			Content: "",
		}
		for _, part := range message.Parts {
			switch part := part.(type) {
			case llms.TextContent:
				msg.Content += part.String()
			case llms.ToolCall:
				msg.Content += "[ToolCall] Function Name:" + part.FunctionCall.Name + "\nArgs: " + part.FunctionCall.Arguments
			case llms.ToolCallResponse:
				msg.Content += "[ToolCallResponse] Tool Name:" + part.Name
			case llms.ImageURLContent:
				msg.Content += "[ImageURLContent] Image URL length (bytes):" + fmt.Sprint(len(part.URL))
			case llms.BinaryContent:
				msg.Content += "[BinaryContent] Content length (bytes):" + fmt.Sprint(len(part.Data))
			}
		}
		semconvGenAIMessages = append(semconvGenAIMessages, *msg)
	}
	return semconvGenAIMessages
}

// processLLMResponse records information about the response to the span.
func (o *OpenTelemetryCallbacksHandler) processLLMResponse(span trace.Span, response *llms.ContentResponse, _ llms.CallOptions) {
	stopReasons := []string{}
	for idx, choice := range response.Choices {
		stopReasons = append(stopReasons, choice.StopReason)
		completionAttributes := []attribute.KeyValue{
			llmCompletionsIndex.Int(idx),
		}
		generationInfo, err := json.Marshal(choice.GenerationInfo)
		if err == nil {
			completionAttributes = append(completionAttributes, llmCompletionsGenerationInfo.String(string(generationInfo)))
		}
		if o.opts.logCompletions {
			completionAttributes = append(completionAttributes, llmCompletionsContent.String(choice.Content))
			if choice.FuncCall != nil {
				completionAttributes = append(completionAttributes, llmCompletionsFuncCallName.String(choice.FuncCall.Name))
				completionAttributes = append(completionAttributes, llmCompletionsFuncCallArguments.String(choice.FuncCall.Arguments))
			}
			if choice.ToolCalls != nil {
				toolCalls := []string{}
				for _, call := range choice.ToolCalls {
					toolCalls = append(toolCalls, fmt.Sprint(json.Marshal(call)))
				}
				completionAttributes = append(completionAttributes, llmCompletionsToolCalls.StringSlice(toolCalls))
			}
		}
		span.AddEvent("gen_ai.completion", trace.WithAttributes(completionAttributes...))
	}

	if response.Usage != nil {
		span.SetAttributes(
			llmUsageTotalTokens.Int(response.Usage.TotalTokens),
			llmUsagePromptTokens.Int(response.Usage.PromptTokens),
			llmUsageCompletionTokens.Int(response.Usage.CompletionTokens),
			llmCompletionsStopReason.StringSlice(stopReasons),
		)
	}
}

// HandleLLMError implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleLLMError(_ context.Context, _ error) {
}

// HandleLLMGenerateContentEnd implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleLLMGenerateContentEnd(_ context.Context, _ *llms.ContentResponse) {
}

// HandleLLMGenerateContentStart implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleLLMGenerateContentStart(_ context.Context, _ []llms.MessageContent) {
}

// HandleLLMStart implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleLLMStart(_ context.Context, _ []string) {
}

// HandleRetrieverEnd implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleRetrieverEnd(_ context.Context, _ string, _ []schema.Document) {
}

// HandleRetrieverStart implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleRetrieverStart(_ context.Context, _ string) {
}

// HandleStreamingFunc implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleStreamingFunc(_ context.Context, _ []byte) {
}

// HandleText implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleText(_ context.Context, _ string) {
}

// HandleToolEnd implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleToolEnd(_ context.Context, _ string) {
}

// HandleToolError implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleToolError(_ context.Context, _ error) {
}

// HandleToolStart implements Handler.
func (o *OpenTelemetryCallbacksHandler) HandleToolStart(_ context.Context, _ string) {
}

// NewOpenTelemetryCallbacksHandler creates a new OpenTelemetryCallbacksHandler.
func NewOpenTelemetryCallbacksHandler(t trace.Tracer, opts ...OpenTelemetryCallbacksHandlerOption) (*OpenTelemetryCallbacksHandler, error) {
	options := OpenTelemetryCallbacksHandlerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	return &OpenTelemetryCallbacksHandler{
		tracer: t,
		opts:   options,
	}, nil
}
