package mistral

import (
	"context"
	"errors"

	sdk "github.com/gage-technologies/mistral-go"
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/mistral/internal/mistralclient"
	"github.com/tmc/langchaingo/schema"
)

// LLM is a mistral LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *sdk.MistralClient
}

// Call implements llms.Model.
func (m *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	callOptions := resolveDefaultsFromSdk(sdk.DefaultChatRequestParams)
	setCallOptions(options, callOptions)
	chatOpts := chatOptsFromCallOpts(callOptions)

	messages := make([]sdk.ChatMessage, 0)
	messages = append(messages, sdk.ChatMessage{
		Role:    "user",
		Content: prompt,
	})
	res, err := m.client.Chat("", messages, &chatOpts)
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

func chatOptsFromCallOpts(callOpts *llms.CallOptions) sdk.ChatRequestParams {
	chatOpts := sdk.DefaultChatRequestParams
	chatOpts.MaxTokens = callOpts.MaxTokens
	chatOpts.Temperature = callOpts.Temperature
	chatOpts.Tools = make([]sdk.Tool, 0)
	for _, function := range callOpts.Functions {
		chatOpts.Tools = append(chatOpts.Tools, sdk.Tool{
			Type: "function",
			Function: sdk.Function{
				Name:        function.Name,
				Description: function.Description,
				Parameters:  function.Parameters,
			}})
	}
	return chatOpts
}

// Supported models: https://docs.mistral.ai/platform/endpoints/
// TODO: Mistral also supports ResponseType, which, when set to "json", ensures the model's output is strictly a JSON object. Should this be made a part of client config and pulled whenever Call or GenerateContent is called?
// The following llms.CallOptions are not supported at the moment by mistral SDK:
// MinLength, MaxLength,N (how many chat completion choices to generate for each input message), RepetitionPenalty, FrequencyPenalty, and PresencePenalty.
func resolveDefaultsFromSdk(sdkDefaults sdk.ChatRequestParams) *llms.CallOptions {
	return &llms.CallOptions{
		Model: "open-mistral-7b",
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

func setCallOptions(options []llms.CallOption, callOpts *llms.CallOptions) {
	for _, opt := range options {
		opt(callOpts)
	}
}

// GenerateContent implements llms.Model.
func (m *LLM) GenerateContent(ctx context.Context, langchainMessages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	callOptions := resolveDefaultsFromSdk(sdk.DefaultChatRequestParams)
	setCallOptions(options, callOptions)

	chatOpts := chatOptsFromCallOpts(callOptions)

	messages, err := convertToMistralChatMessages(langchainMessages)
	if err != nil {
		return nil, err
	}

	m.CallbacksHandler.HandleLLMGenerateContentStart(ctx, langchainMessages)
	res, err := m.client.Chat(callOptions.Model, messages, &chatOpts)
	m.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, nil) // TODO: fix
	if err != nil {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return nil, err
	}

	if len(res.Choices) < 1 {
		m.CallbacksHandler.HandleLLMError(ctx, err)
		return nil, errors.New("unexpected response from Mistral SDK, length of the Choices slice must be greater than or equal 1")
	}

	langchainContentResponse := &llms.ContentResponse{
		Choices: make([]*llms.ContentChoice, len(res.Choices)),
	}
	for idx, choice := range res.Choices {
		langchainContentResponse.Choices[idx] = &llms.ContentChoice{
			Content:    choice.Message.Content,
			StopReason: string(choice.FinishReason),
			GenerationInfo: map[string]any{
				"created": res.Created,
				"model":   res.Model,
				"usage":   res.Usage,
			},
		}
	}
	m.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, langchainContentResponse)

	// // Example: Using Chat Completions Stream
	// chatResChan, err := m.client.ChatStream("mistral-tiny", []mistral.ChatMessage{{Content: "Hello, world!", Role: mistral.RoleUser}}, nil)
	// if err != nil {
	// 	log.Fatalf("Error getting chat completion stream: %v", err)
	// }

	// for chatResChunk := range chatResChan {
	// 	if chatResChunk.Error != nil {
	// 		log.Fatalf("Error while streaming response: %v", chatResChunk.Error)
	// 	}
	// 	log.Printf("Chat completion stream part: %+v\n", chatResChunk)
	// }

	return langchainContentResponse, nil
}

func convertToMistralChatMessages(langchainMessages []llms.MessageContent) ([]sdk.ChatMessage, error) {
	messages := make([]sdk.ChatMessage, len(langchainMessages))
	for _, msg := range langchainMessages {
		msgText := ""
		for _, part := range msg.Parts {
			textContent, ok := part.(llms.TextContent)
			if !ok {
				return nil, errors.New("unsupported content type encountered while preparing chat messages to send to mistral platform")
			}
			msgText = msgText + textContent.Text
		}
		chatMsg := sdk.ChatMessage{Content: msgText, Role: "user"}

		setRole(&msg, &chatMsg)
		messages = append(messages, chatMsg)
	}
	return messages, nil
}

func setRole(msg *llms.MessageContent, chatMsg *sdk.ChatMessage) {
	if msg.Role == schema.ChatMessageTypeAI {
		chatMsg.Role = "assistant"
	} else if msg.Role == schema.ChatMessageTypeGeneric || msg.Role == schema.ChatMessageTypeHuman {
		chatMsg.Role = "user"
	} else if msg.Role == schema.ChatMessageTypeFunction {
		chatMsg.Role = "tool"
	} else if msg.Role == schema.ChatMessageTypeSystem {
		chatMsg.Role = "system"
	}
}

var _ llms.Model = (*LLM)(nil)

// New creates a new mistral LLM implementation.
func New(opts ...mistralclient.Option) *LLM {
	return &LLM{client: mistralclient.NewClient(opts...), CallbacksHandler: callbacks.SimpleHandler{}}
}
