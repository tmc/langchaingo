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
	errs    []error
	ends    int
	lastEnd *llms.ContentResponse
}

func (r *errorRecorder) HandleLLMError(_ context.Context, err error) {
	r.errs = append(r.errs, err)
}

func (r *errorRecorder) HandleLLMGenerateContentEnd(_ context.Context, resp *llms.ContentResponse) {
	r.ends++
	r.lastEnd = resp
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

func TestOneClosingCallbackPerCall(t *testing.T) {
	t.Parallel()

	ask := func(model *Model, rec *errorRecorder, opts ...llms.CallOption) error {
		model.CallbacksHandler = rec
		_, err := model.GenerateContent(context.Background(),
			[]llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, "hi")}, opts...)
		return err
	}

	t.Run("a plain answer ends once, with the answer", func(t *testing.T) {
		t.Parallel()
		rec := &errorRecorder{}
		if err := ask(truncationModel(t, "stop"), rec); err != nil {
			t.Fatalf("GenerateContent: %v", err)
		}
		if rec.ends != 1 {
			t.Errorf("closing callbacks = %d, want exactly 1", rec.ends)
		}
		if rec.lastEnd == nil {
			t.Error("the closing callback must carry the response, not nil")
		}
		if len(rec.errs) != 0 {
			t.Errorf("a successful call reports no error, got %v", rec.errs)
		}
	})

	t.Run("a failure reports the error and does not also end", func(t *testing.T) {
		t.Parallel()
		rec := &errorRecorder{}
		err := ask(truncationModel(t, "length"), rec, llms.WithFailOnTruncation())
		if err == nil {
			t.Fatal("WithFailOnTruncation must fail on a truncated answer")
		}
		if rec.ends != 0 {
			t.Errorf("a failed call must not report a completion, got %d", rec.ends)
		}
		if len(rec.errs) != 1 || rec.errs[0] == nil {
			t.Errorf("want exactly one non-nil error, got %v", rec.errs)
		}
	})

	t.Run("a streamed answer ends once too", func(t *testing.T) {
		t.Parallel()
		rec := &errorRecorder{}
		sink := func(_ context.Context, _ streaming.Chunk) error { return nil }
		if err := ask(streamingModelWithTrailingChunk(t), rec, llms.WithStreamingFunc(sink)); err != nil {
			t.Fatalf("GenerateContent: %v", err)
		}
		if rec.ends != 1 || rec.lastEnd == nil {
			t.Errorf("closing callbacks = %d (response nil: %v), want exactly 1 with the answer",
				rec.ends, rec.lastEnd == nil)
		}
	})
}
