package httprr

import (
	"net/http"
	"sync"
)

var (
	// originalDefaultClient stores the original http.DefaultClient
	originalDefaultClient *http.Client

	// mutex to protect state
	replaceMutex sync.Mutex

	// replaced indicates whether the default client has been replaced
	replaced bool
)

// ReplaceDefaultClient replaces the global http.DefaultClient with a recording client
// and returns a function to restore the original.
func ReplaceDefaultClient() *Recorder {
	replaceMutex.Lock()
	defer replaceMutex.Unlock()

	if replaced {
		panic("httprr: http.DefaultClient already replaced. Call Restore() first.")
	}

	// Save the original client
	originalDefaultClient = http.DefaultClient

	// Create a new recorder using the original transport
	recorder := NewRecorder(originalDefaultClient.Transport)

	// Create a new client with the recorder
	client := &http.Client{
		Transport:     recorder,
		CheckRedirect: originalDefaultClient.CheckRedirect,
		Jar:           originalDefaultClient.Jar,
		Timeout:       originalDefaultClient.Timeout,
	}

	// Replace the default client
	http.DefaultClient = client

	replaced = true
	return recorder
}

// RestoreDefaultClient restores the original http.DefaultClient
func RestoreDefaultClient() {
	replaceMutex.Lock()
	defer replaceMutex.Unlock()

	if !replaced {
		return
	}

	// Restore the original client
	http.DefaultClient = originalDefaultClient
	replaced = false
} 