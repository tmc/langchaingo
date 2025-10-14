package llamafile

import "github.com/vendasta/langchaingo/llms/llamafile/internal/llamafileclient"

type Option func(*llamafileclient.GenerationSettings)

// / WithFrequencyPenalty sets the frequency penalty.
func WithFrequencyPenalty(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.FrequencyPenalty = val
	}
}

// WithGrammar sets the grammar.
func WithGrammar(val string) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Grammar = val
	}
}

// WithIgnoreEOS sets the ignore EOS flag.
func WithIgnoreEOS(val bool) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.IgnoreEOS = val
	}
}

// WithMinP sets the minimum probability.
func WithMinP(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.MinP = val
	}
}

// WithMirostat sets the mirostat.
func WithMirostat(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Mirostat = val
	}
}

// WithMirostatEta sets the mirostat eta.
func WithMirostatEta(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.MirostatEta = val
	}
}

// WithMirostatTau sets the mirostat tau.
func WithMirostatTau(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.MirostatTau = val
	}
}

// WithModel sets the model.
func WithModel(val string) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Model = val
	}
}

// WithLogitBias sets the logit bias.
func WithLogitBias(val []interface{}) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.LogitBias = val
	}
}

// WithPenaltyPromptTokens sets the penalty prompt tokens.
func WithPenaltyPromptTokens(val []interface{}) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.PenaltyPromptTokens = val
	}
}

// WithPresencePenalty sets the presence penalty.
func WithPresencePenalty(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.PresencePenalty = val
	}
}

// WithRepeatLastN sets the repeat last N.
func WithRepeatLastN(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.RepeatLastN = val
	}
}

// WithRepeatPenalty sets the repeat penalty.
func WithRepeatPenalty(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.RepeatPenalty = val
	}
}

// WithSeed sets the seed.
func WithSeed(val uint32) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Seed = val
	}
}

// WithStop sets the stop tokens.
func WithStop(val []string) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Stop = val
	}
}

// WithStream sets the stream mode.
func WithStream(val bool) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Stream = val
	}
}

// WithTemperature sets the temperature.
func WithTemperature(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.Temperature = val
	}
}

// WithTfsZ sets the TfsZ.
func WithTfsZ(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.TfsZ = val
	}
}

// WithTopK sets the top K.
func WithTopK(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.TopK = val
	}
}

// WithTopP sets the top P.
func WithTopP(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.TopP = val
	}
}

// WithTypicalP sets the typical P.
func WithTypicalP(val float64) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.TypicalP = val
	}
}

// WithUsePenaltyPromptTokens sets the use penalty prompt tokens flag.
func WithUsePenaltyPromptTokens(val bool) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.UsePenaltyPromptTokens = val
	}
}

// WithNPredict sets the number of predictions.
func WithNPredict(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.NPredict = val
	}
}

// WithNProbs sets the number of probabilities.
func WithNProbs(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.NProbs = val
	}
}

// WithPenalizeNL sets the penalize newline option.
func WithPenalizeNL(val bool) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.PenalizeNL = val
	}
}

// WithNKeep sets the number of items to keep.
func WithNKeep(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.NKeep = val
	}
}

// WithNCtx sets the context number.
func WithNCtx(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.NCtx = val
	}
}

// set size of embeddings.
func WithEmbeddingSize(val int) Option {
	return func(g *llamafileclient.GenerationSettings) {
		g.EmbeddingSize = val
	}
}
