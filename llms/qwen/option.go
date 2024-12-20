package qwen

import (
	"github.com/tmc/langchaingo/callbacks"
)

const (
	QwenAPIKey = "Qwen_API_KEY"
)

const (
	ModelNameQwen_VL_Plus         = "Qwen-VL-Plus"
	ModelNameQwen_VL_Max          = "Qwen-VL-Max"
	ModelNameQwen_Turbo           = "Qwen-Turbo"
	ModelNameQwen_Plus            = "Qwen-Plus"
	ModelNameQwen_Max             = "Qwen-Max"
	ModelNameQwen_QwQ_32B_Preview = "Qwen-QwQ-32B-Preview"
)

type ModelName string

type Options struct {
	ApiKey           string
	ModelName        ModelName
	CallbacksHandler callbacks.Handler
}

type Option func(*Options)

func WithAK(apiKey string) Option {
	return func(opts *Options) {
		opts.ApiKey = apiKey
	}
}

func WithModelName(modelName ModelName) Option {
	return func(opts *Options) {
		opts.ModelName = modelName
	}
}

func WithCallbackHandler(callbacksHandler callbacks.Handler) Option {
	return func(opts *Options) {
		opts.CallbacksHandler = callbacksHandler
	}
}
