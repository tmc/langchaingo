package openaiclient

import (
	"strings"
	"testing"
)

func TestStatusErrorSurfacesProviderText(t *testing.T) {
	cases := []struct {
		name, body, want string
	}{
		{"openai shape", `{"error":{"message":"bad request","type":"invalid"}}`, "bad request"},
		{"xai shape", `{"code":"invalid-argument","error":"Model grok-4.5 does not support parameter presencePenalty."}`, "presencePenalty"},
		{"plain text", "upstream exploded", "upstream exploded"},
		{"empty body", "", "400"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := statusError(400, strings.NewReader(tc.body))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestStatusErrorCapsTheQuotedBody(t *testing.T) {
	err := statusError(500, strings.NewReader(strings.Repeat("x", maxErrorBodyBytes*4)))
	if len(err.Error()) > maxErrorBodyBytes+128 {
		t.Fatalf("error grew to %d bytes, want it capped near %d", len(err.Error()), maxErrorBodyBytes)
	}
}

func TestStatusErrorParsesABodyLargerThanTwoKilobytes(t *testing.T) {
	padding := strings.Repeat("d", 8<<10)
	body := `{"error":{"message":"content filter triggered","type":"invalid_request_error","detail":"` +
		padding + `"}}`

	err := statusError(400, strings.NewReader(body))
	if err == nil {
		t.Fatal("want an error")
	}
	got := err.Error()
	if !strings.Contains(got, "content filter triggered") {
		t.Errorf("provider message lost: %s", got)
	}
	if strings.Contains(got, `{"error"`) {
		t.Errorf("body was quoted raw instead of parsed: %.120s", got)
	}
	if strings.Contains(got, padding[:64]) {
		t.Errorf("padding leaked into the error: %.120s", got)
	}
}
