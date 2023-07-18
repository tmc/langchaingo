package openai

import (
	"context"
	"errors"
	"os"
	"reflect"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai/internal/openaiclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type ChatMessage = openaiclient.ChatMessage

type LLM struct {
	client *openaiclient.Client
}

var (
	_ llms.LLM           = (*LLM)(nil)
	_ llms.LanguageModel = (*LLM)(nil)
)

// Call requests a completion for the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	r, err := o.Generate(ctx, []string{prompt}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Text, nil
}

func (o *LLM) Generate(ctx context.Context, prompts []string, options ...llms.CallOption) ([]*llms.Generation, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		result, err := o.client.CreateCompletion(ctx, &openaiclient.CompletionRequest{
			Model:            opts.Model,
			Prompt:           prompt,
			MaxTokens:        opts.MaxTokens,
			StopWords:        opts.StopWords,
			Temperature:      opts.Temperature,
			N:                opts.N,
			FrequencyPenalty: opts.FrequencyPenalty,
			PresencePenalty:  opts.PresencePenalty,
			TopP:             opts.TopP,
		})
		if err != nil {
			return nil, err
		}
		generations = append(generations, &llms.Generation{
			Text: result.Text,
		})
	}

	return generations, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens(o.client.Model, text)
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float64, error) {
	embeddings, err := o.client.CreateEmbedding(ctx, &openaiclient.EmbeddingRequest{
		Input: inputTexts,
	})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmptyResponse
	}
	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrUnexpectedResponseLength
	}
	return embeddings, nil
}

type Chat struct {
	client *openaiclient.Client
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// Chat requests a chat response for the given messages.
func (o *Chat) Call(ctx context.Context, messages []schema.ChatMessage, options ...llms.CallOption) (string, error) { // nolint: lll
	r, err := o.Generate(ctx, [][]schema.ChatMessage{messages}, options...)
	if err != nil {
		return "", err
	}
	if len(r) == 0 {
		return "", ErrEmptyResponse
	}
	return r[0].Message.Text, nil
}

func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint:lll
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messageSet := range messageSets {
		msgs := make([]*openaiclient.ChatMessage, len(messageSet))
		for i, m := range messageSet {
			msg := &openaiclient.ChatMessage{
				Content: m.GetText(),
			}
			typ := m.GetType()
			switch typ {
			case schema.ChatMessageTypeSystem:
				msg.Role = "system"
			case schema.ChatMessageTypeAI:
				msg.Role = "assistant"
			case schema.ChatMessageTypeHuman:
				msg.Role = "user"
			case schema.ChatMessageTypeGeneric:
				msg.Role = "user"
				// TODO: support name
			}
			msgs[i] = msg
		}

		result, err := o.client.CreateChat(ctx, &openaiclient.ChatRequest{
			Model:            opts.Model,
			StopWords:        opts.StopWords,
			Messages:         msgs,
			StreamingFunc:    opts.StreamingFunc,
			Temperature:      opts.Temperature,
			MaxTokens:        opts.MaxTokens,
			N:                opts.N,
			FrequencyPenalty: opts.FrequencyPenalty,
			PresencePenalty:  opts.PresencePenalty,
		})
		if err != nil {
			return nil, err
		}
		if len(result.Choices) == 0 {
			return nil, ErrEmptyResponse
		}
		text := result.Choices[0].Message.Content
		generationInfo := make(map[string]any, reflect.ValueOf(result.Usage).NumField())
		generationInfo["CompletionTokens"] = result.Usage.CompletionTokens
		generationInfo["PromptTokens"] = result.Usage.PromptTokens
		generationInfo["TotalTokens"] = result.Usage.TotalTokens
		generations = append(generations, &llms.Generation{
			Message: &schema.AIChatMessage{
				Text: text,
			},
			Text:           text,
			GenerationInfo: generationInfo,
		})
	}

	return generations, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return llms.CountTokens(o.client.Model, text)
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	c, err := newClient(opts...)
	return &LLM{
		client: c,
	}, err
}

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	c, err := newClient(opts...)
	return &Chat{
		client: c,
	}, err
}

func newClient(opts ...Option) (*openaiclient.Client, error) {
	options := &options{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		baseURL:      os.Getenv(baseURLEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		apiType:      APITypeOpenAI,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return openaiclient.New(options.token, options.model, options.baseURL, options.organization,
		openaiclient.APIType(options.apiType), "")
}

// NewAzure returns a new Azure OpenAI LLM .
func NewAzure(opts ...Option) (*LLM, error) {
	c, err := newAzureClient(opts...)
	return &LLM{
		client: c,
	}, err
}

// NewAzureChat returns a new OpenAI chat LLM.
func NewAzureChat(opts ...Option) (*Chat, error) {
	c, err := newAzureClient(opts...)
	return &Chat{
		client: c,
	}, err
}

func newAzureClient(opts ...Option) (*openaiclient.Client, error) {
	options := &options{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		baseURL:      os.Getenv(baseURLEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		apiType:      APITypeAzure,
		apiVersion:   DefaultApiVersion,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	return openaiclient.New(options.token, options.model, options.baseURL,
		options.organization, openaiclient.APIType(options.apiType), options.apiVersion)
}
