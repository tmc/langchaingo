package ollama

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	ollama "github.com/ollama/ollama/api"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("not all input got embedded")
)

// LLM is a ollama LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *ollama.Client
	options          options
}

var _ llms.Model = (*LLM)(nil)

// New creates a new ollama LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := options{ollamaOptions: make(map[string]any)}
	for _, opt := range opts {
		opt(&o)
	}

	var ourl *url.URL
	if o.ollamaServerURL == nil {
		scheme, hostport, ok := strings.Cut(os.Getenv("OLLAMA_HOST"), "://")
		if !ok {
			scheme, hostport = "http", os.Getenv("OLLAMA_HOST")
		}

		host, port, err := net.SplitHostPort(hostport)
		if err != nil {
			host, port = "127.0.0.1", "11434"
			if ip := net.ParseIP(strings.Trim(os.Getenv("OLLAMA_HOST"), "[]")); ip != nil {
				host = ip.String()
			}
		}

		ourl = &url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(host, port),
		}
	}

	var ohttp *http.Client
	if o.httpClient == nil {
		ohttp = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
	} else {
		ohttp = o.httpClient
	}

	return &LLM{
		client:  ollama.NewClient(ourl, ohttp),
		options: o,
	}, nil
}

// Call Implement the call interface for LLM.
func (o *LLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return llms.GenerateFromSinglePrompt(ctx, o, prompt, options...)
}

// GenerateContent implements the Model interface.
// nolint: goerr113
func (o *LLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { // nolint: lll, cyclop, funlen
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
	// We have to convert it to a format Ollama understands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := make([]ollama.Message, 0, len(messages))
	for _, mc := range messages {
		msg := ollama.Message{Role: typeToRole(mc.Role)}

		// Look at all the parts in mc; expect to find a single Text part and
		// any number of binary parts.
		var text string
		foundText := false
		var images []ollama.ImageData

		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				if foundText {
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text
			case llms.BinaryContent:
				images = append(images, ollama.ImageData(pt.Data))
			default:
				return nil, errors.New("only support Text and BinaryContent parts right now")
			}
		}

		msg.Content = text
		msg.Images = images
		chatMsgs = append(chatMsgs, msg)
	}

	format := o.options.format
	if opts.JSONMode {
		format = "json"
	}

	// Get our ollamaOptions from llms.CallOptions
	isStreamed := opts.StreamingFunc != nil
	req := &ollama.ChatRequest{
		Model:    model,
		Format:   format,
		Messages: chatMsgs,
		Options:  makeOllamaOptionsFromOptions(o.options.ollamaOptions, opts),
		Stream:   &isStreamed,
	}

	keepAlive := o.options.keepAlive
	if keepAlive != nil {
		req.KeepAlive = &ollama.Duration{Duration: *o.options.keepAlive}
	}

	var fn ollama.ChatResponseFunc
	streamedResponse := ""
	var resp ollama.ChatResponse

	fn = func(response ollama.ChatResponse) error {
		if opts.StreamingFunc != nil {
			if err := opts.StreamingFunc(ctx, []byte(response.Message.Content)); err != nil {
				return err
			}
		}

		streamedResponse += response.Message.Content

		if req.Stream == nil || !*req.Stream || response.Done {
			resp = response
			resp.Message = ollama.Message{
				Role:    "assistant",
				Content: streamedResponse,
			}
		}
		return nil
	}

	err := o.client.Chat(ctx, req, fn)
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
				"TotalTokens":      resp.EvalCount + resp.PromptEvalCount,
			},
		},
	}

	response := &llms.ContentResponse{Choices: choices}

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func (o *LLM) CreateEmbedding(ctx context.Context, inputTexts []string) ([][]float32, error) {
	var embeddings [][]float32

	req := &ollama.EmbedRequest{
		Input: inputTexts,
		Model: o.options.model,
	}
	if o.options.keepAlive != nil {
		req.KeepAlive = &ollama.Duration{Duration: *o.options.keepAlive}
	}

	embedding, err := o.client.Embed(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(embedding.Embeddings) == 0 {
		return nil, ErrEmptyResponse
	}

	embeddings = embedding.Embeddings

	if len(inputTexts) != len(embeddings) {
		return embeddings, ErrIncompleteEmbedding
	}

	return embeddings, nil
}

func typeToRole(typ llms.ChatMessageType) string {
	switch typ {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeHuman:
		fallthrough
	case llms.ChatMessageTypeGeneric:
		return "user"
	case llms.ChatMessageTypeFunction:
		return "function"
	case llms.ChatMessageTypeTool:
		return "tool"
	}
	return ""
}

func makeOllamaOptionsFromOptions(ollamaOptions map[string]any, opts llms.CallOptions) map[string]any {
	ollamaOptions["num_predict"] = opts.MaxTokens
	ollamaOptions["seed"] = opts.Seed
	ollamaOptions["top_k"] = opts.TopK
	ollamaOptions["top_p"] = opts.TopP
	ollamaOptions["temperature"] = opts.Temperature
	ollamaOptions["repeat_penalty"] = opts.RepetitionPenalty
	ollamaOptions["presence_penalty"] = opts.PresencePenalty
	ollamaOptions["frequency_penalty"] = opts.FrequencyPenalty
	ollamaOptions["stop"] = opts.StopWords

	return ollamaOptions
}
