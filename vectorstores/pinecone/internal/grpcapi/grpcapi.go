package grpcapi

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"github.com/tmc/langchaingo/schema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrEmptyResponse  = errors.New("empty response")
	ErrMissingTextKey = errors.New("missing text key in vector metadata")
)

func GetGRPCConn(
	ctx context.Context,
	indexName,
	projectName,
	environment string,
	apiKey string,
) (*grpc.ClientConn, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	target := fmt.Sprintf(
		"%s-%s.svc.%s.pinecone.io:443",
		indexName,
		projectName,
		environment,
	)

	ctx = metadata.AppendToOutgoingContext(ctx, "api-key", apiKey)

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

	return conn, nil
}

func Upsert(
	ctx context.Context,
	conn *grpc.ClientConn,
	vectors [][]float64,
	metadatas []map[string]any,
	nameSpace string,
) error {
	client := pinecone_grpc.NewVectorServiceClient(conn)

	pineconeVectors := make([]*pinecone_grpc.Vector, 0, len(vectors))
	for i := 0; i < len(vectors); i++ {
		metadataStruct, err := structpb.NewStruct(metadatas[i])
		if err != nil {
			return err
		}

		pineconeVectors = append(
			pineconeVectors,
			&pinecone_grpc.Vector{
				Id:       uuid.New().String(),
				Values:   float64ToFloat32(vectors[i]),
				Metadata: metadataStruct,
			},
		)
	}

	_, err := client.Upsert(ctx, &pinecone_grpc.UpsertRequest{
		Vectors:   pineconeVectors,
		Namespace: nameSpace,
	})

	return err
}

func Query(
	ctx context.Context,
	conn *grpc.ClientConn,
	vector []float64,
	numDocs int,
	nameSpace string,
	textKey string,
) ([]schema.Document, error) {
	client := pinecone_grpc.NewVectorServiceClient(conn)

	queryResult, err := client.Query(ctx, &pinecone_grpc.QueryRequest{
		Queries: []*pinecone_grpc.QueryVector{
			{Values: float64ToFloat32(vector)},
		},
		TopK:          uint32(numDocs),
		IncludeValues: false,
		Namespace:     nameSpace,
	})
	if err != nil {
		return nil, err
	}

	if len(queryResult.Results) == 0 {
		return nil, ErrEmptyResponse
	}

	resultDocuments := make([]schema.Document, 0)
	for _, match := range queryResult.Results[0].Matches {
		metadata := match.Metadata.AsMap()

		pageContent, ok := metadata[textKey].(string)
		if !ok {
			return nil, ErrMissingTextKey
		}
		delete(metadata, textKey)

		resultDocuments = append(resultDocuments, schema.Document{
			PageContent: pageContent,
			Metadata:    metadata,
		})
	}

	return resultDocuments, nil
}

func float64ToFloat32(input []float64) []float32 {
	output := make([]float32, len(input))
	for i, v := range input {
		output[i] = float32(v)
	}
	return output
}
