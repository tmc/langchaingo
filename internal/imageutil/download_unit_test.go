package imageutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadImageData(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		wantType   string
		wantData   []byte
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successful PNG download",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				w.Write([]byte{0x89, 0x50, 0x4E, 0x47}) // PNG header
			},
			wantType: "png",
			wantData: []byte{0x89, 0x50, 0x4E, 0x47},
			wantErr:  false,
		},
		{
			name: "successful JPEG download",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // JPEG header
			},
			wantType: "jpeg",
			wantData: []byte{0xFF, 0xD8, 0xFF, 0xE0},
			wantErr:  false,
		},
		{
			name: "successful GIF download",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/gif")
				w.Write([]byte("GIF89a"))
			},
			wantType: "gif",
			wantData: []byte("GIF89a"),
			wantErr:  false,
		},
		{
			name: "invalid mime type - missing slash",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "imagepng")
				w.Write([]byte{0x89, 0x50, 0x4E, 0x47})
			},
			wantErr:    true,
			wantErrMsg: "invalid mime type imagepng",
		},
		{
			name: "invalid mime type - too many parts",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png/extra")
				w.Write([]byte{0x89, 0x50, 0x4E, 0x47})
			},
			wantErr:    true,
			wantErrMsg: "invalid mime type image/png/extra",
		},
		{
			name: "server error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantType: "",
			wantData: []byte{},
			wantErr:  false, // http.Get doesn't return error for non-2xx status
		},
		{
			name: "empty response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "image/png")
				// No data written
			},
			wantType: "png",
			wantData: []byte{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			imageType, data, err := DownloadImageData(server.URL)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantType, imageType)
				assert.Equal(t, tt.wantData, data)
			}
		})
	}
}

func TestDownloadImageData_InvalidURL(t *testing.T) {
	// Test with invalid URL
	_, _, err := DownloadImageData("http://[::1]:99999/invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch image from url")
}
