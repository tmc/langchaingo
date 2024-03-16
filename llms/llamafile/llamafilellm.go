package llamafile

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/llamafile/internal/llamafileclient"
	"github.com/tmc/langchaingo/schema"
)

var (
	ErrEmptyResponse       = errors.New("no response")
	ErrIncompleteEmbedding = errors.New("no all input got emmbedded")
)

// LLM is a llamafile LLM implementation.
type LLM struct {
	CallbacksHandler callbacks.Handler
	client           *llamafileclient.Client
	options          llamafileclient.GenerationSettings
}

var _ llms.Model = (*LLM)(nil)

// New creates a new llamafile LLM implementation.
func New(opts ...Option) (*LLM, error) {
	o := llamafileclient.GenerationSettings{}
	for _, opt := range opts {
		opt(&o)
	}

	client, err := llamafileclient.NewClient(o.LlamafileServerURL, o.HTTPClient)
	if err != nil {
		return nil, err
	}

	return &LLM{client: client, options: o}, nil
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

	// Our input is a sequence of MessageContent, each of which potentially has
	// a sequence of Part that could be text, images etc.
	// We have to convert it to a format Ollama undestands: ChatRequest, which
	// has a sequence of Message, each of which has a role and content - single
	// text + potential images.
	chatMsgs := make([]*llamafileclient.Message, 0, len(messages))
	for _, mc := range messages {
		msg := &llamafileclient.Message{Role: typeToRole(mc.Role)}

		// Look at all the parts in mc; expect to find a single Text part and
		// any number of binary parts.
		var text string
		foundText := false
		var images []llamafileclient.ImageData

		for _, p := range mc.Parts {
			switch pt := p.(type) {
			case llms.TextContent:
				if foundText {
					return nil, errors.New("expecting a single Text content")
				}
				foundText = true
				text = pt.Text
			case llms.BinaryContent:
				images = append(images, llamafileclient.ImageData(pt.Data))
			default:
				return nil, errors.New("only support Text and BinaryContent parts right now")
			}
		}

		msg.Content = text
		msg.Images = images
		chatMsgs = append(chatMsgs, msg)
	}

	req := &llamafileclient.ChatRequest{
		Messages: chatMsgs,
		Stream:   func(b bool) *bool { return &b }(opts.StreamingFunc != nil),
	}

	req = makeLlamaOptionsFromOptions(req, opts)

	streamedResponse := ""
	fn := func(response llamafileclient.ChatResponse) error {
		if opts.StreamingFunc != nil && response.Content != "" {
			if err := opts.StreamingFunc(ctx, []byte(response.Content)); err != nil {
				return err
			}
		}
		if response.Content != "" {
			streamedResponse += response.Content
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

	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: streamedResponse,
			},
		},
	}, nil
}

func (o *LLM) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := o.client.CreateEmbedding(ctx, texts)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	return resp, nil
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
	return "user"
}

func makeLlamaOptionsFromOptions(input *llamafileclient.ChatRequest, opts llms.CallOptions) *llamafileclient.ChatRequest {
	// Initialize llamaOptions with values from opts
	streamValue := opts.StreamingFunc != nil

	input.FrequencyPenalty = opts.FrequencyPenalty // Assuming FrequencyPenalty correlates to FrequencyPenalty; adjust if necessary
	input.MinP = float64(opts.MinLength)           // Assuming there's a direct correlation; adjust if necessary
	input.Model = opts.Model                       // Assuming Model correlates to Model; adjust if necessary
	input.NCtx = opts.N                            // Assuming N corresponds to NCtx; if not, adjust.
	input.NPredict = opts.MaxTokens                // Assuming MaxTokens correlates to NPredict;
	input.PresencePenalty = opts.PresencePenalty   // Assuming PresencePenalty correlates to PresencePenalty;
	input.RepeatPenalty = opts.RepetitionPenalty   // Assuming RepetitionPenalty correlates to RepeatPenalty;
	input.Seed = uint32(opts.Seed)                 // Convert int to uint32
	input.Stop = opts.StopWords                    // Assuming StopWords correlates to Stop;
	input.Stream = &streamValue                    // True if StreamingFunc provided; adjust logic as needed.
	input.Temperature = opts.Temperature           // Assuming Temperature correlates to Temperature for precision;
	input.TopK = opts.TopK                         // Assuming TopK correlates to TopK;
	input.TopP = opts.TopP                         // Assuming TopP correlates to TopP;

	return input
}
