package internal

// NoCredentialsError is thrown when no valid credentials are passed to the client.
type NoCredentialsError struct{}

func (e NoCredentialsError) Error() string {
	return "Must pass a APIKey or AccessToken"
}
