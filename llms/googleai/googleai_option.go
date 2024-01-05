package googleai

import (
	"github.com/tmc/langchaingo/llms"
)

// nolint: gochecknoglobals
var defaultCallOptions *llms.CallOptions = &llms.CallOptions{
	Model: "gemini-pro",
}
