package watsonx

import (
	"context"
	"errors"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"

	wx "github.com/IBM/watsonx-go/pkg/models"
)

var (
	ErrInvalidPrompt = errors.New("invalid prompt")
	ErrEmptyResponse = errors.New("no response")
)

type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *wx.Client

	modelID string
}

var _ llms.Model = (*LLM)(nil)

// Call implements the LLM interface.
func (wx *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, wx, prompt, options...)
}

// GenerateContent implements the Model interface.
func (wx *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, whitespace

	if wx.CallbacksHandler != nil {
		wx.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	prompt, err := getPrompt(messages)
	if err != nil {
		return nil, err
	}

	result, err := wx.client.GenerateText(
		wx.modelID,
		prompt,
		toWatsonxOptions(&options)...,
	)
	if err != nil {
		if wx.CallbacksHandler != nil {
			wx.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	if result.Text == "" {
		return nil, ErrEmptyResponse
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

func New(modelID string, opts ...wx.ClientOption) (*LLM, error) {
	c, err := wx.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &LLM{
		client:  c,
		modelID: modelID,
	}, nil
}

func getPrompt(messages []llms.MessageContent) (string, error) {
	// Assume we get a single text message
	msg0 := messages[0]
	part := msg0.Parts[0]
	prompt, ok := part.(llms.TextContent)
	if !ok {
		return "", ErrInvalidPrompt
	}

	return prompt.Text, nil
}

func getDefaultCallOptions() *llms.CallOptions {
	return &llms.CallOptions{
		TopP:              getFloatPointer(-1),
		TopK:              getIntPointer(-1),
		Temperature:       getFloatPointer(-1),
		Seed:              getIntPointer(-1),
		RepetitionPenalty: getFloatPointer(-1),
		MaxTokens:         getIntPointer(-1),
	}
}

func toWatsonxOptions(options *[]llms.CallOption) []wx.GenerateOption {
	opts := getDefaultCallOptions()
	for _, opt := range *options {
		opt(opts)
	}

	o := []wx.GenerateOption{}
	if topP := opts.GetTopP(); topP != -1 {
		o = append(o, wx.WithTopP(topP))
	}
	if topK := opts.GetTopK(); topK != -1 {
		o = append(o, wx.WithTopK(uint(topK)))
	}
	if temperature := opts.GetTemperature(); temperature != -1 {
		o = append(o, wx.WithTemperature(temperature))
	}
	if seed := opts.GetSeed(); seed != -1 {
		o = append(o, wx.WithRandomSeed(uint(seed)))
	}
	if repetitionPenalty := opts.GetRepetitionPenalty(); repetitionPenalty != -1 {
		o = append(o, wx.WithRepetitionPenalty(repetitionPenalty))
	}
	if maxTokens := opts.GetMaxTokens(); maxTokens != -1 {
		o = append(o, wx.WithMaxNewTokens(uint(maxTokens)))
	}
	if len(opts.StopWords) > 0 {
		o = append(o, wx.WithStopSequences(opts.StopWords))
	}

	/*
	   watsonx options not supported:

	   	wx.WithMinNewTokens(minNewTokens)
	   	wx.WithDecodingMethod(decodingMethod)
	   	wx.WithLengthPenalty(decayFactor, startIndex)
	   	wx.WithTimeLimit(timeLimit)
	   	wx.WithTruncateInputTokens(truncateInputTokens)
	   	wx.WithReturnOptions(inputText, generatedTokens, inputTokens, tokenLogProbs, tokenRanks, topNTokens)
	*/

	return o
}

func getFloatPointer(f float64) *float64 {
	return &f
}

func getIntPointer(i int) *int {
	return &i
}
