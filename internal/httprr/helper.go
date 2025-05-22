package httprr

import (
	"net/http"
	"os"
	"path/filepath"
)

// NewTestClient creates an HTTP client with httprr recording capabilities.
// It automatically detects if it should record or replay based on environment variables.
func NewTestClient(testName string) *http.Client {
	mode := getRecordingMode()
	recordingsDir := getRecordingsDir(testName)
	
	recorder := New(recordingsDir, mode)
	return recorder.Client()
}

// NewTestClientWithRecorder creates an HTTP client with a custom httprr recorder.
func NewTestClientWithRecorder(recorder *Recorder) *http.Client {
	return recorder.Client()
}

// WrapHTTPClient wraps an existing HTTP client with httprr recording.
func WrapHTTPClient(client *http.Client, testName string) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	
	mode := getRecordingMode()
	recordingsDir := getRecordingsDir(testName)
	
	recorder := New(recordingsDir, mode)
	client.Transport = recorder.Transport(client.Transport)
	
	return client
}

// getRecordingMode determines the recording mode based on environment variables.
func getRecordingMode() Mode {
	switch os.Getenv("HTTPRR_MODE") {
	case "record":
		return ModeRecord
	case "replay":
		return ModeReplay
	case "record_once":
		return ModeRecordOnce
	default:
		// Default to record_once for tests
		return ModeRecordOnce
	}
}

// getRecordingsDir determines the recordings directory for a test.
func getRecordingsDir(testName string) string {
	baseDir := os.Getenv("HTTPRR_RECORDINGS_DIR")
	if baseDir == "" {
		baseDir = "testdata/recordings"
	}
	
	return filepath.Join(baseDir, testName)
}

// TestClientOption is a function type for configuring test clients.
type TestClientOption func(*TestClientConfig)

// TestClientConfig holds configuration for test clients.
type TestClientConfig struct {
	RecordingsDir string
	Mode          Mode
	TestName      string
}

// WithRecordingsDir sets the recordings directory.
func WithRecordingsDir(dir string) TestClientOption {
	return func(c *TestClientConfig) {
		c.RecordingsDir = dir
	}
}

// WithMode sets the recording mode.
func WithMode(mode Mode) TestClientOption {
	return func(c *TestClientConfig) {
		c.Mode = mode
	}
}

// NewTestClientWithOptions creates a test client with custom options.
func NewTestClientWithOptions(testName string, opts ...TestClientOption) *http.Client {
	config := &TestClientConfig{
		TestName:      testName,
		RecordingsDir: getRecordingsDir(testName),
		Mode:          getRecordingMode(),
	}
	
	for _, opt := range opts {
		opt(config)
	}
	
	recorder := New(config.RecordingsDir, config.Mode)
	return recorder.Client()
}