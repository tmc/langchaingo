package bedrockclient

// Details keys and values of the content filter errors returned when Bedrock
// blocks content.
const (
	DetailType        = "type"
	DetailCategory    = "category"
	DetailExplanation = "explanation"

	BlockedByGuardrail = "guardrail"
	BlockedByRefusal   = "refusal"
)
