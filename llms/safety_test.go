package llms_test

import (
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestWithSafetyConfig(t *testing.T) {
	var opts llms.CallOptions

	llms.WithSafetyConfig(map[string]any{
		"identifier": "safe-profile",
		"version":    "v1",
	})(&opts)

	if opts.SafetyConfig == nil {
		t.Fatal("expected safety config to be present")
	}
	if opts.SafetyConfig["identifier"] != "safe-profile" {
		t.Fatalf("identifier = %q, want %q", opts.SafetyConfig["identifier"], "safe-profile")
	}
	if opts.SafetyConfig["version"] != "v1" {
		t.Fatalf("version = %q, want %q", opts.SafetyConfig["version"], "v1")
	}
}

func TestWithSafetyConfig_Nil(t *testing.T) {
	var opts llms.CallOptions

	llms.WithSafetyConfig(nil)(&opts)

	if opts.SafetyConfig != nil {
		t.Fatal("expected safety config to be nil")
	}
}

func TestWithSafetyConfig_Overwrite(t *testing.T) {
	var opts llms.CallOptions

	llms.WithSafetyConfig(map[string]any{
		"identifier": "safe-profile",
		"version":    "v1",
	})(&opts)

	llms.WithSafetyConfig(map[string]any{
		"identifier": "new-profile",
		"version":    "v2",
	})(&opts)

	if opts.SafetyConfig == nil {
		t.Fatal("expected safety config to be present")
	}
	if opts.SafetyConfig["identifier"] != "new-profile" {
		t.Fatalf("identifier = %q, want %q", opts.SafetyConfig["identifier"], "new-profile")
	}
	if opts.SafetyConfig["version"] != "v2" {
		t.Fatalf("version = %q, want %q", opts.SafetyConfig["version"], "v2")
	}
}
