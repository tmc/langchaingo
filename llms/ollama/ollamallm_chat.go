package ollama

import (
	"context"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama/internal/ollamaclient"
	"github.com/tmc/langchaingo/schema"
)

// LLM is a ollama LLM implementation.
type Chat struct {
	CallbacksHandler callbacks.Handler
	client           *ollamaclient.Client
	options          options
}

var _ llms.ChatLLM = (*Chat)(nil)

// New creates a new ollama LLM implementation.
func NewChat(opts ...Option) (*Chat, error) {
	o := options{}
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
		req, err := messagesToChatRequest(messages)
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
			if o.CallbacksHandler != nil {
				o.CallbacksHandler.HandleLLMError(ctx, err)
			}
			return []*llms.Generation{}, err
		}

		generations = append(generations, makeGenerationFromChatResponse(resp))
	}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMEnd(ctx, llms.LLMResult{Generations: [][]*llms.Generation{generations}})
	}

	return generations, nil
}

// GenerateContent implements the Model interface.
// nolint: goerr113
func (o *Chat) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen
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
	// We have to convert it to a format Ollama undestands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := make([]*ollamaclient.Message, 0, len(messages))
	for _, mc := range messages {
		msg := &ollamaclient.Message{Role: typeToRole(mc.Role)}

		// Look at all the parts in mc; expect to find a single Text part and
		// any number of binary parts.
		var text string
		foundText := false
		var images []ollamaclient.ImageData

		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				if foundText {
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text
			case llms.BinaryContent:
				images = append(images, ollamaclient.ImageData(pt.Data))
			default:
				return nil, errors.New("only support Text and BinaryContent parts right now")
			}
		}

		msg.Content = text
		msg.Images = images
		chatMsgs = append(chatMsgs, msg)
	}

	// Get our ollamaOptions from llms.CallOptions
	ollamaOptions := makeOllamaOptionsFromOptions(o.options.ollamaOptions, opts)
	req := &ollamaclient.ChatRequest{
		Model:    model,
		Messages: chatMsgs,
		Options:  ollamaOptions,
		Stream:   func(b bool) *bool { return &b }(opts.StreamingFunc != nil),
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
		if response.Done {
			resp = response
			resp.Message = &ollamaclient.Message{
				Role:    "assistant",
				Content: streamedResponse,
			}
		}
		return nil
	}

	err := o.client.GenerateChat(ctx, req, fn)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	choices := []*llms.ContentChoice{
		{
			Content: resp.Message.Content,
			GenerationInfo: map[string]any{
				"CompletionTokens": resp.EvalCount,
				"PromptTokens":     resp.PromptEvalCount,
				"TotalTokesn":      resp.EvalCount + resp.PromptEvalCount,
			},
		},
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
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

func (o Chat) getPromptsFromMessageSets(messageSets [][]schema.ChatMessage) []string {
	prompts := make([]string, 0, len(messageSets))
	for i := 0; i < len(messageSets); i++ {
		curPrompt := ""
		for j := 0; j < len(messageSets[i]); j++ {
			curPrompt += messageSets[i][j].GetContent()
		}
		prompts = append(prompts, curPrompt)
	}
	return prompts
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
