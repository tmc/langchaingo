package httprr

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// TestMode determines how httprr behaves in tests.
type TestMode string

const (
	// TestModeRecord records real HTTP interactions.
	TestModeRecord TestMode = "record"
	// TestModeReplay replays recorded HTTP interactions.
	TestModeReplay TestMode = "replay"
	// TestModeDisabled disables httprr (passes through to real HTTP).
	TestModeDisabled TestMode = "disabled"
)

// GetTestMode returns the test mode from environment variables.
// Defaults to replay mode for deterministic tests.
func GetTestMode() TestMode {
	mode := os.Getenv("HTTPRR_MODE")
	switch mode {
	case "record":
		return TestModeRecord
	case "replay":
		return TestModeReplay
	case "disabled":
		return TestModeDisabled
	default:
		return TestModeReplay
	}
}

// TestClient creates an HTTP client configured for testing with httprr.
// It automatically determines the cassette path based on the test name.
func TestClient(t *testing.T, cassetteName string) *http.Client {
	t.Helper()
	
	mode := GetTestMode()
	if mode == TestModeDisabled {
		return http.DefaultClient
	}
	
	// Create cassette directory
	cassetteDir := filepath.Join("testdata", "cassettes")
	if err := os.MkdirAll(cassetteDir, 0755); err != nil {
		t.Fatalf("Failed to create cassette directory: %v", err)
	}
	
	cassettePath := filepath.Join(cassetteDir, cassetteName+".json")
	
	var httprMode Mode
	switch mode {
	case TestModeRecord:
		httprMode = ModeRecord
	case TestModeReplay:
		httprMode = ModeReplay
	default:
		httprMode = ModeDisabled
	}
	
	return Client(cassettePath, httprMode)
}

// SkipIfNotRecording skips the test if not in recording mode.
// Useful for tests that need to make real HTTP calls.
func SkipIfNotRecording(t *testing.T) {
	t.Helper()
	if GetTestMode() != TestModeRecord {
		t.Skip("Skipping test - not in recording mode")
	}
}

// SkipIfRecording skips the test if in recording mode.
// Useful for tests that should only run with recorded data.
func SkipIfRecording(t *testing.T) {
	t.Helper()
	if GetTestMode() == TestModeRecord {
		t.Skip("Skipping test - in recording mode")
	}
}