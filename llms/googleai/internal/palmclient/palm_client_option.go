package palmclient

import (
	"google.golang.org/api/option"
)

const (
	embeddingModelName = "textembedding-gecko"
	textModelName      = "text-bison"
	chatModelName      = "chat-bison"
)

type Options struct {
	EmbeddingModelName string
	TextModelName      string
	ChatModelName      string
	ClientOptions      []option.ClientOption
}

type Option func(*Options)

func WithEmbeddingModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.EmbeddingModelName = modelName
		}
	}
}

func WithTextModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.TextModelName = modelName
		}
	}
}

func WithChatModelName(modelName string) Option {
	return func(o *Options) {
		if modelName != "" {
			o.ChatModelName = modelName
		}
	}
}

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
