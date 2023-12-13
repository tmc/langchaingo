package ollama

import (
	"context"
	"fmt"
	"math"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama/internal/ollamaclient"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
)

// LLM is a ollama LLM implementation.
type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *ollamaclient.Client
	options          chatOptions
}

var (
	_ llms.ChatLLM       = (*Chat)(nil)
	_ llms.LanguageModel = (*Chat)(nil)
)

// New creates a new ollama LLM implementation.
func NewChat(opts ...ChatOption) (*Chat, error) {
	o := defaultChatOptions()

	for _, opt := range opts {
		opt(&o)
	}

	client, err := ollamaclient.NewClient(o.ollamaServerURL)
	if err != nil {
		return nil, err
	}

	return &Chat{client: client, options: o}, nil
}

// Call Implement the call interface for LLM.
func (o *Chat) Call(ctx context.Context, messages []schema.ChatMessage, options ...llms.CallOption) (*schema.AIChatMessage, error) { //nolint:lll
	r, err := o.Generate(ctx, [][]schema.ChatMessage{messages}, options...)
	if err != nil {
		return nil, err
	}
	if len(r) == 0 {
		return nil, ErrEmptyResponse
	}
	return r[0].Message, nil
}

// Generate implemente the generate interface for LLM.
func (o *Chat) Generate(ctx context.Context, messageSets [][]schema.ChatMessage, options ...llms.CallOption) ([]*llms.Generation, error) { //nolint:lll,cyclop
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMStart(ctx, o.getPromptsFromMessageSets(messageSets))
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

	// Get our ollamaOptions from llms.CallOptions
	ollamaOptions := makeOllamaOptionsFromOptions(o.options.ollamaOptions, opts)

	generations := make([]*llms.Generation, 0, len(messageSets))
	for _, messages := range messageSets {
		req, err := o.messagesToClientChatMessages(messages)
		if err != nil {
			return nil, err
		}

		req.Model = model
		req.Options = ollamaOptions
		req.Stream = func(b bool) *bool { return &b }(opts.StreamingFunc != nil)

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
			if response.Done {
				resp = response
				resp.Message = &ollamaclient.Message{
					Role:    "assistant",
					Content: streamedResponse,
				}
			}
			return nil
		}

		err = o.client.GenerateChat(ctx, req, fn)
		if err != nil {
			return []*llms.Generation{}, err
		}

		generations = append(generations, makeGenerationFromChatResponse(resp))
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

func makeGenerationFromChatResponse(resp ollamaclient.ChatResponse) *llms.Generation {
	msg := &schema.AIChatMessage{
		Content: resp.Message.Content,
	}

	gen := &llms.Generation{
		Message:        msg,
		Text:           msg.Content,
		GenerationInfo: make(map[string]any),
	}

	gen.GenerationInfo["CompletionTokens"] = resp.EvalCount
	gen.GenerationInfo["PromptTokens"] = resp.PromptEvalCount
	gen.GenerationInfo["TotalTokens"] = resp.PromptEvalCount + resp.EvalCount

	return gen
}

func makeOllamaOptionsFromOptions(ollamaOptions ollamaclient.Options, opts llms.CallOptions) ollamaclient.Options {
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

func (o *Chat) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	embeddings := [][]float32{}

	for _, input := range inputTexts {
		embedding, err := o.client.CreateEmbedding(ctx, &ollamaclient.EmbeddingRequest{
			Prompt: input,
			Model:  o.options.model,
		})
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

func (o *Chat) GeneratePrompt(ctx context.Context, prompts []schema.PromptValue, options ...llms.CallOption) (llms.LLMResult, error) { //nolint:lll
	return llms.GenerateChatPrompt(ctx, o, prompts, options...)
}

func (o *Chat) GetNumTokens(text string) int {
	return llms.CountTokens(o.options.model, text)
}

func (o Chat) getPromptsFromMessageSets(messageSets [][]schema.ChatMessage) []string {
	prompts := make([]string, 0, len(messageSets))
	for _, m := range messageSets {
		r, _ := o.messagesToClientMessages(m)
		prompts = append(prompts, r.Prompt)
	}
	return prompts
}

// convert chatMessage to ollamaclient.GenrateRequest.
func (o Chat) messagesToClientChatMessages(messages []schema.ChatMessage) (*ollamaclient.ChatRequest, error) {
	// Use the template if any
	return messagesToChatRequest(messages)
}

func messagesToChatRequest(messages []schema.ChatMessage) (*ollamaclient.ChatRequest, error) {
	req := &ollamaclient.ChatRequest{}
	for _, m := range messages {
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			fallthrough
		case schema.ChatMessageTypeAI:
			req.Messages = append(req.Messages, &ollamaclient.Message{
				Role:    typeToRole(typ),
				Content: m.GetContent(),
			})
		case schema.ChatMessageTypeHuman:
			fallthrough
		case schema.ChatMessageTypeGeneric:
			req.Messages = append(req.Messages, &ollamaclient.Message{
				Role:    typeToRole(typ),
				Content: m.GetContent(),
			})
		case schema.ChatMessageTypeFunction:
			return nil, fmt.Errorf("chat message type %s not implemented", typ) //nolint:goerr113
		}
	}
	return req, nil
}

func typeToRole(typ schema.ChatMessageType) string {
	switch typ {
	case schema.ChatMessageTypeSystem:
		return "system"
	case schema.ChatMessageTypeAI:
		return "assistant"
	case schema.ChatMessageTypeHuman:
		fallthrough
	case schema.ChatMessageTypeGeneric:
		return "user"
	case schema.ChatMessageTypeFunction:
		return "function"
	}
	return ""
}

// convert chatMessage to ollamaclient.GenrateRequest.
func (o Chat) messagesToClientMessages(messages []schema.ChatMessage) (*ollamaclient.GenerateRequest, error) {
	// Use the template if any
	if o.options.chatTemplate != "" {
		return messagesToClientMessagesWithChatTemlate(o.options.chatTemplate, messages)
	}
	return messagesToGenerateRequestWithoutChatTemplate(messages)
}

func messagesToGenerateRequestWithoutChatTemplate(messages []schema.ChatMessage) (*ollamaclient.GenerateRequest, error) { //nolint:lll
	var prompt string
	req := &ollamaclient.GenerateRequest{}
	for _, m := range messages {
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			req.System = m.GetContent()
		case schema.ChatMessageTypeAI:
			prompt += fmt.Sprintf("%s: %s\n", typ, m.GetContent())
		case schema.ChatMessageTypeHuman:
			fallthrough
		case schema.ChatMessageTypeGeneric:
			if n, ok := m.(schema.Named); ok {
				prompt += fmt.Sprintf("%s: %s: %s\n", typ, n.GetName(), m.GetContent())
			} else {
				prompt += fmt.Sprintf("%s: %s\n", typ, m.GetContent())
			}
		case schema.ChatMessageTypeFunction:
			// not implemented
		}
	}
	req.Prompt = prompt

	return req, nil
}

func messagesToClientMessagesWithChatTemlate(template string, messages []schema.ChatMessage) (*ollamaclient.GenerateRequest, error) { //nolint:lll
	var err error
	req := &ollamaclient.GenerateRequest{}

	p := prompts.PromptTemplate{
		Template:       template,
		InputVariables: []string{"system", "messagesPair"},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	// build or vars
	system := ""
	messagesPair := make([][2]string, int(math.Ceil(float64(len(messages))/2.0))) //nolint:gomnd

	c := 0
	for _, m := range messages {
		typ := m.GetType()
		switch typ {
		case schema.ChatMessageTypeSystem:
			system = m.GetContent()
		case schema.ChatMessageTypeAI:
			messagesPair[c][1] = m.GetContent()
			c++
		case schema.ChatMessageTypeHuman:
			fallthrough
		case schema.ChatMessageTypeGeneric:
			messagesPair[c][0] = m.GetContent()
		case schema.ChatMessageTypeFunction:
			// not implemented
		}
	}
	req.Prompt, err = p.Format(map[string]any{"system": system, "messagesPair": messagesPair})
	if err != nil {
		return nil, err
	}
	req.Template = "{{.Prompt}}"
	return req, nil
}
