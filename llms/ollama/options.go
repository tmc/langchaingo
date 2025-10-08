package ollama

import (
	"log"
	"net/http"
	"net/url"

	"github.com/tmc/langchaingo/llms/ollama/internal/ollamaclient"
)

type options struct {
	ollamaServerURL     *url.URL
	httpClient          *http.Client
	model               string
	ollamaOptions       ollamaclient.Options
	customModelTemplate string
	system              string
	format              string
	keepAlive           string
}

type Option func(*options)

// WithModel Set the model to use.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithFormat Sets the Ollama output format (currently Ollama only supports "json").
func WithFormat(format string) Option {
	return func(opts *options) {
		opts.format = format
	}
}

// WithKeepAlive controls how long the model will stay loaded into memory following the request (default: 5m)
// only supported by ollama v0.1.23 and later
//
//	If set to a positive duration (e.g. 20m, 1h or 30), the model will stay loaded for the provided duration
//	If set to a negative duration (e.g. -1), the model will stay loaded indefinitely
//	If set to 0, the model will be unloaded immediately once finished
//	If not set, the model will stay loaded for 5 minutes by default
func WithKeepAlive(keepAlive string) Option {
	return func(opts *options) {
		opts.keepAlive = keepAlive
	}
}

// WithSystem Set the system prompt. This is only valid if
// WithCustomTemplate is not set and the ollama model use
// .System in its model template OR if WithCustomTemplate
// is set using {{.System}}.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}

// WithCustomTemplate To override the templating done on Ollama model side.
func WithCustomTemplate(template string) Option {
	return func(opts *options) {
		opts.customModelTemplate = template
	}
}

// WithServerURL Set the URL of the ollama instance to use.
func WithServerURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.ollamaServerURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WithHTTPClient Set custom http client.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithBackendUseNUMA Use NUMA optimization on certain systems.
func WithRunnerUseNUMA(numa bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseNUMA = numa
	}
}

// WithRunnerNumCtx Sets the size of the context window used to generate the next token (Default: 2048).
func WithRunnerNumCtx(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumCtx = num
	}
}

// WithRunnerNumKeep Specify the number of tokens from the initial prompt to retain when the model resets
// its internal context.
func WithRunnerNumKeep(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumKeep = num
	}
}

// WithRunnerNumBatch Set the batch size for prompt processing (default: 512).
func WithRunnerNumBatch(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumBatch = num
	}
}

// WithRunnerNumThread Set the number of threads to use during computation (default: auto).
func WithRunnerNumThread(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumThread = num
	}
}

// WithRunnerNumGQA The number of GQA groups in the transformer layer. Required for some models.
func WithRunnerNumGQA(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumGQA = num
	}
}

// WithRunnerNumGPU The number of layers to send to the GPU(s).
// On macOS it defaults to 1 to enable metal support, 0 to disable.
func WithRunnerNumGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumGPU = num
	}
}

// WithRunnerMainGPU When using multiple GPUs this option controls which GPU is used for small tensors
// for which the overhead of splitting the computation across all GPUs is not worthwhile.
// The GPU in question will use slightly more VRAM to store a scratch buffer for temporary results.
// By default GPU 0 is used.
func WithRunnerMainGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.MainGPU = num
	}
}

// WithRunnerLowVRAM Do not allocate a VRAM scratch buffer for holding temporary results.
// Reduces VRAM usage at the cost of performance, particularly prompt processing speed.
func WithRunnerLowVRAM(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.LowVRAM = val
	}
}

// WithRunnerF16KV If set to falsem, use 32-bit floats instead of 16-bit floats for memory key+value.
func WithRunnerF16KV(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.F16KV = val
	}
}

// WithRunnerLogitsAll Return logits for all tokens, not just the last token.
func WithRunnerLogitsAll(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.LogitsAll = val
	}
}

// WithRunnerVocabOnly Only load the vocabulary, no weights.
func WithRunnerVocabOnly(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.VocabOnly = val
	}
}

// WithRunnerUseMMap Set to false to not memory-map the model.
// By default, models are mapped into memory, which allows the system to load only the necessary parts
// of the model as needed.
func WithRunnerUseMMap(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseMMap = val
	}
}

// WithRunnerUseMLock Force system to keep model in RAM.
func WithRunnerUseMLock(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseMLock = val
	}
}

// WithRunnerEmbeddingOnly Only return the embbeding.
func WithRunnerEmbeddingOnly(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.EmbeddingOnly = val
	}
}

// WithRunnerRopeFrequencyBase RoPE base frequency (default: loaded from model).
func WithRunnerRopeFrequencyBase(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.RopeFrequencyBase = val
	}
}

// WithRunnerRopeFrequencyScale Rope frequency scaling factor (default: loaded from model).
func WithRunnerRopeFrequencyScale(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.RopeFrequencyScale = val
	}
}

// WithPredictTFSZ Tail free sampling is used to reduce the impact of less probable tokens from the output.
// A higher value (e.g., 2.0) will reduce the impact more, while a value of 1.0 disables this setting (default: 1).
func WithPredictTFSZ(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TFSZ = val
	}
}

// WithPredictTypicalP Enable locally typical sampling with parameter p (default: 1.0, 1.0 = disabled).
func WithPredictTypicalP(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TypicalP = val
	}
}

// WithPredictRepeatLastN Sets how far back for the model to look back to prevent repetition
// (Default: 64, 0 = disabled, -1 = num_ctx).
func WithPredictRepeatLastN(val int) Option {
	return func(opts *options) {
		opts.ollamaOptions.RepeatLastN = val
	}
}

// WithPredictMirostat Enable Mirostat sampling for controlling perplexity
// (default: 0, 0 = disabled, 1 = Mirostat, 2 = Mirostat 2.0).
func WithPredictMirostat(val int) Option {
	return func(opts *options) {
		opts.ollamaOptions.Mirostat = val
	}
}

// WithPredictMirostatTau Controls the balance between coherence and diversity of the output.
// A lower value will result in more focused and coherent text (Default: 5.0).
func WithPredictMirostatTau(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.MirostatTau = val
	}
}

// WithPredictMirostatEta Influences how quickly the algorithm responds to feedback from the generated text.
// A lower learning rate will result in slower adjustments, while a higher learning rate will make the
// algorithm more responsive (Default: 0.1).
func WithPredictMirostatEta(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.MirostatEta = val
	}
}

// WithPredictPenalizeNewline Penalize newline tokens when applying the repeat penalty (default: true).
func WithPredictPenalizeNewline(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.PenalizeNewline = val
	}
}
