package chains_test

import (
	"context"
	"testing"

	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
)

// externalChain is an example Chain implementation living outside the chains
// package. Its existence is the regression test for issue #1085: before
// ChainCallOptions was exported, no custom Chain could observe the values set
// by the caller's ChainCallOption functions.
type externalChain struct {
	captured chains.ChainCallOptions
}

var _ chains.Chain = (*externalChain)(nil)

func (c *externalChain) Call(_ context.Context, _ map[string]any, opts ...chains.ChainCallOption) (map[string]any, error) {
	c.captured = chains.ApplyChainCallOptions(opts...)
	return map[string]any{"out": ""}, nil
}

func (c *externalChain) GetMemory() schema.Memory { return nil }
func (c *externalChain) GetInputKeys() []string   { return nil }
func (c *externalChain) GetOutputKeys() []string  { return []string{"out"} }

func TestApplyChainCallOptions_ExternalChain(t *testing.T) {
	c := &externalChain{}

	if _, err := c.Call(context.Background(), nil,
		chains.WithModel("gpt-4o-mini"),
		chains.WithMaxTokens(256),
		chains.WithTemperature(0.2),
		chains.WithStopWords([]string{"END"}),
	); err != nil {
		t.Fatalf("Call: %v", err)
	}

	got := c.captured
	if !got.ModelSet || got.Model != "gpt-4o-mini" {
		t.Errorf("Model = %q set=%v, want gpt-4o-mini set=true", got.Model, got.ModelSet)
	}
	if !got.MaxTokensSet || got.MaxTokens != 256 {
		t.Errorf("MaxTokens = %d set=%v, want 256 set=true", got.MaxTokens, got.MaxTokensSet)
	}
	if !got.TemperatureSet || got.Temperature != 0.2 {
		t.Errorf("Temperature = %v set=%v, want 0.2 set=true", got.Temperature, got.TemperatureSet)
	}
	if !got.StopWordsSet || len(got.StopWords) != 1 || got.StopWords[0] != "END" {
		t.Errorf("StopWords = %v set=%v, want [END] set=true", got.StopWords, got.StopWordsSet)
	}
	if got.TopKSet || got.TopPSet || got.SeedSet {
		t.Errorf("unset flags leaked: %+v", got)
	}
}

func TestApplyChainCallOptions_Empty(t *testing.T) {
	got := chains.ApplyChainCallOptions()
	if got.ModelSet || got.MaxTokensSet || got.TemperatureSet || got.StopWordsSet ||
		got.TopKSet || got.TopPSet || got.SeedSet ||
		got.MinLengthSet || got.MaxLengthSet || got.RepetitionPenaltySet {
		t.Errorf("empty apply produced set flags: %+v", got)
	}
	if got.StreamingFunc != nil || got.CallbackHandler != nil {
		t.Errorf("empty apply produced non-nil closures: %+v", got)
	}
}
