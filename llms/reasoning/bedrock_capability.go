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

// NovaClearsInferenceConfigAt reports whether Nova refuses temperature, topP and
// maxTokens alongside the given effort.
func NovaClearsInferenceConfigAt(effort string) bool {
	return NovaEffort(effort) == "high"
}

// IsBedrockAlwaysReasoningModel reports whether the Bedrock family reasons on
// every request and takes no thinking configuration alongside it.
func IsBedrockAlwaysReasoningModel(model string) bool {
	return strings.Contains(baseModelName(model), "deepseek.r1")
}

// IsGrokModel reports whether the Bedrock model belongs to the xAI Grok family,
// which carries its effort in a reasoning object rather than a thinking one.
func IsGrokModel(model string) bool {
	return strings.Contains(baseModelName(model), "xai.grok")
}

// GrokEffort maps a requested effort onto what the model accepts. Only the
// generations that kept none can carry it; a request to disable a model that
// always reasons leaves the field off the wire rather than inventing a level.
func GrokEffort(model, effort string) string {
	switch strings.ToLower(effort) {
	case "none":
		if mandatoryThinking(model) {
			return ""
		}
		return "none"
	case "minimal", "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	case "xhigh", "max":
		return "xhigh"
	default:
		return ""
	}
}
