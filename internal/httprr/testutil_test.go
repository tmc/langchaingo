package httprr

import (
	"os"
	"testing"
)

func TestGetTestMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected TestMode
	}{
		{
			name:     "record mode",
			envValue: "record",
			expected: TestModeRecord,
		},
		{
			name:     "replay mode",
			envValue: "replay",
			expected: TestModeReplay,
		},
		{
			name:     "disabled mode",
			envValue: "disabled",
			expected: TestModeDisabled,
		},
		{
			name:     "default to replay",
			envValue: "",
			expected: TestModeReplay,
		},
		{
			name:     "invalid mode defaults to replay",
			envValue: "invalid",
			expected: TestModeReplay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("HTTPRR_MODE", tt.envValue)
			} else {
				os.Unsetenv("HTTPRR_MODE")
			}
			defer os.Unsetenv("HTTPRR_MODE")

			result := GetTestMode()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTestClient(t *testing.T) {
	// Test that TestClient returns a valid client
	client := TestClient(t, "test")
	if client == nil {
		t.Fatal("TestClient returned nil")
	}

	// Check that the transport is set correctly based on mode
	os.Setenv("HTTPRR_MODE", "disabled")
	defer os.Unsetenv("HTTPRR_MODE")
	
	disabledClient := TestClient(t, "test")
	if disabledClient.Transport != nil {
		// Default client has nil transport, which means http.DefaultTransport
		t.Log("Client transport is configured")
	}
}