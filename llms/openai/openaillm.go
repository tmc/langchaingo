// nolint:gci
package openai

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/sashabaranov/go-openai"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

const (
	defaultChatModel       = "gpt-3.5-turbo"
	defaultCompletionModel = "text-davinci-003"
	defaultEmbeddingModel  = "text-embedding-ada-002"

	defaultMaxTokens = 1024
)

var (
	ErrEmptyResponse = errors.New("no response")
	ErrMissingToken  = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")

	ErrUnexpectedEmbeddingModel = errors.New("unexpected embedding model")
)

type LLM struct {
	model  string
	client *openai.Client
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
	opts := llms.CallOptions{MaxTokens: defaultMaxTokens}
	for _, opt := range options {
		opt(&opts)
	}

	model := opts.Model
	if len(model) == 0 {
		model = o.model
	}

	request := openai.CompletionRequest{
		Model:            model,
		MaxTokens:        opts.MaxTokens,
		Temperature:      float32(opts.Temperature),
		TopP:             float32(opts.TopP),
		Stream:           opts.StreamingFunc != nil,
		Stop:             opts.StopWords,
		FrequencyPenalty: float32(opts.RepetitionPenalty),
	}

	generations := make([]*llms.Generation, 0, len(prompts))
	for _, prompt := range prompts {
		request.Prompt = prompt
		if request.Stream {
			generation, err := o.createCompletionStream(ctx, request, opts)
			if err != nil {
				return nil, err
			}
			generations = append(generations, generation)
		} else {
			generation, err := o.createCompletion(ctx, request)
			if err != nil {
				return nil, err
			}
			generations = append(generations, generation)
		}
	}

	return generations, nil
}

func (o *LLM) createCompletionStream(ctx context.Context, request openai.CompletionRequest, opts llms.CallOptions) (*llms.Generation, error) { // nolint:lll
	stream, err := o.client.CreateCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	output := ""
	finishReason := ""
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		if len(resp.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		text := resp.Choices[0].Text
		err = opts.StreamingFunc(ctx, []byte(text))
		if err != nil {
			return nil, err
		}

		output += text
		finishReason = resp.Choices[0].FinishReason
	}

	return &llms.Generation{
		Text: output,
		GenerationInfo: map[string]any{
			"finishReason": finishReason,
		},
	}, nil
}

func (o *LLM) createCompletion(ctx context.Context, request openai.CompletionRequest) (*llms.Generation, error) {
	resp, err := o.client.CreateCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	text := resp.Choices[0].Text
	finishReason := resp.Choices[0].FinishReason
	return &llms.Generation{
		Text: text,
		GenerationInfo: map[string]any{
			"finishReason": finishReason,
		},
	}, nil
}

func (o *LLM) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GeneratePrompt(ctx, o, promptValues, options...)
}

func (o *LLM) GetNumTokens(text string) int {
	return llms.CountTokens(o.model, text)
}

type EmbeddingLLM struct {
	Client *openai.Client
}

// CreateEmbedding creates embeddings for the given input texts.
func (o *EmbeddingLLM) CreateEmbedding(ctx context.Context, model string, inputTexts []string) ([][]float64, error) {
	if len(model) == 0 {
		model = defaultEmbeddingModel
	}

	embeddingModel, ok := stringToEmbeddingModel[model]
	if !ok {
		return nil, ErrUnexpectedEmbeddingModel
	}

	resp, err := o.Client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: embeddingModel,
		Input: inputTexts,
	})
	if err != nil {
		return [][]float64{}, err
	}

	data := resp.Data
	if len(data) == 0 {
		return [][]float64{}, ErrEmptyResponse
	}
	if len(inputTexts) != len(data) {
		return [][]float64{}, ErrUnexpectedResponseLength
	}

	embeddings := make([][]float64, len(data))
	for i := range data {
		embedding := make([]float64, len(data[i].Embedding))
		for j := range data[i].Embedding {
			embedding[j] = float64(data[i].Embedding[j])
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

type Chat struct {
	model  string
	client *openai.Client
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

func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { // nolint:lll,cyclop
	opts := llms.CallOptions{MaxTokens: defaultMaxTokens}
	for _, opt := range options {
		opt(&opts)
	}

	model := opts.Model
	if len(model) == 0 {
		model = o.model
	}

	request := openai.ChatCompletionRequest{
		Model:            model,
		MaxTokens:        opts.MaxTokens,
		Temperature:      float32(opts.Temperature),
		TopP:             float32(opts.TopP),
		Stream:           opts.StreamingFunc != nil,
		Stop:             opts.StopWords,
		FrequencyPenalty: float32(opts.RepetitionPenalty),
	}

	generations := make([]*llms.Generation, 0, len(messageSets))

	openaiMessageSets := make([][]openai.ChatCompletionMessage, len(messageSets))
	for i, messageSet := range messageSets {
		msgs := make([]openai.ChatCompletionMessage, len(messageSet))
		for j, m := range messageSet {
			msg := openai.ChatCompletionMessage{
				Content: m.GetText(),
			}
			typ := m.GetType()
			switch typ {
			case schema.ChatMessageTypeSystem:
				msg.Role = openai.ChatMessageRoleSystem
			case schema.ChatMessageTypeAI:
				msg.Role = openai.ChatMessageRoleAssistant
			case schema.ChatMessageTypeHuman:
				msg.Role = openai.ChatMessageRoleUser
			case schema.ChatMessageTypeGeneric:
				msg.Role = openai.ChatMessageRoleUser
				// TODO: support name
			}
			msgs[j] = msg
		}
		openaiMessageSets[i] = msgs
	}

	for _, msgs := range openaiMessageSets {
		request.Messages = msgs
		if request.Stream {
			generation, err := o.createChatCompletionStream(ctx, request, opts)
			if err != nil {
				return nil, err
			}
			generations = append(generations, generation)
		} else {
			generation, err := o.createChatCompletion(ctx, request)
			if err != nil {
				return nil, err
			}
			generations = append(generations, generation)
		}
	}

	return generations, nil
}

func (o *Chat) createChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest, opts llms.CallOptions) (*llms.Generation, error) { // nolint:lll
	stream, err := o.client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	text := ""
	finishReason := ""
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}

		if len(resp.Choices) == 0 {
			return nil, ErrEmptyResponse
		}

		content := resp.Choices[0].Delta.Content
		err = opts.StreamingFunc(ctx, []byte(content))
		if err != nil {
			return nil, err
		}

		text += content
		finishReason = string(resp.Choices[0].FinishReason)
	}

	return &llms.Generation{
		Message: &schema.AIChatMessage{
			Text: text,
		},
		GenerationInfo: map[string]any{
			"finishReason": finishReason,
		},
	}, nil
}

func (o *Chat) createChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (*llms.Generation, error) { // nolint:lll
	resp, err := o.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	text := resp.Choices[0].Message.Content
	finishReason := string(resp.Choices[0].FinishReason)
	return &llms.Generation{
		Message: &schema.AIChatMessage{
			Text: text,
		},
		GenerationInfo: map[string]any{
			"finishReason": finishReason,
		},
	}, nil
}

func (o *Chat) GetNumTokens(text string) int {
	return llms.CountTokens(o.model, text)
}

func (o *Chat) GeneratePrompt(ctx context.Context, promptValues []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, promptValues, options...)
}

// New returns a new OpenAI LLM.
func New(opts ...Option) (*LLM, error) {
	options := &options{
		token:   os.Getenv(tokenEnvVarName),
		model:   os.Getenv(modelEnvVarName),
		baseURL: os.Getenv(baseURLEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.model) == 0 {
		options.model = defaultCompletionModel
	}

	c, err := newClient(options)
	return &LLM{
		model:  options.model,
		client: c,
	}, err
}

// NewChat returns a new OpenAI chat LLM.
func NewChat(opts ...Option) (*Chat, error) {
	options := &options{
		token:   os.Getenv(tokenEnvVarName),
		model:   os.Getenv(modelEnvVarName),
		baseURL: os.Getenv(baseURLEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.model) == 0 {
		options.model = defaultChatModel
	}

	c, err := newClient(options)
	return &Chat{
		model:  options.model,
		client: c,
	}, err
}

// NewEmbedding returns a new OpenAI embedding LLM.
func NewEmbedding(opts ...Option) (*EmbeddingLLM, error) {
	options := &options{
		token:   os.Getenv(tokenEnvVarName),
		baseURL: os.Getenv(baseURLEnvVarName),
	}

	for _, opt := range opts {
		opt(options)
	}

	c, err := newClient(options)
	if err != nil {
		return nil, err
	}
	return &EmbeddingLLM{
		Client: c,
	}, nil
}

func newClient(options *options) (*openai.Client, error) {
	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}
	config := openai.DefaultConfig(options.token)
	config.BaseURL = options.baseURL
	client := openai.NewClientWithConfig(config)
	return client, nil
}
