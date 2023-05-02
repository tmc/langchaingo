package logger

import (
	"github.com/fatih/color"
)

// LLMRequest logs a request to the LLM.
func LLMRequest(msg string) {
	// Display banner
	llmBanner()

	// Display question
	message("Submitted query", msg, color.Cyan)
}

// LLMResponse logs a response from the LLM.
func LLMResponse(msg string) {
	// Display banner
	llmBanner()

	// Display answer
	message("Received response", msg, color.Green)
}

// LLMError logs an error from the LLM.
func LLMError(err error) {
	// Display banner
	llmBanner()

	// Display error
	message("Received error", err.Error(), color.Red)
}

func llmBanner() {
	// Display banner
	banner("LLM Query")
}
