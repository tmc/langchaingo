package mistral

import (
	"context"
	"errors"
	"os"

	sdk "github.com/gage-technologies/mistral-go"
	"github.com/starmvp/langchaingo/callbacks"
	"github.com/starmvp/langchaingo/llms"
)

// Model encapsulates an instantiated Mistral client, the client options used to instantiate the client, and a callback handler provided by Langchain Go.
type Model struct {
	client           *sdk.MistralClient
	clientOptions    *clientOptions
	CallbacksHandler callbacks.Handler
}

// Assertion to ensure the Mistral `Model` type conforms to the langchaingo llms.Model interface.
var _ llms.Model = (*Model)(nil)

// Instantiates a new Mistral Model.
func New(opts ...Option) (*Model, error) {
	options := &clientOptions{
		apiKey:     os.Getenv("MISTRAL_API_KEY"),
		endpoint:   sdk.Endpoint,
		maxRetries: sdk.DefaultMaxRetries,
		timeout:    sdk.DefaultTimeout,
		model:      sdk.ModelOpenMistral7b,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &Model{
		clientOptions:    options,
		client:           sdk.NewMistralClient(options.apiKey, options.endpoint, options.maxRetries, options.timeout),
		CallbacksHandler: callbacks.SimpleHandler{},
	}, nil
}

// Call implements the langchaingo llms.Model interface.
func (m *Model) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	callOptions := resolveDefaultOptions(sdk.DefaultChatRequestParams, m.clientOptions)
	setCallOptions(options, callOptions)
	mistralChatParams := mistralChatParamsFromCallOptions(callOptions)

	messages := make([]sdk.ChatMessage, 0)
	messages = append(messages, sdk.ChatMessage{
		Role:    "user",
		Content: prompt,
	})
	res, err := m.client.Chat("", messages, &mistralChatParams)
	if err != nil {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return "", err
	}
	if len(res.Choices) != 1 {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return "", errors.New("unexpected response from Mistral SDK, length of the Choices slice must be 1")
	}

	return res.Choices[0].Message.Content, nil
}

// GenerateContent implements the langchaingo llms.Model interface.
func (m *Model) GenerateContent(ctx context.Context, langchainMessages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	callOptions := resolveDefaultOptions(sdk.DefaultChatRequestParams, m.clientOptions)
	setCallOptions(options, callOptions)
	m.CallbacksHandler.HandleLLMGenerateContentStart(ctx, langchainMessages)

	chatOpts := mistralChatParamsFromCallOptions(callOptions)

	messages, err := convertToMistralChatMessages(langchainMessages)
	if err != nil {
		return nil, err
	}

	if callOptions.StreamingFunc != nil {
		return generateStreamingContent(ctx, m, callOptions, messages, chatOpts)
	}
	return generateNonStreamingContent(ctx, m, callOptions, messages, chatOpts)
}

func setCallOptions(options []llms.CallOption, callOpts *llms.CallOptions) {
	for _, opt := range options {
		opt(callOpts)
	}
}

func resolveDefaultOptions(sdkDefaults sdk.ChatRequestParams, c *clientOptions) *llms.CallOptions {
	// Supported models: https://docs.mistral.ai/platform/endpoints/
	// TODO: Mistral also supports ResponseType, which, when set to "json", ensures the model's output is strictly a JSON object.
	// Question: Should `ResponseType` be made a part of llms.CallOptions and pulled whenever Call or GenerateContent is called?
	// The following llms.CallOptions are not supported at the moment by mistral SDK:
	// MinLength, MaxLength,N (how many chat completion choices to generate for each input message), RepetitionPenalty, FrequencyPenalty, and PresencePenalty.
	return &llms.CallOptions{
		Model: c.model,
		// MaxTokens is the maximum number of tokens to generate.
		MaxTokens: sdkDefaults.MaxTokens,
		// Temperature is the temperature for sampling, between 0 and 1.
		Temperature: sdkDefaults.Temperature,
		// TopP is the cumulative probability for top-p sampling.
		TopP: sdkDefaults.TopP,
		// Seed is a seed for deterministic sampling.
		Seed: sdkDefaults.RandomSeed,
		// Function defitions to include in the request.
		Functions: make([]llms.FunctionDefinition, 0),
	}
}

func mistralChatParamsFromCallOptions(callOpts *llms.CallOptions) sdk.ChatRequestParams {
	chatOpts := sdk.DefaultChatRequestParams
	chatOpts.MaxTokens = callOpts.MaxTokens
	chatOpts.Temperature = callOpts.Temperature
	chatOpts.RandomSeed = callOpts.Seed
	chatOpts.Tools = make([]sdk.Tool, 0)
	if len(callOpts.Tools) > 0 {
		for _, tool := range callOpts.Tools {
			chatOpts.Tools = append(chatOpts.Tools, sdk.Tool{
				Type: "function",
				Function: sdk.Function{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			})
		}
	} else {
		for _, function := range callOpts.Functions {
			chatOpts.Tools = append(chatOpts.Tools, sdk.Tool{
				Type: "function",
				Function: sdk.Function{
					Name:        function.Name,
					Description: function.Description,
					Parameters:  function.Parameters,
				},
			})
		}
	}
	return chatOpts
}

func generateNonStreamingContent(ctx context.Context, m *Model, callOptions *llms.CallOptions, messages []sdk.ChatMessage, chatOpts sdk.ChatRequestParams) (*llms.ContentResponse, error) {
	res, err := m.client.Chat(callOptions.Model, messages, &chatOpts)
	m.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, nil)
	if err != nil {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return nil, err
	}

	if len(res.Choices) < 1 {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return nil, errors.New("unexpected response from Mistral SDK, length of the Choices slice must be greater than or equal 1")
	}

	langchainContentResponse := &llms.ContentResponse{
		Choices: make([]*llms.ContentChoice, 0),
	}
	for idx, choice := range res.Choices {
		langchainContentResponse.Choices = append(langchainContentResponse.Choices, &llms.ContentChoice{
			Content:    choice.Message.Content,
			StopReason: string(choice.FinishReason),
			GenerationInfo: map[string]any{
				"created": res.Created,
				"model":   res.Model,
				"usage":   res.Usage,
			},
		})
		toolCalls := choice.Message.ToolCalls
		if len(toolCalls) > 0 {
			langchainContentResponse.Choices[idx].FuncCall = (*llms.FunctionCall)(&toolCalls[0].Function)
			for _, tool := range toolCalls {
				langchainContentResponse.Choices[0].ToolCalls = append(langchainContentResponse.Choices[0].ToolCalls, llms.ToolCall{
					ID:   tool.Id,
					Type: string(tool.Type),
					FunctionCall: &llms.FunctionCall{
						Name:      tool.Function.Name,
						Arguments: tool.Function.Arguments,
					},
				})
			}
		}
	}
	m.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, langchainContentResponse)

	return langchainContentResponse, nil
}

func generateStreamingContent(ctx context.Context, m *Model, callOptions *llms.CallOptions, messages []sdk.ChatMessage, chatOpts sdk.ChatRequestParams) (*llms.ContentResponse, error) {
	chatResChan, err := m.client.ChatStream(callOptions.Model, messages, &chatOpts)
	if err != nil {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return nil, err
	}
	langchainContentResponse := &llms.ContentResponse{
		Choices: make([]*llms.ContentChoice, 1),
	}
	langchainContentResponse.Choices[0] = &llms.ContentChoice{
		Content:        "",
		GenerationInfo: map[string]any{},
	}

	for chatResChunk := range chatResChan {
		chunkStr := ""
		langchainContentResponse.Choices[0].GenerationInfo["created"] = chatResChunk.Created
		langchainContentResponse.Choices[0].GenerationInfo["model"] = chatResChunk.Model
		langchainContentResponse.Choices[0].GenerationInfo["usage"] = chatResChunk.Usage
		if chatResChunk.Error == nil {
			for _, choice := range chatResChunk.Choices {
				chunkStr += choice.Delta.Content
				langchainContentResponse.Choices[0].Content += choice.Delta.Content
				langchainContentResponse.Choices[0].StopReason = string(choice.FinishReason)
				if len(choice.Delta.ToolCalls) > 0 {
					langchainContentResponse.Choices[0].FuncCall = (*llms.FunctionCall)(&choice.Delta.ToolCalls[0].Function)
					for _, tool := range choice.Delta.ToolCalls {
						langchainContentResponse.Choices[0].ToolCalls = append(langchainContentResponse.Choices[0].ToolCalls, llms.ToolCall{
							ID:   tool.Id,
							Type: string(tool.Type),
							FunctionCall: &llms.FunctionCall{
								Name:      tool.Function.Name,
								Arguments: tool.Function.Arguments,
							},
						})
					}
				}
			}
			err := callOptions.StreamingFunc(ctx, []byte(chunkStr))
			if err != nil {
				return langchainContentResponse, err
			}
		} else {
			return langchainContentResponse, chatResChunk.Error
		}
	}

	return langchainContentResponse, nil
}

func convertToMistralChatMessages(langchainMessages []llms.MessageContent) ([]sdk.ChatMessage, error) {
	messages := make([]sdk.ChatMessage, 0)
	for _, msg := range langchainMessages {
		for _, part := range msg.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				chatMsg := sdk.ChatMessage{Content: p.Text, Role: string(msg.Role)}
				setMistralChatMessageRole(&msg, &chatMsg) // #nosec G601
				if chatMsg.Content != "" && chatMsg.Role != "" {
					messages = append(messages, chatMsg)
				}
			case llms.ToolCallResponse:
				chatMsg := sdk.ChatMessage{Role: string(msg.Role), Content: p.Content}
				setMistralChatMessageRole(&msg, &chatMsg) // #nosec G601
				messages = append(messages, chatMsg)
			case llms.ToolCall:
				chatMsg := sdk.ChatMessage{Role: string(msg.Role), ToolCalls: []sdk.ToolCall{{Id: p.ID, Type: sdk.ToolTypeFunction, Function: sdk.FunctionCall{Name: p.FunctionCall.Name, Arguments: p.FunctionCall.Arguments}}}}
				setMistralChatMessageRole(&msg, &chatMsg) // #nosec G601
				messages = append(messages, chatMsg)
			default:
				return nil, errors.New("unsupported content type encountered while preparing chat messages to send to mistral platform")
			}
		}
	}
	return messages, nil
}

func setMistralChatMessageRole(msg *llms.MessageContent, chatMsg *sdk.ChatMessage) {
	switch msg.Role {
	case llms.ChatMessageTypeAI:
		chatMsg.Role = "assistant"
	case llms.ChatMessageTypeGeneric, llms.ChatMessageTypeHuman:
		chatMsg.Role = "user"
	case llms.ChatMessageTypeFunction, llms.ChatMessageTypeTool:
		chatMsg.Role = "tool"
	case llms.ChatMessageTypeSystem:
		chatMsg.Role = "system"
	}
}
