package reasoning

// Mechanism is how a model thinks on one request.
type Mechanism int

const (
	MechanismNone Mechanism = iota
	MechanismAdaptive
	MechanismBudget
	MechanismNovaReasoningConfig
	MechanismGrokEffort
)

// ResolveMechanism picks the thinking mechanism for a model the caller asked to
// think. claudeFamily and reasons are the door's own answers: whether the id
// belongs to Anthropic on that platform, and whether the door treats the model
// as a reasoning model at all. A family with no mechanism of its own resolves to
// MechanismNone and the request travels without a thinking configuration.
func ResolveMechanism(model string, adaptive, claudeFamily, reasons bool) Mechanism {
	switch {
	case ClaudeSupportsThinking(model):
		if ResolveClaudeAdaptive(model, adaptive) {
			return MechanismAdaptive
		}
		return MechanismBudget
	case adaptive && claudeFamily && !ClaudePredatesAdaptive(model):
		return MechanismAdaptive
	case IsNovaReasoningModel(model):
		return MechanismNovaReasoningConfig
	case IsGrokModel(model):
		return MechanismGrokEffort
	case IsBedrockAlwaysReasoningModel(model):
		return MechanismNone
	case reasons && claudeFamily:
		return MechanismBudget
	default:
		return MechanismNone
	}
}
