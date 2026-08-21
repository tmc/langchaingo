package llms

import (
	"fmt"
	"strings"
)

var truncationStopReasons = map[string]bool{
	"length":     true,
	"max_tokens": true,
}

// IsTruncated reports whether a vendor stop reason means generation stopped at
// the output token limit rather than finishing. Matching is case-insensitive.
func IsTruncated(stopReason string) bool {
	return truncationStopReasons[strings.ToLower(strings.TrimSpace(stopReason))]
}

// CheckTruncation reports a truncated answer as an error, but only for a caller
// that opted in with [WithFailOnTruncation]; otherwise a truncated answer stays
// a successful response carrying [ContentChoice.Truncated].
func CheckTruncation(resp *ContentResponse, opts CallOptions) error {
	if !opts.FailOnTruncation || resp == nil {
		return nil
	}
	for i, choice := range resp.Choices {
		if choice == nil || !choice.Truncated {
			continue
		}
		return NewError(ErrCodeTruncated, "", fmt.Sprintf(
			"choice %d stopped at the output token limit (stop reason %q)", i, choice.StopReason))
	}
	return nil
}
