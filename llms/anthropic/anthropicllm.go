package anthropic

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic/internal/anthropicclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the Anthropic API key, set it in the ANTHROPIC_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
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
	return &LLM{
		client: c,
	}, err
}

func newClient(opts ...Option) (*anthropicclient.Client, error) {
	options := &options{
		token:             os.Getenv(tokenEnvVarName),
		baseURL:           anthropicclient.DefaultBaseURL,
		httpClient:        http.DefaultClient,
		useCompletionsAPI: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return anthropicclient.New(options.token, options.model, options.baseURL, options.httpClient)
}

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := &llms.CallOptions{}
	for _, opt := range options {
		opt(opts)
	}

	if o.client.UseCompletionsAPI {
		// Assume we get a single text message
		msg0 := messages[0]
		part := msg0.Parts[0]
		partText, ok := part.(llms.TextContent)
		if !ok {
			return nil, fmt.Errorf("unexpected message type: %T", part)
		}
		prompt := fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", partText.Text)
		result, err := o.client.CreateCompletion(ctx, &anthropicclient.CompletionRequest{
			Model:         opts.Model,
			Prompt:        prompt,
			MaxTokens:     opts.MaxTokens,
			StopWords:     opts.StopWords,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			StreamingFunc: opts.StreamingFunc,
		})
		if err != nil {
			if o.CallbacksHandler != nil {
				o.CallbacksHandler.HandleLLMError(ctx, err)
			}
			return nil, err
		}

		resp := &llms.ContentResponse{
			Choices: []*llms.ContentChoice{
				{
					Content: result.Text,
				},
			},
		}
		return resp, nil
	} else {
		chatMessages := make([]anthropicclient.ChatMessage, 0, len(messages))
		systemPrompt := ""
		for _, msg := range messages {
			switch msg.Role {
			case schema.ChatMessageTypeSystem:
				systemPrompt += msg.Parts[0].(llms.TextContent).Text
			case schema.ChatMessageTypeHuman:
				chatMessages = append(chatMessages, anthropicclient.ChatMessage{
					Role:    RoleUser,
					Content: msg.Parts[0].(llms.TextContent).Text,
				})
			case schema.ChatMessageTypeAI:
				chatMessages = append(chatMessages, anthropicclient.ChatMessage{
					Role:    RoleAssistant,
					Content: msg.Parts[0].(llms.TextContent).Text,
				})
			default:
				return nil, fmt.Errorf("role %v not supported", msg.Role)
			}
		}

		result, err := o.client.CreateMessage(ctx, &anthropicclient.MessageRequest{
			Model:         opts.Model,
			Messages:      chatMessages,
			System:        systemPrompt,
			MaxTokens:     opts.MaxTokens,
			StopWords:     opts.StopWords,
			Temperature:   opts.Temperature,
			TopP:          opts.TopP,
			StreamingFunc: opts.StreamingFunc,
		})
		if err != nil {
			if o.CallbacksHandler != nil {
				o.CallbacksHandler.HandleLLMError(ctx, err)
			}
			return nil, err
		}

		choices := make([]*llms.ContentChoice, len(result.Content))
		for i, content := range result.Content {
			choices[i] = &llms.ContentChoice{
				Content:    content.Text,
				StopReason: result.StopReason,
				GenerationInfo: map[string]any{
					"InputTokens":  result.Usage.InputTokens,
					"OutputTokens": result.Usage.OutputTokens,
				},
			}
		}

		resp := &llms.ContentResponse{
			Choices: choices,
		}
		return resp, nil
	}
}
