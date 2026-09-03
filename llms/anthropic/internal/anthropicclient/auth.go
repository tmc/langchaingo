package anthropicclient

import (
	"context"
	"net/http"
)

type authorizer interface {
	authorize(ctx context.Context, req *http.Request) error
}

type staticKey string

func (k staticKey) authorize(_ context.Context, req *http.Request) error {
	req.Header.Set("x-api-key", string(k)) //nolint:canonicalheader
	return nil
}
