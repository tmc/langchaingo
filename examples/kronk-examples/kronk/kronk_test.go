package kronk

import (
	"reflect"
	"testing"

	"github.com/tmc/langchaingo/llms"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
)

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		name string
		opts llms.CallOptions
		want model.D
	}{
		{
			name: "defaults",
			want: model.D{},
		},
		{
			name: "call sampling options",
			opts: llms.CallOptions{
				MaxTokens:         128,
				Temperature:       0.2,
				TopK:              20,
				TopP:              0.95,
				Seed:              42,
				StopWords:         []string{"END"},
				RepetitionPenalty: 1.1,
				FrequencyPenalty:  0.3,
				PresencePenalty:   0.4,
			},
			want: model.D{
				"max_tokens":        128,
				"temperature":       0.2,
				"top_k":             20,
				"top_p":             0.95,
				"seed":              42,
				"stop":              []string{"END"},
				"repeat_penalty":    1.1,
				"frequency_penalty": 0.3,
				"presence_penalty":  0.4,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{}
			got := client.applyOptions(model.D{}, tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyOptions: got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestModelOptionsPreserveOrder(t *testing.T) {
	var opts options
	for _, opt := range []Option{
		WithAutoTune(true),
		WithConfig(model.Config{}),
		WithAutoTune(true),
		WithContextWindow(8192),
	} {
		opt(&opts)
	}

	cfg := model.NewConfig(opts.modelOptions...)
	if !cfg.AutoTune {
		t.Error("AutoTune: got false, want true")
	}
	if cfg.ContextWindow() != 8192 {
		t.Errorf("ContextWindow: got %d, want 8192", cfg.ContextWindow())
	}
}
