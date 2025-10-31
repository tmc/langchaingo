package yzma

import (
	"context"
	"errors"
	"os"

	"github.com/hybridgroup/yzma/pkg/llama"
	"github.com/tmc/langchaingo/llms"
)

const (
	defaultTemperature = 0.8
	defaultTopK        = 40
	defaultTopP        = 0.9
)

// LLM is a yzma local implementation wrapper to call directly to llama.cpp libs using the FFI interface.
type LLM struct {
	model   string
	options options
}

// New creates a new yzma LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	libPath := os.Getenv("YZMA_LIB")
	if libPath == "" {
		return nil, errors.New("no path to yzma libs")
	}

	if err := llama.Load(""); err != nil {
		return nil, err
	}

	llama.LogSet(llama.LogSilent())
	llama.Init()

	llm := LLM{
		model:   o.model,
		options: o,
	}

	return &llm, nil
}

// Close frees all resources.
func (o *LLM) Close() {
	llama.BackendFree()
}

// Call calls yzma with the given prompt.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	modelName := o.model
	if opts.Model != "" {
		modelName = opts.Model
	}

	maxTokens := int32(1024)
	if opts.MaxTokens > 0 {
		maxTokens = int32(opts.MaxTokens)
	}

	// TODO: allow for setting any passed model params
	model := llama.ModelLoadFromFile(modelName, llama.ModelDefaultParams())
	if model == llama.Model(0) {
		return nil, errors.New("unable to load model")
	}
	defer llama.ModelFree(model)

	// TODO: allow for setting any passed context options
	ctxParams := llama.ContextDefaultParams()
	ctxParams.NCtx = uint32(4096)
	ctxParams.NBatch = uint32(2048)

	lctx := llama.InitFromModel(model, ctxParams)
	if lctx == llama.Context(0) {
		return nil, errors.New("unable to init model")
	}

	defer llama.Free(lctx)

	vocab := llama.ModelGetVocab(model)
	sampler := initSampler(opts)

	msg := chatTemplate(templateForModel(model), convertMessageContent(messages), true)

	// call once to get the size of the tokens from the prompt
	count := llama.Tokenize(vocab, msg, nil, true, true)

	// now get the actual tokens
	tokens := make([]llama.Token, count)
	llama.Tokenize(vocab, msg, tokens, true, true)

	batch := llama.BatchGetOne(tokens)

	if llama.ModelHasEncoder(model) {
		llama.Encode(lctx, batch)

		start := llama.ModelDecoderStartToken(model)
		if start == llama.TokenNull {
			start = llama.VocabBOS(vocab)
		}

		batch = llama.BatchGetOne([]llama.Token{start})
	}

	result := decodeResults(lctx, vocab, batch, sampler, maxTokens)
	choices := []*llms.ContentChoice{
		{
			Content: result,
		},
	}

	response := &llms.ContentResponse{Choices: choices}
	return response, nil
}

func initSampler(opts llms.CallOptions) llama.Sampler {
	temperature := defaultTemperature
	if opts.Temperature > 0 {
		temperature = opts.Temperature
	}
	topK := defaultTopK
	if opts.TopK > 0 {
		topK = opts.TopK
	}

	minP := 0.1

	topP := defaultTopP
	if opts.TopP > 0 {
		topP = opts.TopP
	}

	sampler := llama.SamplerChainInit(llama.SamplerChainDefaultParams())
	llama.SamplerChainAdd(sampler, llama.SamplerInitTopK(int32(topK)))
	llama.SamplerChainAdd(sampler, llama.SamplerInitTopP(float32(topP), 1))
	llama.SamplerChainAdd(sampler, llama.SamplerInitMinP(float32(minP), 1))
	llama.SamplerChainAdd(sampler, llama.SamplerInitTempExt(float32(temperature), 0, 1.0))
	llama.SamplerChainAdd(sampler, llama.SamplerInitDist(llama.DefaultSeed))

	return sampler
}

func templateForModel(model llama.Model) string {
	template := llama.ModelChatTemplate(model, "")
	if template == "" {
		template = "chatml"
	}
	return template
}

func convertMessageContent(msgs []llms.MessageContent) []llama.ChatMessage {
	chatMsgs := []llama.ChatMessage{}
	for _, m := range msgs {
		p := m.Parts[0]
		switch pt := p.(type) {
		case llms.TextContent:
			chatMsgs = append(chatMsgs, llama.NewChatMessage(string(m.Role), pt.Text))
		}
	}
	return chatMsgs
}

func chatTemplate(template string, msgs []llama.ChatMessage, add bool) string {
	buf := make([]byte, 2048)
	len := llama.ChatApplyTemplate(template, msgs, add, buf)
	result := string(buf[:len])
	return result
}

func decodeResults(lctx llama.Context, vocab llama.Vocab, batch llama.Batch, sampler llama.Sampler, maxTokens int32) string {
	result := ""

	for pos := int32(0); pos < maxTokens; pos += batch.NTokens {
		llama.Decode(lctx, batch)
		token := llama.SamplerSample(sampler, lctx, -1)

		if llama.VocabIsEOG(vocab, token) {
			break
		}

		buf := make([]byte, 64)
		len := llama.TokenToPiece(vocab, token, buf, 0, true)

		result = result + string(buf[:len])
		batch = llama.BatchGetOne([]llama.Token{token})
	}

	return result
}
