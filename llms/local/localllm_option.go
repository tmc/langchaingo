package local

const (
	// The name of the environment variable that contains the path to the local LLM binary.
	localLLMBinVarName = "LOCAL_LLM_BIN"
	// The name of the environment variable that contains the CLI arguments to pass to the local LLM binary.
	localLLMArgsVarName = "LOCAL_LLM_ARGS"
)

type options struct {
	bin          string
	args         string
	globalAsArgs bool // build key-value arguments from global llms.Options
}

type Option func(*options)

// WithBin passes the path to the local LLM binary to the client.
// If not set, then will be used the LOCAL_LLM_BIN environment variable.
func WithBin(bin string) Option {
	return func(opts *options) {
		opts.bin = bin
	}
}

// WithArgs passes the CLI arguments to the local LLM binary.
// If not set, then will be used the LOCAL_LLM_ARGS environment variable.
func WithArgs(args string) Option {
	return func(opts *options) {
		opts.args = args
	}
}

// WithGlobalAsArgs passes the CLI arguments to the local LLM binary
// formed from global llms.Options.
func WithGlobalAsArgs() Option {
	return func(opts *options) {
		opts.globalAsArgs = true
	}
}
