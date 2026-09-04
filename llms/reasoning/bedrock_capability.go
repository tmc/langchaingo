package reasoning

import "strings"

// IsNovaReasoningModel reports whether the Bedrock model takes Amazon Nova's
// reasoningConfig.
func IsNovaReasoningModel(model string) bool {
	return strings.Contains(baseModelName(model), "amazon.nova-2")
}

// NovaEffort maps a requested effort onto the three levels Nova accepts. Bedrock
// answers anything else with "is not a valid enum value".
func NovaEffort(effort string) string {
	switch strings.ToLower(effort) {
	case "minimal", "low":
		return "low"
	case "medium":
		return "medium"
	default:
		return "high"
	}
}

// NovaRejectsSamplingAt reports whether Nova refuses temperature, topP and topK
// alongside the given effort.
func NovaRejectsSamplingAt(effort string) bool {
	return NovaEffort(effort) == "high"
}

// IsGrokModel reports whether the Bedrock model belongs to the xAI Grok family,
// which carries its effort in a reasoning object rather than a thinking one.
func IsGrokModel(model string) bool {
	return strings.Contains(baseModelName(model), "xai.grok")
}

// grokDropsNone are the Grok generations that always reason.
var grokDropsNone = []string{"grok-4.6"}

// GrokCanDisable reports whether the model accepts effort none.
func GrokCanDisable(model string) bool {
	return !containsAny(baseModelName(model), grokDropsNone)
}

// GrokEffort maps a requested effort onto what the model accepts.
func GrokEffort(model, effort string) string {
	switch strings.ToLower(effort) {
	case "none", "":
		if GrokCanDisable(model) {
			return "none"
		}
		return "low"
	case "minimal", "low":
		return "low"
	case "medium":
		return "medium"
	case "xhigh", "max":
		if !GrokCanDisable(model) {
			return "xhigh"
		}
		return "high"
	default:
		return "high"
	}
}
