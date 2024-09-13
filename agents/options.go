package agents

import (
	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/i18n"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

type translatable struct {
	promptPrefix       string
	formatInstructions string
	promptSuffix       string
	outputKey          string
}

type Options struct {
	prompt                  prompts.PromptTemplate
	memory                  schema.Memory
	callbacksHandler        callbacks.Handler
	errorHandler            *ParserErrorHandler
	maxIterations           int
	returnIntermediateSteps bool
	lang                    i18n.Lang
	translatable

	// openai
	systemMessage string
	extraMessages []prompts.MessageFormatter
}

// Option is a function type that can be used to modify the creation of the agents
// and executors.
type Option func(*Options)

func defaultOptions() Options {
	return Options{
		lang: i18n.DefaultLang,
	}
}

func executorDefaultOptions() Options {
	options := defaultOptions()
	options.maxIterations = _defaultMaxIterations
	options.memory = memory.NewSimple()
	return options
}

func mrklDefaultOptions() Options {
	return defaultOptions()
}

func conversationalDefaultOptions() Options {
	return defaultOptions()
}

func openAIFunctionsDefaultOptions() Options {
	return defaultOptions()
}

func (co *Options) loadExecutorTranslatable() {
	co.outputKey = i18n.AgentsMustPhrase(co.lang, "output key")
}

func (co *Options) loadMrklTranslatable() {
	co.promptPrefix = i18n.AgentsMustLoad(co.lang, "mrkl_prompt_prefix.txt")
	co.formatInstructions = i18n.AgentsMustLoad(co.lang, "mrkl_prompt_format_instructions.txt")
	co.promptSuffix = i18n.AgentsMustLoad(co.lang, "mrkl_prompt_suffix.txt")
	co.outputKey = i18n.AgentsMustPhrase(co.lang, "output key")
}

func (co *Options) loadConversationalTranslatable() {
	co.promptPrefix = i18n.AgentsMustLoad(co.lang, "conversational_prompt_prefix.txt")
	co.formatInstructions = i18n.AgentsMustLoad(co.lang, "conversational_prompt_format_instructions.txt")
	co.promptSuffix = i18n.AgentsMustLoad(co.lang, "conversational_prompt_suffix.txt")
	co.outputKey = i18n.AgentsMustPhrase(co.lang, "output key")
}

func (co *Options) loadOpenAIFunctionsTranslatable() {
	co.systemMessage = i18n.AgentsMustPhrase(co.lang, "system message")
	co.outputKey = i18n.AgentsMustPhrase(co.lang, "output key")
}

func (co *Options) getMrklPrompt(tools []tools.Tool) prompts.PromptTemplate {
	if co.prompt.Template != "" {
		return co.prompt
	}

	return createMRKLPrompt(
		tools,
		co.promptPrefix,
		co.formatInstructions,
		co.promptSuffix,
	)
}

func (co *Options) getConversationalPrompt(tools []tools.Tool) prompts.PromptTemplate {
	if co.prompt.Template != "" {
		return co.prompt
	}

	return createConversationalPrompt(
		tools,
		co.promptPrefix,
		co.formatInstructions,
		co.promptSuffix,
	)
}

// WithMaxIterations is an option for setting the max number of iterations the executor
// will complete.
func WithMaxIterations(iterations int) Option {
	return func(co *Options) {
		co.maxIterations = iterations
	}
}

// WithOutputKey is an option for setting the output key of the agent.
func WithOutputKey(outputKey string) Option {
	return func(co *Options) {
		co.outputKey = outputKey
	}
}

// WithPromptPrefix is an option for setting the prefix of the prompt used by the agent.
func WithPromptPrefix(prefix string) Option {
	return func(co *Options) {
		co.promptPrefix = prefix
	}
}

// WithPromptFormatInstructions is an option for setting the format instructions of the prompt
// used by the agent.
func WithPromptFormatInstructions(instructions string) Option {
	return func(co *Options) {
		co.formatInstructions = instructions
	}
}

// WithPromptSuffix is an option for setting the suffix of the prompt used by the agent.
func WithPromptSuffix(suffix string) Option {
	return func(co *Options) {
		co.promptSuffix = suffix
	}
}

// WithPrompt is an option for setting the prompt the agent will use.
func WithPrompt(prompt prompts.PromptTemplate) Option {
	return func(co *Options) {
		co.prompt = prompt
	}
}

// WithReturnIntermediateSteps is an option for making the executor return the intermediate steps
// taken.
func WithReturnIntermediateSteps() Option {
	return func(co *Options) {
		co.returnIntermediateSteps = true
	}
}

// WithMemory is an option for setting the memory of the executor.
func WithMemory(m schema.Memory) Option {
	return func(co *Options) {
		co.memory = m
	}
}

// WithCallbacksHandler is an option for setting a callback handler to an executor.
func WithCallbacksHandler(handler callbacks.Handler) Option {
	return func(co *Options) {
		co.callbacksHandler = handler
	}
}

// WithParserErrorHandler is an option for setting a parser error handler to an executor.
func WithParserErrorHandler(errorHandler *ParserErrorHandler) Option {
	return func(co *Options) {
		co.errorHandler = errorHandler
	}
}

// WithLanguage is an option for setting language the prompt will use.
func WithLanguage(language i18n.Lang) Option {
	return func(co *Options) {
		co.lang = language
	}
}

type OpenAIOption struct{}

func NewOpenAIOption() OpenAIOption {
	return OpenAIOption{}
}

func (o OpenAIOption) WithSystemMessage(msg string) Option {
	return func(co *Options) {
		co.systemMessage = msg
	}
}

func (o OpenAIOption) WithExtraMessages(extraMessages []prompts.MessageFormatter) Option {
	return func(co *Options) {
		co.extraMessages = extraMessages
	}
}
