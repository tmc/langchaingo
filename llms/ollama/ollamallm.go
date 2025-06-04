package ollama

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama/internal/ollamaclient"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
)

// LLM is a ollama LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *ollamaclient.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new ollama LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	client, err := ollamaclient.NewClient(o.ollamaServerURL, o.httpClient)
	if err != nil {
		return nil, err
	}

	return &LLM{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
// nolint: goerr113
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Override LLM model if set as llms.CallOption
	model := o.options.model
	if opts.Model != "" {
		model = opts.Model
	}

	// Our input is a sequence of MessageContent, each of which potentially has
	// a sequence of Part that could be text, images etc.
	// We have to convert it to a format Ollama understands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs, err := makeOllamaMessages(messages)
	if err != nil {
		return nil, err
	}

	format := o.options.format
	if opts.JSONMode {
		format = "json"
	}

	tools := o.options.tools
	if len(opts.Tools) > 0 {
		tools = makeOllamaTools(opts.Tools)
	}

	// Get our ollamaOptions from llms.CallOptions
	ollamaOptions := makeOllamaOptions(o.options.ollamaOptions, opts)
	req := &ollamaclient.ChatRequest{
		Model:    model,
		Format:   format,
		Messages: chatMsgs,
		Options:  ollamaOptions,
		Stream:   opts.StreamingFunc != nil && len(opts.Tools) == 0,
		Tools:    tools,
	}

	keepAlive := o.options.keepAlive
	if keepAlive != "" {
		req.KeepAlive = keepAlive
	}

	var fn ollamaclient.ChatResponseFunc
	streamedResponse := ""
	var resp ollamaclient.ChatResponse

	fn = func(response ollamaclient.ChatResponse) error {
		if opts.StreamingFunc != nil && response.Message != nil {
			if err := opts.StreamingFunc(ctx, []byte(response.Message.Content)); err != nil {
				return err
			}
		}
		if response.Message != nil {
			streamedResponse += response.Message.Content
		}

		if !req.Stream || response.Done {
			resp = response
			resp.Message = &ollamaclient.Message{
				Role:      "assistant",
				Content:   streamedResponse,
				ToolCalls: response.Message.ToolCalls,
			}
		}
		return nil
	}

	err = o.client.GenerateChat(ctx, req, fn)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	toolCalls := makeLLMSToolCall(resp.Message.ToolCalls)

	choices := []*llms.ContentChoice{
		{
			Content: resp.Message.Content,
			GenerationInfo: map[string]any{
				"CompletionTokens": resp.EvalCount,
				"PromptTokens":     resp.PromptEvalCount,
				"TotalTokens":      resp.EvalCount + resp.PromptEvalCount,
			},
			ToolCalls: toolCalls,
		},
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings := [][]float32{}

	for _, input := range inputTexts {
		req := &ollamaclient.EmbeddingRequest{
			Prompt: input,
			Model:  o.options.model,
		}
		if o.options.keepAlive != "" {
			req.KeepAlive = o.options.keepAlive
		}

		embedding, err := o.client.CreateEmbedding(ctx, req)
		if err != nil {
			return nil, err
		}

		if len(embedding.Embedding) == 0 {
			return nil, ErrEmptyResponse
		}

		embeddings = append(embeddings, embedding.Embedding)
	}

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrIncompleteEmbedding
	}

	return embeddings, nil
}

// makeOllamaOptions make ollamaclient.Options from llms.CallOptions.
func makeOllamaOptions(ollamaOptions ollamaclient.Options, opts llms.CallOptions) ollamaclient.Options {
	// Load back CallOptions as ollamaOptions
	ollamaOptions.NumPredict = opts.MaxTokens
	ollamaOptions.Temperature = float32(opts.Temperature)
	ollamaOptions.Stop = opts.StopWords
	ollamaOptions.TopK = opts.TopK
	ollamaOptions.TopP = float32(opts.TopP)
	ollamaOptions.Seed = opts.Seed
	ollamaOptions.RepeatPenalty = float32(opts.RepetitionPenalty)
	ollamaOptions.FrequencyPenalty = float32(opts.FrequencyPenalty)
	ollamaOptions.PresencePenalty = float32(opts.PresencePenalty)

	return ollamaOptions
}

// makeOllamaTools make ollamaclient.Tool from llms.Tool.
func makeOllamaTools(tools []llms.Tool) []ollamaclient.Tool {
	ollamaTools := make([]ollamaclient.Tool, 0, len(tools))
	for _, tool := range tools {
		functionDef := ollamaclient.ToolFunction{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			Parameters:  tool.Function.Parameters,
		}
		ollamaTools = append(ollamaTools, ollamaclient.Tool{
			Type:     tool.Type,
			Function: functionDef,
		})
	}
	return ollamaTools
}

// makeOllamaMessages make ollamaclient.Message from message llms.MessageContent.
func makeOllamaMessages(messages []llms.MessageContent) ([]*ollamaclient.Message, error) {
	chatMsgs := make([]*ollamaclient.Message, 0, len(messages))
	for _, mc := range messages {
		msg := &ollamaclient.Message{}
		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			msg.Role = "system"
		case llms.ChatMessageTypeAI:
			msg.Role = "assistant"
		case llms.ChatMessageTypeHuman:
			fallthrough
		case llms.ChatMessageTypeGeneric:
			msg.Role = "user"
		case llms.ChatMessageTypeFunction:
			msg.Role = "function"
		case llms.ChatMessageTypeTool:
			msg.Role = "tool"

			if len(mc.Parts) != 1 {
				return nil, fmt.Errorf("expected exactly one part for role %v, got %v", mc.Role, len(mc.Parts))
			}
			switch p := mc.Parts[0].(type) {
			case llms.ToolCallResponse:
				msg.Content = p.Content
			default:
				return nil, fmt.Errorf("expected part of type ToolCallResponse for role %v, got %T", mc.Role, mc.Parts[0])
			}
		}

		text, images, tools, err := makeOllamaContent(mc)
		if err != nil {
			return nil, err
		}

		msg.Content = text
		msg.Images = images
		msg.ToolCalls = tools
		chatMsgs = append(chatMsgs, msg)
	}

	return chatMsgs, nil
}

// makeOllamaContent make ollamaclient Content, ImageData and ToolCall from llms.MessageContent.
func makeOllamaContent(mc llms.MessageContent) (string, []ollamaclient.ImageData, []ollamaclient.ToolCall, error) {
	// Look at all the parts in mc; expect to find a single Text part and
	// any number of binary parts.
	var text string
	foundText := false
	var images []ollamaclient.ImageData
	var tools []ollamaclient.ToolCall
	for _, p := range mc.Parts {
		switch pt := p.(type) {
		case llms.TextContent:
			if foundText {
				return "", nil, nil, errors.New("expecting a single Text content")
			}
			foundText = true
			text = pt.Text
		case llms.BinaryContent:
			images = append(images, ollamaclient.ImageData(pt.Data))
		case llms.ToolCall:
			tools = append(tools, ollamaclient.ToolCall{
				Function: ollamaclient.ToolCallFunction{
					Name: pt.FunctionCall.Name,
					Arguments: ollamaclient.ToolCallFunctionArguments{
						Content: pt.FunctionCall.Arguments,
					},
				},
			})
		}
	}
	return text, images, tools, nil
}

// makeLLMSToolCall make llms.ToolCall from ollamaclient.ToolCall.
func makeLLMSToolCall(toolCalls []ollamaclient.ToolCall) []llms.ToolCall {
	llmsToolCalls := make([]llms.ToolCall, 0, len(toolCalls))
	for _, tool := range toolCalls {
		llmsToolCalls = append(llmsToolCalls, llms.ToolCall{
			Type: "function",
			FunctionCall: &llms.FunctionCall{
				Name:      tool.Function.Name,
				Arguments: tool.Function.Arguments.Content,
			},
		})
	}
	return llmsToolCalls
}
