package httputil

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserAgent(t *testing.T) {
	ua := UserAgent()

	// Check that User-Agent is not empty
	assert.NotEmpty(t, ua)

	// Check that it contains expected components
	assert.Contains(t, ua, "langchaingo/")
	assert.Contains(t, ua, "Go/"+runtime.Version())
	assert.Contains(t, ua, runtime.GOOS)
	assert.Contains(t, ua, runtime.GOARCH)

	// Verify format includes parentheses for OS/arch
	assert.Contains(t, ua, "("+runtime.GOOS+" "+runtime.GOARCH+")")

	// Verify it doesn't have extra spaces
	assert.NotContains(t, ua, "  ")

	// Check that subsequent calls return the same value (cached)
	ua2 := UserAgent()
	assert.Equal(t, ua, ua2)
}

func TestUserAgentFormat(t *testing.T) {
	ua := UserAgent()

	// Check that User-Agent contains expected format
	assert.Contains(t, ua, "langchaingo/")
	assert.Contains(t, ua, "Go/")

	// Check Go version format
	assert.Regexp(t, `Go/go\d+\.\d+(\.\d+)?`, ua, "Should contain valid Go version")

	// Check OS/arch format - should end with "(OS ARCH)"
	assert.Regexp(t, `\([a-z0-9]+ [a-z0-9_]+\)$`, ua, "Should end with '(OS ARCH)' format")
}

func TestUserAgentComponents(t *testing.T) {
	tests := []struct {
		name     string
		contains []string
	}{
		{
			name: "contains langchaingo version",
			contains: []string{
				"langchaingo/",
			},
		},
		{
			name: "contains Go version",
			contains: []string{
				"Go/" + runtime.Version(),
			},
		},
		{
			name: "contains OS and architecture",
			contains: []string{
				runtime.GOOS,
				runtime.GOARCH,
			},
		},
	}

	ua := UserAgent()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, substr := range tt.contains {
				assert.Contains(t, ua, substr)
			}
		})
	}
}
