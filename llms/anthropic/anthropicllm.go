package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
)

var (
	ErrEmptyResponse            = errors.New("no response")
	ErrMissingToken             = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")
	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
	ErrInvalidContentType       = errors.New("invalid content type")
	ErrUnsupportedMessageType   = errors.New("unsupported message type")
	ErrUnsupportedContentType   = errors.New("unsupported content type")
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *anthropicclient.Client
}

var _ llms.Model = (*LLM)(nil)

// New returns a new Anthropic LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to create client: %w", err)
	}
	return &LLM{
		client: c,
	}, nil
}

func newClient(opts ...Option) (*anthropicclient.Client, error) {
	options := &options{
		token:      os.Getenv(tokenEnvVarName),
		baseURL:    anthropicclient.DefaultBaseURL,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return anthropicclient.New(options.token, options.model, options.baseURL,
		anthropicclient.WithHTTPClient(options.httpClient),
		anthropicclient.WithLegacyTextCompletionsAPI(options.useLegacyTextCompletionsAPI),
		anthropicclient.WithAnthropicBetaHeader(options.anthropicBetaHeader),
	)
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if o.client.UseLegacyTextCompletionsAPI {
		return generateCompletionsContent(ctx, o, messages, opts)
	}
	return generateMessagesContent(ctx, o, messages, opts)
}

func generateCompletionsContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	if len(messages) == 0 || len(messages[0].Parts) == 0 {
		return nil, ErrEmptyResponse
	}

	msg0 := messages[0]
	part := msg0.Parts[0]
	partText, ok := part.(llms.TextContent)
	if !ok {
		return nil, fmt.Errorf("anthropic: unexpected message type: %T", part)
	}
	prompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", partText.Text)
	result, err := o.client.CreateCompletion(ctx, &anthropicclient.CompletionRequest{
		Model:     opts.Model,
		Prompt:    prompt,
		MaxTokens: opts.MaxTokens,
		StopWords: func() []string {
			if len(opts.StopSequences) > 0 {
				return opts.StopSequences
			}
			return opts.StopWords
		}(),
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("anthropic: failed to create completion: %w", err)
	}

	resp := &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: result.Text,
			},
		},
	}
	return resp, nil
}

func generateMessagesContent(ctx context.Context, o *LLM, messages []llms.MessageContent, opts *llms.CallOptions) (*llms.ContentResponse, error) {
	// Process messages and handle errors
	chatMessages, systemPrompt, err := processMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to process messages: %w", err)
	}

	// Create message and handle errors
	result, err := createMessage(ctx, o, chatMessages, systemPrompt, opts)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, ErrEmptyResponse
	}

	// Process content choices
	choices := make([]*llms.ContentChoice, len(result.Content))
	for i, content := range result.Content {
		choice, err := processContent(content, result)
		if err != nil {
			return nil, fmt.Errorf("anthropic: failed to process content: %w", err)
		}
		choices[i] = choice
	}

	return &llms.ContentResponse{
		Choices: choices,
	}, nil
}

// Helper function to create message
func createMessage(ctx context.Context, o *LLM, chatMessages []*anthropicclient.ChatMessage, systemPrompt string, opts *llms.CallOptions) (*anthropicclient.MessageResponsePayload, error) {
	tools := toolsToTools(opts.Tools)
	messages := make([]anthropicclient.ChatMessage, len(chatMessages))
	for i, msg := range chatMessages {
		messages[i] = *msg
	}
	result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
		Model:     opts.Model,
		Messages:  messages,
		System:    systemPrompt,
		MaxTokens: opts.MaxTokens,
		StopWords: func() []string {
			if len(opts.StopSequences) > 0 {
				return opts.StopSequences
			}
			return opts.StopWords
		}(),
		Temperature:   opts.Temperature,
		TopP:          opts.TopP,
		Tools:         tools,
		StreamingFunc: opts.StreamingFunc,
	})
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, fmt.Errorf("anthropic: failed to create message: %w", err)
	}
	return result, nil
}

// Helper function to process content
func processContent(content anthropicclient.Content, result *anthropicclient.MessageResponsePayload) (*llms.ContentChoice, error) {
	switch content.GetType() {
	case "text":
		return processTextContent(content, result)
	case "tool_use":
		return processToolUseContent(content, result)
	default:
		return nil, fmt.Errorf("anthropic: %w: %v", ErrUnsupportedContentType, content.GetType())
	}
}

// Helper function to process text content
func processTextContent(content anthropicclient.Content, result *anthropicclient.MessageResponsePayload) (*llms.ContentChoice, error) {
	textContent, ok := content.(*anthropicclient.TextContent)
	if !ok {
		return nil, fmt.Errorf("anthropic: %w for text message", ErrInvalidContentType)
	}
	return &llms.ContentChoice{
		Content:    textContent.Text,
		StopReason: result.StopReason,
		GenerationInfo: map[string]any{
			"InputTokens":  result.Usage.InputTokens,
			"OutputTokens": result.Usage.OutputTokens,
		},
	}, nil
}

// Helper function to process tool use content
func processToolUseContent(content anthropicclient.Content, result *anthropicclient.MessageResponsePayload) (*llms.ContentChoice, error) {
	toolUseContent, ok := content.(*anthropicclient.ToolUseContent)
	if !ok {
		return nil, fmt.Errorf("anthropic: %w for tool use message", ErrInvalidContentType)
	}
	argumentsJSON, err := json.Marshal(toolUseContent.Input)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to marshal tool use arguments: %w", err)
	}
	return &llms.ContentChoice{
		ToolCalls: []llms.ToolCall{
			{
				ID: toolUseContent.ID,
				FunctionCall: &llms.FunctionCall{
					Name:      toolUseContent.Name,
					Arguments: string(argumentsJSON),
				},
			},
		},
		StopReason: result.StopReason,
		GenerationInfo: map[string]any{
			"InputTokens":  result.Usage.InputTokens,
			"OutputTokens": result.Usage.OutputTokens,
		},
	}, nil
}

func toolsToTools(tools []llms.Tool) []anthropicclient.Tool {
	toolReq := make([]anthropicclient.Tool, len(tools))
	for i, tool := range tools {
		toolReq[i] = anthropicclient.Tool{
			Name:        tool.Function.Name,
			Description: tool.Function.Description,
			InputSchema: tool.Function.Parameters,
		}
	}
	return toolReq
}

func processMessages(messages []llms.MessageContent) ([]*anthropicclient.ChatMessage, string, error) {
	chatMessages := make([]*anthropicclient.ChatMessage, 0, len(messages))
	systemPrompt := ""
	for _, msg := range messages {
		switch msg.Role {
		case llms.ChatMessageTypeSystem:
			content, err := handleSystemMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle system message: %w", err)
			}
			systemPrompt += content
		case llms.ChatMessageTypeHuman:
			chatMessage, err := handleHumanMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle human message: %w", err)
			}
			chatMessages = append(chatMessages, &chatMessage)
		case llms.ChatMessageTypeAI:
			chatMessage, err := handleAIMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle AI message: %w", err)
			}
			chatMessages = append(chatMessages, &chatMessage)
		case llms.ChatMessageTypeTool:
			chatMessage, err := handleToolMessage(msg)
			if err != nil {
				return nil, "", fmt.Errorf("anthropic: failed to handle tool message: %w", err)
			}
			chatMessages = append(chatMessages, &chatMessage)
		case llms.ChatMessageTypeGeneric, llms.ChatMessageTypeFunction:
			return nil, "", fmt.Errorf("anthropic: %w: %v", ErrUnsupportedMessageType, msg.Role)
		default:
			return nil, "", fmt.Errorf("anthropic: %w: %v", ErrUnsupportedMessageType, msg.Role)
		}
	}
	return chatMessages, systemPrompt, nil
}

func handleSystemMessage(msg llms.MessageContent) (string, error) {
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return textContent.Text, nil
	}
	return "", fmt.Errorf("anthropic: %w for system message", ErrInvalidContentType)
}

func handleHumanMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return anthropicclient.ChatMessage{
			Role:    RoleUser,
			Content: textContent.Text,
		}, nil
	}
	return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for human message", ErrInvalidContentType)
}

func handleAIMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if toolCall, ok := msg.Parts[0].(llms.ToolCall); ok {
		var inputStruct map[string]interface{}
		err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &inputStruct)
		if err != nil {
			return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: failed to unmarshal tool call arguments: %w", err)
		}
		toolUse := anthropicclient.ToolUseContent{
			Type:  "tool_use",
			ID:    toolCall.ID,
			Name:  toolCall.FunctionCall.Name,
			Input: inputStruct,
		}

		return anthropicclient.ChatMessage{
			Role:    RoleAssistant,
			Content: []anthropicclient.Content{toolUse},
		}, nil
	}
	if textContent, ok := msg.Parts[0].(llms.TextContent); ok {
		return anthropicclient.ChatMessage{
			Role: RoleAssistant,
			Content: []anthropicclient.Content{&anthropicclient.TextContent{
				Type: "text",
				Text: textContent.Text,
			}},
		}, nil
	}
	return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for AI message", ErrInvalidContentType)
}

type ToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

func handleToolMessage(msg llms.MessageContent) (anthropicclient.ChatMessage, error) {
	if toolCallResponse, ok := msg.Parts[0].(llms.ToolCallResponse); ok {
		toolContent := anthropicclient.ToolResultContent{
			Type:      "tool_result",
			ToolUseID: toolCallResponse.ToolCallID,
			Content:   toolCallResponse.Content,
		}

		return anthropicclient.ChatMessage{
			Role:    RoleUser,
			Content: []anthropicclient.Content{toolContent},
		}, nil
	}
	return anthropicclient.ChatMessage{}, fmt.Errorf("anthropic: %w for tool message", ErrInvalidContentType)
}
