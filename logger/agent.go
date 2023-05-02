package logger

import (
	"github.com/fatih/color"
)

// AgentThought logs a thought from the agent.
//
//nolint:forbidigo // This is a debug function.
func AgentThought(msg string) {
	// Display banner
	agentBanner()

	// Display thought
	message("Thought", msg, color.HiMagenta)
}

func agentBanner() {
	// Display banner
	banner("Agent Action")
}
