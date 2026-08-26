package reasoning

// Mechanism is how a model thinks on one request.
type Mechanism int

const (
	// MechanismNone sends no thinking configuration at all.
	MechanismNone Mechanism = iota
	// MechanismAdaptive lets the model size its own thinking from an effort.
	MechanismAdaptive
	// MechanismBudget gives the model a token budget to think within.
	MechanismBudget
)

// ResolveMechanism picks the thinking mechanism for a model the caller asked to
// think. claudeFamily and reasons are the door's own answers: whether the id
// belongs to Anthropic on that platform, and whether the door treats the model
// as a reasoning model at all.
func ResolveMechanism(model string, adaptive, claudeFamily, reasons bool) Mechanism {
	switch {
	case ClaudeSupportsThinking(model):
		if ResolveClaudeAdaptive(model, adaptive) {
			return MechanismAdaptive
		}
		return MechanismBudget
	case adaptive && claudeFamily && !ClaudePredatesAdaptive(model):
		return MechanismAdaptive
	case reasons:
		return MechanismBudget
	default:
		return MechanismNone
	}
}
