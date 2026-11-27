package bedrock

import (
	"time"

	"github.com/tmc/langchaingo/llms"
)

// EphemeralCache returns an ephemeral cache control for Bedrock prompt
// caching with no ttl field on the wire. Bedrock applies its default
// 5-minute behavior.
func EphemeralCache() *llms.CacheControl {
	return &llms.CacheControl{Type: "ephemeral"}
}

// EphemeralCacheOneHour returns an ephemeral cache control with ttl=1h.
func EphemeralCacheOneHour() *llms.CacheControl {
	return &llms.CacheControl{Type: "ephemeral", Duration: time.Hour}
}
