package palmclient

import (
	"google.golang.org/api/option"
)

const (
	embeddingModelName = "text-embedding-005"
	textModelName      = "text-bison"
	chatModelName      = "chat-bison"
)

// Options are the Palm client options
type Options struct {
	EmbeddingModelName string
	TextModelName      string
	ChatModelName      string
	ClientOptions      []option.ClientOption
}

// Option is an option
type Option func(*Options)

// WithEmbeddingModelName sets the default embedding model
func WithEmbeddingModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.EmbeddingModelName = modelName
		}
	}
}

// WithTextModelName sets the default text model
func WithTextModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.TextModelName = modelName
		}
	}
}

// WithChatModelName sets the default chat model
func WithChatModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.ChatModelName = modelName
		}
	}
}

// WithClientOptions sets the client options for the Google API client
func WithClientOptions(opts ...option.ClientOption) Option {
	return func(o *Options) {
		o.ClientOptions = append(o.ClientOptions, opts...)
	}
}

func defaultOptions() Options {
	return Options{
		EmbeddingModelName: embeddingModelName,
		TextModelName:      textModelName,
		ChatModelName:      chatModelName,
	}
}
