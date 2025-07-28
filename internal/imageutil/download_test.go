package imageutil

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/0xDezzy/langchaingo/httputil"
	"github.com/0xDezzy/langchaingo/internal/httprr"
)

func requireHttprrRecording(t *testing.T) *httprr.RecordReplay {
	t.Helper()

	// Check if we have httprr recording
	testName := httprr.CleanFileName(t.Name())
	httprrFile := filepath.Join("testdata", testName+".httprr")
	httprrGzFile := httprrFile + ".gz"
	if _, err := os.Stat(httprrFile); os.IsNotExist(err) {
		if _, err := os.Stat(httprrGzFile); os.IsNotExist(err) {
			t.Skip("No httprr recording available for external HTTP calls")
		}
	}

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	return rr
}

func TestDownloadImageData_Integration(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr := requireHttprrRecording(t)
	defer rr.Close()

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() {
		httputil.DefaultClient = oldClient
	}()

	// Test downloading a PNG image
	imageType, data, err := DownloadImageData("https://via.placeholder.com/150/FF0000/FFFFFF?text=Test")
	require.NoError(t, err)
	require.Equal(t, "png", imageType)
	require.NotEmpty(t, data)
}

func TestDownloadImageData_JPEG(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr := requireHttprrRecording(t)
	defer rr.Close()

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() {
		httputil.DefaultClient = oldClient
	}()

	// Test downloading a JPEG image
	imageType, data, err := DownloadImageData("https://via.placeholder.com/150.jpg")
	require.NoError(t, err)
	require.Equal(t, "jpeg", imageType)
	require.NotEmpty(t, data)
}

func TestDownloadImageData_InvalidURL_Integration(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr := requireHttprrRecording(t)
	defer rr.Close()

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() {
		httputil.DefaultClient = oldClient
	}()

	// Test with invalid URL
	_, _, err := DownloadImageData("not-a-valid-url")
	require.Error(t, err)
}

func TestDownloadImageData_NotFound(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr := requireHttprrRecording(t)
	defer rr.Close()

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() {
		httputil.DefaultClient = oldClient
	}()

	// Test with 404 response
	imageType, data, err := DownloadImageData("https://httpbin.org/status/404")
	require.NoError(t, err)               // The function doesn't check status codes
	require.NotEqual(t, "png", imageType) // Likely to be "html" or similar
	require.NotEmpty(t, data)
}

func TestDownloadImageData_InvalidMimeType(t *testing.T) {
	t.Parallel()

	// Setup HTTP record/replay
	rr := requireHttprrRecording(t)
	defer rr.Close()

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() {
		httputil.DefaultClient = oldClient
	}()

	// Test with text content (which should return text/plain or text/html)
	imageType, data, err := DownloadImageData("https://httpbin.org/robots.txt")
	require.NoError(t, err)
	require.Equal(t, "plain", imageType) // text/plain -> plain
	require.NotEmpty(t, data)
}
