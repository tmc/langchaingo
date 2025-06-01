package httprr

import (
	"bufio"
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Unit tests that don't require external dependencies

func TestCleanFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple test name",
			input:    "TestMyFunction",
			expected: "TestMyFunction",
		},
		{
			name:     "test with subtest",
			input:    "TestMyFunction/subtest",
			expected: "TestMyFunction-subtest",
		},
		{
			name:     "complex test name",
			input:    "Test API/Complex_Case",
			expected: "Test-API-Complex_Case",
		},
		{
			name:     "test with multiple separators",
			input:    "Test\\Function:With*Special?Characters",
			expected: "Test-Function-With-Special-Characters",
		},
		{
			name:     "test with quotes and brackets",
			input:    "Test\"Function<With>Special|Characters",
			expected: "Test-Function-With-Special-Characters",
		},
		{
			name:     "test with spaces",
			input:    "Test Function With Spaces",
			expected: "Test-Function-With-Spaces",
		},
		{
			name:     "test with multiple consecutive separators",
			input:    "Test///Function",
			expected: "Test-Function",
		},
		{
			name:     "test with leading and trailing separators",
			input:    "/Test Function/",
			expected: "Test-Function",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only separators",
			input:    "///",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBodyReadClose(t *testing.T) {
	t.Parallel()

	data := []byte("test data")
	body := &Body{Data: data}

	// Test reading
	buffer := make([]byte, 5)
	n, err := body.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte("test "), buffer)
	assert.Equal(t, 5, body.ReadOffset)

	// Test reading remainder
	buffer = make([]byte, 10)
	n, err = body.Read(buffer)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("data"), buffer[:n])
	assert.Equal(t, 9, body.ReadOffset)

	// Test EOF
	n, err = body.Read(buffer)
	assert.Error(t, err)
	assert.Equal(t, 0, n)

	// Test Close (should be no-op)
	err = body.Close()
	assert.NoError(t, err)
}

func TestRecordReplayBasics(t *testing.T) {
	t.Parallel()

	// Test basic struct initialization
	rr := &RecordReplay{
		file:   "test.httprr",
		replay: make(map[string]string),
	}

	assert.Equal(t, "test.httprr", rr.file)
	assert.NotNil(t, rr.replay)
	assert.False(t, rr.Recording())

	// Test with record mode
	rr.record = &os.File{}
	assert.True(t, rr.Recording())
}

func TestScrubReqResp(t *testing.T) {
	t.Parallel()

	rr := &RecordReplay{}

	// Test adding request scrubbers
	scrub1 := func(req *http.Request) error { return nil }
	scrub2 := func(req *http.Request) error { return nil }

	rr.ScrubReq(scrub1, scrub2)
	assert.Len(t, rr.reqScrub, 2)

	// Test adding response scrubbers
	respScrub1 := func(buf *bytes.Buffer) error { return nil }
	respScrub2 := func(buf *bytes.Buffer) error { return nil }

	rr.ScrubResp(respScrub1, respScrub2)
	assert.Len(t, rr.respScrub, 2)

	// Test cumulative addition
	rr.ScrubReq(scrub1)
	assert.Len(t, rr.reqScrub, 3)

	rr.ScrubResp(respScrub1)
	assert.Len(t, rr.respScrub, 3)
}

func TestRecordingFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		flagValue string
		filename  string
		expected  bool
		expectErr bool
	}{
		{
			name:      "empty flag",
			flagValue: "",
			filename:  "test.httprr",
			expected:  false,
		},
		{
			name:      "exact match",
			flagValue: "test.httprr",
			filename:  "test.httprr",
			expected:  true,
		},
		{
			name:      "regex match",
			flagValue: ".*\\.httprr",
			filename:  "test.httprr",
			expected:  true,
		},
		{
			name:      "no match",
			flagValue: "other.httprr",
			filename:  "test.httprr",
			expected:  false,
		},
		{
			name:      "invalid regex",
			flagValue: "[invalid",
			filename:  "test.httprr",
			expected:  false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the recording flag
			restore := setRecordForTesting(tt.flagValue)
			defer restore()

			result, err := Recording(tt.filename)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	t.Parallel()

	rr := &RecordReplay{}

	// Initially no error
	err := rr.writeError()
	assert.NoError(t, err)

	// Set write error
	testErr := assert.AnError
	rr.writeErr = testErr

	err = rr.writeError()
	assert.Equal(t, testErr, err)
}

func TestClient(t *testing.T) {
	t.Parallel()

	rr := &RecordReplay{}
	client := rr.Client()

	assert.NotNil(t, client)
	assert.Equal(t, rr, client.Transport)
}

func TestDefaultRequestScrubbers(t *testing.T) {
	t.Parallel()

	scrubbers := getDefaultRequestScrubbers()
	assert.Len(t, scrubbers, 1)

	// Test the scrubber functionality
	req, err := http.NewRequest("GET", "https://api.example.com", nil)
	require.NoError(t, err)

	// Add headers that should be scrubbed
	req.Header.Set("API-Key", "secret123")
	req.Header.Set("Authorization", "Bearer secrettoken")
	req.Header.Set("X-API-Token", "anothersecret")
	req.Header.Set("Openai-Organization", "org-123")
	req.Header.Set("User-Agent", "custom-agent")

	// Apply scrubber
	err = scrubbers[0](req)
	assert.NoError(t, err)

	// Verify scrubbing
	assert.Equal(t, "test-api-key", req.Header.Get("API-Key"))
	assert.Equal(t, "Bearer test-api-key", req.Header.Get("Authorization"))
	assert.Equal(t, "test-api-key", req.Header.Get("X-API-Token"))
	assert.Equal(t, "lcgo-tst", req.Header.Get("Openai-Organization"))
	assert.Equal(t, "langchaingo-httprr", req.Header.Get("User-Agent"))
}

func TestDefaultResponseScrubbers(t *testing.T) {
	t.Parallel()

	scrubbers := getDefaultResponseScrubbers()
	assert.Len(t, scrubbers, 1)

	// Create a test HTTP response
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}
	resp.Header.Set("Cf-Ray", "123456789-ABC")
	resp.Header.Set("Set-Cookie", "session=secret")
	resp.Header.Set("Openai-Organization", "org-123")

	// Serialize response
	var buf bytes.Buffer
	err := resp.Write(&buf)
	require.NoError(t, err)

	// Apply scrubber
	err = scrubbers[0](&buf)
	assert.NoError(t, err)

	// Parse the scrubbed response
	scrubbedResp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(buf.Bytes())), nil)
	require.NoError(t, err)

	// Verify scrubbing
	assert.Empty(t, scrubbedResp.Header.Get("Cf-Ray"))
	assert.Empty(t, scrubbedResp.Header.Get("Set-Cookie"))
	assert.Equal(t, "lcgo-tst", scrubbedResp.Header.Get("Openai-Organization"))
}

func TestEmbeddingJSONFormatter(t *testing.T) {
	t.Parallel()

	formatter := EmbeddingJSONFormatter()

	// Test with a simple buffer that doesn't include HTTP headers
	t.Run("simple test", func(t *testing.T) {
		buf := &bytes.Buffer{}
		err := formatter(buf)
		assert.NoError(t, err)
	})

	t.Run("formatter function creation", func(t *testing.T) {
		// Just test that we can create the formatter
		formatter := EmbeddingJSONFormatter()
		assert.NotNil(t, formatter)
	})
}

func TestFormatJSONValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "integer value",
			input:    42.0,
			expected: "42",
		},
		{
			name:     "float value",
			input:    3.14,
			expected: "3.14",
		},
		{
			name:     "boolean value",
			input:    true,
			expected: "true",
		},
		{
			name:     "null value",
			input:    nil,
			expected: "null",
		},
		{
			name:     "empty object",
			input:    map[string]interface{}{},
			expected: "{}",
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: "[]",
		},
		{
			name:     "number array",
			input:    []interface{}{1.0, 2.0, 3.5},
			expected: "[1, 2, 3.5]",
		},
		{
			name:  "simple object",
			input: map[string]interface{}{"key": "value"},
			expected: `{
  "key": "value"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJSONValue(tt.input, 0)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatJSONBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "non-JSON input",
			input:    []byte("plain text"),
			expected: "plain text",
		},
		{
			name:     "invalid JSON",
			input:    []byte(`{"invalid": json}`),
			expected: `{"invalid": json}`,
		},
		{
			name:  "valid JSON",
			input: []byte(`{"key": "value"}`),
			expected: `{
  "key": "value"
}`,
		},
		{
			name:     "JSON array",
			input:    []byte(`[1, 2, 3]`),
			expected: "[1, 2, 3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatJSONBody(tt.input)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestHasRequiredCredentials(t *testing.T) {
	t.Parallel()

	// Set some environment variables for testing
	originalValue := os.Getenv("TEST_CRED")
	defer func() {
		if originalValue == "" {
			os.Unsetenv("TEST_CRED")
		} else {
			os.Setenv("TEST_CRED", originalValue)
		}
	}()

	os.Setenv("TEST_CRED", "value")

	tests := []struct {
		name     string
		envVars  []string
		expected bool
	}{
		{
			name:     "no env vars",
			envVars:  []string{},
			expected: false,
		},
		{
			name:     "existing env var",
			envVars:  []string{"TEST_CRED"},
			expected: true,
		},
		{
			name:     "non-existing env var",
			envVars:  []string{"NON_EXISTING_VAR"},
			expected: false,
		},
		{
			name:     "mixed env vars",
			envVars:  []string{"NON_EXISTING_VAR", "TEST_CRED"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRequiredCredentials(tt.envVars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()

	// Test that flags are properly initialized
	assert.NotNil(t, record)
	assert.NotNil(t, debug)
	assert.NotNil(t, httpDebug)
	assert.NotNil(t, recordDelay)
}

func TestTestWriter(t *testing.T) {
	t.Parallel()

	// Create a mock test for the writer
	writer := testWriter{t}

	data := []byte("test data")
	n, err := writer.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
}

func TestSetRecordForTesting(t *testing.T) {
	t.Parallel()

	// Get original value
	originalValue := *record

	// Set test value
	restore := setRecordForTesting("test-value")
	assert.Equal(t, "test-value", *record)

	// Restore original value
	restore()
	assert.Equal(t, originalValue, *record)
}

func TestRecordReplayClose(t *testing.T) {
	t.Parallel()

	t.Run("close with no record file", func(t *testing.T) {
		rr := &RecordReplay{}
		err := rr.Close()
		assert.NoError(t, err)
	})

	t.Run("close with write error", func(t *testing.T) {
		rr := &RecordReplay{
			writeErr: assert.AnError,
		}
		err := rr.Close()
		assert.Equal(t, assert.AnError, err)
	})
}

func TestInternalStructs(t *testing.T) {
	t.Parallel()

	// Test basic Body struct functionality
	body := &Body{
		Data:       []byte("test data"),
		ReadOffset: 0,
	}

	assert.Equal(t, []byte("test data"), body.Data)
	assert.Equal(t, 0, body.ReadOffset)
}
