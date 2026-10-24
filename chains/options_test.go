package chains

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestGetLLMCallOptionsStreamingPresence(t *testing.T) {
	stream := func(context.Context, []byte) error { return nil }

	tests := []struct {
		name    string
		options []ChainCallOption
		want    func(context.Context, []byte) error
	}{
		{name: "absent", want: stream},
		{name: "explicit nil", options: []ChainCallOption{WithStreamingFunc(nil)}},
		{name: "explicit function", options: []ChainCallOption{WithStreamingFunc(stream)}, want: stream},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := llms.CallOptions{StreamingFunc: stream}
			for _, option := range GetLLMCallOptions(test.options...) {
				option(&opts)
			}
			if opts.StreamingFunc == nil && test.want != nil {
				t.Fatal("streaming function was cleared")
			}
			if opts.StreamingFunc != nil && test.want == nil {
				t.Fatal("streaming function was not cleared")
			}
		})
	}
}
