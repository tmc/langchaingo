package googleai

import "github.com/tmc/langchaingo/llms"

// defaultCallOptions is the default set of call options used by the GoogleAI.GenerateContent method.
// nolint: gochecknoglobals
var defaultCallOptions llms.CallOptions = llms.CallOptions{
	Model: "gemini-pro",
}
