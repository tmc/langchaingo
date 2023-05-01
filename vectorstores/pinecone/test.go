package pinecone

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/pinecone-io/go-pinecone/pinecone_grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func test(indexName, projectName, pineconeEnv, apiKey string) {
	rand.Seed(time.Now().UTC().UnixNano())

	config := &tls.Config{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("content-type", "application/grpc"))

	ctx = metadata.AppendToOutgoingContext(ctx, "api-key", apiKey)
	target := fmt.Sprintf("%s-%s.svc.%s.pinecone.io:443", indexName, projectName, pineconeEnv)
	log.Printf("connecting to %v", target)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(credentials.NewTLS(config)),
		grpc.WithAuthority(target),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := pinecone_grpc.NewVectorServiceClient(conn)

	// upsert
	log.Print("upserting data...")
	upsertResult, upsertErr := client.Upsert(ctx, &pinecone_grpc.UpsertRequest{
		Vectors: []*pinecone_grpc.Vector{
			{
				Id:     "example-vector-1",
				Values: []float32{0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01},
			},
			{
				Id:     "example-vector-2",
				Values: []float32{0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02},
			},
		},
		Namespace: "example-namespace",
	})
	if upsertErr != nil {
		log.Fatalf("upsert error: %v", upsertErr)
	} else {
		log.Printf("upsert result: %v", upsertResult)
	}
}
