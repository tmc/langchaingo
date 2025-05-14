package bedrockknowledgebases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	eg "golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type s3API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type metadata struct {
	MetadataAttributes map[string]any `json:"metadataAttributes"`
}

func (kb *KnowledgeBase) getBucketName(bucketARN string) string {
	const prefix = "arn:aws:s3:::"
	return strings.TrimPrefix(bucketARN, prefix)
}

func (kb *KnowledgeBase) addToS3(ctx context.Context, bucketArn string, docs []NamedDocument) error {
	sem := semaphore.NewWeighted(kb.s3Config.maxConcurrency)
	uploadDocs, ctx := eg.WithContext(ctx)
	for _, doc := range docs {
		uploadDocs.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %w", err)
			}
			defer sem.Release(1)

			if err := kb.uploadS3Object(ctx, bucketArn, doc); err != nil {
				return fmt.Errorf("failed to upload document: %w", err)
			}
			return nil
		})
	}
	return uploadDocs.Wait()
}

func (kb *KnowledgeBase) removeFromS3(ctx context.Context, bucketArn string, docs []NamedDocument) {
	sem := semaphore.NewWeighted(kb.s3Config.maxConcurrency)
	var wg sync.WaitGroup
	for _, doc := range docs {
		wg.Add(1)
		go func(doc NamedDocument) {
			defer wg.Done()
			if err := sem.Acquire(ctx, 1); err != nil {
				// Log error and continue
				fmt.Printf("failed to acquire semaphore: %v\n", err)
				return
			}
			defer sem.Release(1)

			if err := kb.removeS3Object(ctx, bucketArn, doc); err != nil {
				// Log error and continue
				fmt.Printf("failed to remove document: %v\n", err)
			}
		}(doc)
	}
	wg.Wait()
}

// uploadS3 uploads the provided file content to the specified S3 bucket and key.
func (kb *KnowledgeBase) uploadS3Object(ctx context.Context, bucketArn string, doc NamedDocument) error {
	bucketName := kb.getBucketName(bucketArn)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(doc.Name),
		Body:   bytes.NewReader([]byte(doc.Document.PageContent)),
	}

	if _, err := kb.s3Config.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("error uploading file to S3: %w", err)
	}

	if doc.Document.Metadata != nil {
		bodyBytes, err := json.Marshal(metadata{doc.Document.Metadata})
		if err != nil {
			return fmt.Errorf("failed to marshal JSON request body: %w", err)
		}

		metaDataInput := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(doc.Name + _metadataSuffix),
			Body:   bytes.NewReader(bodyBytes),
		}

		if _, err := kb.s3Config.client.PutObject(ctx, metaDataInput); err != nil {
			return fmt.Errorf("error uploading file to S3: %w", err)
		}
	}

	return nil
}

func (kb *KnowledgeBase) removeS3Object(ctx context.Context, bucketArn string, doc NamedDocument) error {
	bucketName := kb.getBucketName(bucketArn)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(doc.Name),
	}

	if _, err := kb.s3Config.client.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("error deleting file from S3: %w", err)
	}

	if doc.Document.Metadata != nil {
		metaDataInput := &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(doc.Name + _metadataSuffix),
		}

		if _, err := kb.s3Config.client.DeleteObject(ctx, metaDataInput); err != nil {
			return fmt.Errorf("error deleting file from S3: %w", err)
		}
	}

	return nil
}
