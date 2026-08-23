package mistral

import (
	"context"
	"testing"

	"github.com/vxcontrol/langchaingo/callbacks"
	"github.com/vxcontrol/langchaingo/llms"
	"github.com/vxcontrol/langchaingo/llms/streaming"
)

type errorRecorder struct {
	callbacks.SimpleHandler
	errs []error
	ends int
}

func (r *errorRecorder) HandleLLMError(_ context.Context, err error) {
	r.errs = append(r.errs, err)
}

func (r *errorRecorder) HandleLLMGenerateContentEnd(_ context.Context, _ *llms.ContentResponse) {
	r.ends++
}

func TestStreamingErrorsReachTheCallbackHandler(t *testing.T) {
	t.Parallel()

	model := streamingModelWithTrailingChunk(t)
	rec := &errorRecorder{}
	model.CallbacksHandler = rec

	sink := func(_ context.Context, _ streaming.Chunk) error { return nil }
	_, err := model.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(sink), llms.WithFailOnTruncation())
	if err == nil {
		t.Fatal("a truncated stream with WithFailOnTruncation must fail")
	}

	if len(rec.errs) == 0 {
		t.Fatal("the streaming path must report its error to the handler, as the non-streaming path does")
	}
	if rec.errs[len(rec.errs)-1] == nil {
		t.Error("the handler must receive the error itself, not nil")
	}
}

func TestStreamingSuccessReachesTheCallbackHandler(t *testing.T) {
	t.Parallel()

	model := streamingModelWithTrailingChunk(t)
	rec := &errorRecorder{}
	model.CallbacksHandler = rec

	sink := func(_ context.Context, _ streaming.Chunk) error { return nil }
	if _, err := model.GenerateContent(context.Background(),
		[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")},
		llms.WithStreamingFunc(sink)); err != nil {
		t.Fatalf("GenerateContent: %v", err)
	}

	if rec.ends != 1 {
		t.Errorf("the streaming path must close the generation once, as the non-streaming path does; got %d", rec.ends)
	}
	if len(rec.errs) != 0 {
		t.Errorf("a successful stream must report no error, got %v", rec.errs)
	}
}
