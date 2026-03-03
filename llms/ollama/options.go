package ollama

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"
)

type options struct {
	ollamaServerURL     *url.URL
	httpClient          *http.Client
	apiKey              string
	model               string
	ollamaOptions       api.Options
	pullProgressFunc    api.PullProgressFunc
	customModelTemplate string
	system              string
	format              string
	keepAlive           *time.Duration
	pullModel           bool
	pullTimeout         time.Duration
}

type Option func(*options)

// WithModel sets the name of the Ollama model to use for generation.
// This should be a model name familiar to Ollama from the library at https://ollama.com/library
// (e.g., "llama3.2", "mistral", "codellama"). The model must be available locally or pulled first.
func WithModel(model string) Option {
	return func(opts *options) {
		opts.model = model
	}
}

// WithFormat specifies the format for the model's response output.
// Currently Ollama supports "json" to force JSON-formatted responses from compatible models.
// When set to "json", the model will structure its output as valid JSON.
// Leave empty for default text format.
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
		ka, err := time.ParseDuration(keepAlive)
		if err != nil {
			log.Fatal(err)
		}
		opts.keepAlive = &ka
	}
}

// WithSystemPrompt sets the system message that guides the model's behavior and personality.
// The system prompt is prepended to conversations and helps establish context, tone, and instructions.
// This is only effective if the model's template includes {{.System}} or if you're using WithCustomTemplate
// with {{.System}}. Most modern chat models support system prompts for role-playing and behavior control.
func WithSystemPrompt(p string) Option {
	return func(opts *options) {
		opts.system = p
	}
}

// WithCustomTemplate overrides the default prompt template used by the model.
// This allows you to customize how prompts are formatted before being sent to the model.
// The template uses Go template syntax with variables like {{.System}}, {{.Prompt}}, etc.
// Use this when you need fine-grained control over prompt formatting or when working with custom models.
func WithCustomTemplate(template string) Option {
	return func(opts *options) {
		opts.customModelTemplate = template
	}
}

// WithServerURL sets the base URL of the Ollama server instance to connect to.
// Use this to connect to a remote Ollama server or when running Ollama on a non-default port.
// The URL should include the protocol and port (e.g., "http://localhost:11434", "https://my-ollama-server.com").
// Defaults to "http://localhost:11434" if not specified.
func WithServerURL(rawURL string) Option {
	return func(opts *options) {
		var err error
		opts.ollamaServerURL, err = url.Parse(rawURL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// WithHTTPClient sets a custom HTTP client for requests to the Ollama server.
// This allows you to configure custom timeouts, retry policies, proxy settings, or TLS configuration.
// Useful for production environments that require specific networking configurations or when integrating
// with existing HTTP client pools and middleware.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *options) {
		opts.httpClient = client
	}
}

// WithAPIKey sets the API key for authenticating with Ollama Cloud.
// When using Ollama Cloud (https://ollama.com), you need to provide an API key for authentication.
// You can create an API key at https://ollama.com/settings/keys.
// This option can also be set via the OLLAMA_API_KEY environment variable.
// The API key is added as an "Authorization: Bearer <key>" header to all requests.
// This option works seamlessly with custom HTTP clients provided via WithHTTPClient.
func WithAPIKey(apiKey string) Option {
	return func(opts *options) {
		opts.apiKey = apiKey
	}
}

// WithRunnerNumCtx sets the size of the context window used to generate the next token.
// The context window determines how many tokens the model can consider when generating responses.
// Larger context windows allow for longer conversations and more detailed responses, but require more memory.
// This setting must be configured when the model is loaded (default: varies by model, typically 2048-4096).
func WithRunnerNumCtx(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumCtx = num
	}
}

// WithRunnerNumKeep specifies the number of tokens from the initial prompt to retain when the model resets
// its internal context due to context length limits. This helps maintain conversation continuity by keeping
// important initial context (like system prompts) even when the context window fills up.
// Set to 0 to reset completely, or use a positive number to retain key tokens (default: 4).
func WithRunnerNumKeep(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumKeep = num
	}
}

// WithRunnerNumBatch sets the batch size for prompt processing to optimize inference performance.
// Larger batch sizes can improve throughput for long prompts but require more memory.
// Smaller batch sizes reduce memory usage but may be slower for processing large inputs.
// This setting affects how the model processes input tokens in chunks (default: 512).
func WithRunnerNumBatch(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumBatch = num
	}
}

// WithRunnerNumThread sets the number of CPU threads to use during model computation.
// More threads can improve performance on multi-core systems, but too many threads may cause overhead.
// Set to 0 to let the runtime automatically determine the optimal number of threads based on your CPU.
// Useful for fine-tuning performance on specific hardware configurations (default: 0, auto-detect).
func WithRunnerNumThread(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumThread = num
	}
}

// WithRunnerNumGPU controls the number of model layers to offload to GPU(s) for acceleration.
// Higher values offload more layers to GPU, improving performance but requiring more VRAM.
// Set to 0 to run entirely on CPU, or -1 to automatically determine optimal GPU usage.
// On macOS with Metal support, this enables hardware acceleration (default: -1, auto-detect).
func WithRunnerNumGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumGPU = num
	}
}

// WithRunnerMainGPU specifies which GPU to use as the primary device in multi-GPU setups.
// When using multiple GPUs, this controls which GPU handles small tensors and operations where
// the overhead of splitting computation across all GPUs is not beneficial. The selected GPU
// will use slightly more VRAM to store scratch buffers for temporary results.
// Only relevant when you have multiple GPUs and NumGPU > 1 (default: 0, first GPU).
func WithRunnerMainGPU(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.MainGPU = num
	}
}

// WithSeed sets the random number seed for generation to ensure reproducible outputs.
// Setting this to a specific number will make the model generate the same text for the same prompt.
// Use -1 for random seed (default: -1).
func WithSeed(seed int) Option {
	return func(opts *options) {
		opts.ollamaOptions.Seed = seed
	}
}

// WithNumPredict sets the maximum number of tokens to predict when generating text.
// This parameter controls the maximum length of the response. Use -1 for infinite generation,
// -2 to fill context, or a positive number to limit tokens (default: -1).
func WithNumPredict(num int) Option {
	return func(opts *options) {
		opts.ollamaOptions.NumPredict = num
	}
}

// WithTopK reduces the probability of generating nonsense by limiting token selection to the top K choices.
// A higher value (e.g., 100) will give more diverse answers, while a lower value (e.g., 10)
// will be more conservative and focused (default: 40).
func WithTopK(topK int) Option {
	return func(opts *options) {
		opts.ollamaOptions.TopK = topK
	}
}

// WithTopP controls nucleus sampling by setting the cumulative probability threshold for token selection.
// Works together with top-k. A higher value (e.g., 0.95) will lead to more diverse text,
// while a lower value (e.g., 0.5) will generate more focused and conservative text (default: 0.9).
func WithTopP(topP float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TopP = topP
	}
}

// WithMinP sets the minimum probability threshold for token selection relative to the most likely token.
// This is an alternative to top-p that aims to ensure a balance of quality and variety.
// For example, with min_p=0.05 and the most likely token having probability 0.9,
// tokens with probability less than 0.045 are filtered out (default: 0.0).
func WithMinP(minP float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.MinP = minP
	}
}

// WithTypicalP enables locally typical sampling, which selects tokens that are close to the expected
// information content. This method can produce higher quality output than nucleus sampling.
// Values closer to 1.0 are more conservative, while lower values increase diversity (default: 1.0).
func WithTypicalP(typicalP float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.TypicalP = typicalP
	}
}

// WithRunnerUseMMap controls whether to memory-map the model file.
// When enabled (true), models are mapped into memory, allowing the system to load only the necessary parts
// of the model as needed, which can improve memory efficiency. Set to false to disable memory mapping
// if you encounter issues or want to load the entire model into RAM (default: true).
func WithRunnerUseMMap(val bool) Option {
	return func(opts *options) {
		opts.ollamaOptions.UseMMap = &val
	}
}

// WithPredictRepeatLastN sets how far back the model should look to prevent repetition.
// This parameter controls the sliding window of tokens that the model considers when applying
// repeat penalties. Use 0 to disable, -1 to use num_ctx (context size), or a positive number
// to specify the exact number of recent tokens to consider (default: 64).
func WithPredictRepeatLastN(val int) Option {
	return func(opts *options) {
		opts.ollamaOptions.RepeatLastN = val
	}
}

// WithPredictRepeatPenalty controls how strongly to penalize repetitions in the generated text.
// A higher value (e.g., 1.5) will penalize repetitions more strongly and encourage diversity,
// while a lower value (e.g., 0.9) will be more lenient and allow more repetition.
// Values > 1.0 discourage repetition, values < 1.0 encourage repetition (default: 1.1).
func WithPredictRepeatPenalty(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.RepeatPenalty = val
	}
}

// WithPredictPresencePenalty penalizes new tokens based on whether they appear in the text so far.
// Unlike frequency penalty, this applies the same penalty to any token that has already been used,
// regardless of how many times it appeared. Positive values (0.0 to 2.0) encourage the model
// to talk about new topics, while negative values encourage repetition (default: 0.0).
func WithPredictPresencePenalty(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.PresencePenalty = val
	}
}

// WithPredictFrequencyPenalty penalizes new tokens based on their existing frequency in the text so far.
// The penalty scales with how often a token has been used - more frequent tokens receive higher penalties.
// Positive values (0.0 to 2.0) reduce repetition and encourage diverse vocabulary,
// while negative values encourage the model to reuse frequent terms (default: 0.0).
func WithPredictFrequencyPenalty(val float32) Option {
	return func(opts *options) {
		opts.ollamaOptions.FrequencyPenalty = val
	}
}

// WithPredictStop sets the stop tokens that will halt text generation when encountered.
// When the model generates any of these strings, it will immediately stop generating further tokens.
// This is useful for creating structured outputs or stopping generation at specific points.
// Pass an empty slice to disable stop tokens (default: empty slice).
func WithPredictStop(stop []string) Option {
	return func(opts *options) {
		opts.ollamaOptions.Stop = stop
	}
}

// WithPullModel enables automatic model downloading before use.
// When enabled, the client will check if the specified model exists locally and automatically
// download it from the Ollama library if not available. This is convenient for ensuring models
// are available without manual intervention, but may cause delays on first use (default: false).
func WithPullModel() Option {
	return func(opts *options) {
		opts.pullModel = true
	}
}

// WithPullProgressFunc sets a function to be called when the model is being downloaded.
// This allows you to track the progress of the download and display it to the user.
func WithPullProgressFunc(progressFunc api.PullProgressFunc) Option {
	return func(opts *options) {
		opts.pullProgressFunc = progressFunc
	}
}

// WithPullTimeout sets a maximum duration for model download operations.
// This prevents model pulling from hanging indefinitely on slow connections or server issues.
// If not set or duration is 0, pull operations will use the request context without additional timeout.
// This option only takes effect when WithPullModel is also enabled. Use reasonable values like 10-30 minutes
// depending on model size and connection speed (default: no timeout, uses request context).
func WithPullTimeout(timeout time.Duration) Option {
	return func(opts *options) {
		opts.pullTimeout = timeout
	}
}
