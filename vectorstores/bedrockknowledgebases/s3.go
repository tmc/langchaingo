package bedrockknowledgebases

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	eg "golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const _cleanupTimeout = 30 * time.Second

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

// objectKeys returns the keys of the S3 objects that hold docs: the content and, when present, the metadata.
func objectKeys(docs []NamedDocument) []string {
	keys := make([]string, 0, 2*len(docs))
	for _, doc := range docs {
		keys = append(keys, doc.Name)
		if doc.Document.Metadata != nil {
			keys = append(keys, doc.Name+_metadataSuffix)
		}
	}
	return keys
}

// upload uploads docs to the bucket. If an upload fails, it removes the objects already uploaded.
func (kb *KnowledgeBase) upload(ctx context.Context, bucketArn string, docs []NamedDocument) error {
	uploaded, err := kb.addToS3(ctx, bucketArn, docs)
	if err != nil {
		return errors.Join(
			fmt.Errorf("failed to upload documents to S3: %w", err),
			kb.removeFromS3(ctx, bucketArn, uploaded),
		)
	}
	return nil
}

// addToS3 uploads docs to the bucket and returns the keys of the objects uploaded, also when it fails.
func (kb *KnowledgeBase) addToS3(ctx context.Context, bucketArn string, docs []NamedDocument) ([]string, error) {
	sem := semaphore.NewWeighted(kb.s3Config.maxConcurrency)
	uploadDocs, ctx := eg.WithContext(ctx)
	var mu sync.Mutex
	var uploaded []string
	for _, doc := range docs {
		uploadDocs.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("failed to acquire semaphore: %w", err)
			}
			defer sem.Release(1)

			keys, err := kb.uploadS3Object(ctx, bucketArn, doc)
			mu.Lock()
			uploaded = append(uploaded, keys...)
			mu.Unlock()
			if err != nil {
				return fmt.Errorf("failed to upload document: %w", err)
			}
			return nil
		})
	}
	err := uploadDocs.Wait()
	return uploaded, err
}

// removeFromS3 removes the objects with keys from the bucket. It returns an error that names the objects
// it could not remove.
func (kb *KnowledgeBase) removeFromS3(ctx context.Context, bucketArn string, keys []string) error {
	// The removal undoes a failed operation, so it must run also when ctx is the reason of the failure.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), _cleanupTimeout)
	defer cancel()

	bucketName := kb.getBucketName(bucketArn)
	sem := semaphore.NewWeighted(kb.s3Config.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var left []string
	var errs []error
	for _, key := range keys {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := sem.Acquire(ctx, 1)
			if err == nil {
				defer sem.Release(1)
				_, err = kb.s3Config.client.DeleteObject(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String(bucketName),
					Key:    aws.String(key),
				})
			}
			if err != nil {
				mu.Lock()
				left = append(left, key)
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if len(left) > 0 {
		return fmt.Errorf("failed to remove from S3 the objects %s, left in bucket %s: %w", strings.Join(left, ", "), bucketName, errors.Join(errs...))
	}
	return nil
}

// uploadS3Object uploads the content and, when present, the metadata of doc to the bucket. It returns
// the keys of the objects uploaded, also when it fails.
func (kb *KnowledgeBase) uploadS3Object(ctx context.Context, bucketArn string, doc NamedDocument) ([]string, error) {
	bucketName := kb.getBucketName(bucketArn)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(doc.Name),
		Body:   bytes.NewReader([]byte(doc.Document.PageContent)),
	}

	if _, err := kb.s3Config.client.PutObject(ctx, input); err != nil {
		return nil, fmt.Errorf("error uploading file to S3: %w", err)
	}
	uploaded := []string{doc.Name}

	if doc.Document.Metadata != nil {
		bodyBytes, err := json.Marshal(metadata{doc.Document.Metadata})
		if err != nil {
			return uploaded, fmt.Errorf("failed to marshal JSON request body: %w", err)
		}

		metaDataInput := &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(doc.Name + _metadataSuffix),
			Body:   bytes.NewReader(bodyBytes),
		}

		if _, err := kb.s3Config.client.PutObject(ctx, metaDataInput); err != nil {
			return uploaded, fmt.Errorf("error uploading file to S3: %w", err)
		}
		uploaded = append(uploaded, doc.Name+_metadataSuffix)
	}

	return uploaded, nil
}
