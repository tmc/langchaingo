package pinecone

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type clientOption struct {
	indexName   string
	projectName string
	environment string
	apiKey      string
}

// Options is a function type that can be used to modify the AgentOption.
type Options func(p *clientOption)

func defaultOptions() clientOption {
	return clientOption{
		"indexName": false,
	}
}

type Client struct {
}

func New(ctx context.Context, opt []Option) (*Client, error) {
	target := fmt.Sprintf(
		"%s-%s.svc.%s.pinecone.io:443",
		indexName,
		projectName,
		pineconeEnv,
	)

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithAuthority(target),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
}
