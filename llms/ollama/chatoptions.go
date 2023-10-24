package ollama

type chatOptions struct {
	chatTemplate string
	options
}

type ChatOption func(*chatOptions)

const (
	// a default chat template for LLama2 from https://www.philschmid.de/llama-2#how-to-prompt-llama-2-chat.
	// LLamacpp will add the BOS by default https://github.com/ggerganov/llama.cpp/tree/master/examples/server#api-endpoints.
	_defaultLLamaChatTemplate = `[INST] <<SYS>>
{{.system}}
<</SYS>>

{{ range $m := .messagesPair }}
 {{- $u := index $m 0 }}
 {{- $a := index $m 1 }}
 {{- if $u }}
   {{- $u }} [/INST]
 {{- end }}
 {{- if $a }}
 {{- " " }}{{- $a }}</s><s>[INST]{{" "}}
 {{- end }}
{{- end }}`
)

func defaultChatOptions() chatOptions {
	return chatOptions{
		chatTemplate: _defaultLLamaChatTemplate,
	}
}

// WithLLMOptions Set underlying LLM options.
func WithLLMOptions(opts ...Option) ChatOption {
	return func(copts *chatOptions) {
		for _, opt := range opts {
			opt(&copts.options)
		}
	}
}

// WithChatTemplate Set the chat go template to use (default: _defaultLLamaChatTemplate)
// The chat template expects the inputs variables system as string and messagesPair as
// a list of pair of string starting with user one.
func WithChatTemplate(t string) ChatOption {
	return func(opts *chatOptions) {
		opts.chatTemplate = t
	}
}
