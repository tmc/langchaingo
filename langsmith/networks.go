package langsmith

import (
	"net/url"
)

func isLocalhost(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	hostname := u.Hostname()
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1"
}
