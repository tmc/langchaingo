package internal

// ErrNoCredentials is thrown when no valid credentials are passed to the client.
type ErrNoCredentials struct{}

func (e ErrNoCredentials) Error() string {
	return "Must pass a APIKey or AccessToken"
}
