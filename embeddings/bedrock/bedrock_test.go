package bedrock_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/vxcontrol/langchaingo/embeddings/bedrock"
	"github.com/vxcontrol/langchaingo/internal/httprr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/stretchr/testify/require"
)

func setUpTestWithTransport(rr *httprr.RecordReplay) (*bedrockruntime.Client, error) {
	// Configure request scrubbing to remove dynamic AWS headers
	rr.ScrubReq(func(req *http.Request) error {
		req.Header.Del("Amz-Sdk-Invocation-Id")
		req.Header.Del("Amz-Sdk-Request")
		req.Header.Del("X-Amz-Date")
		// Scrub the actual AWS signature to make it reproducible
		if auth := req.Header.Get("Authorization"); auth != "" {
			req.Header.Set("Authorization", "AWS4-HMAC-SHA256 test-api-key")
		}
		return nil
	})

	httpClient := &http.Client{
		Transport: rr,
	}

	cfgOpts := []func(*config.LoadOptions) error{
		config.WithHTTPClient(httpClient),
	}

	// When replaying, provide fake credentials to avoid IMDS calls
	if !rr.Recording() {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(
			&fakeCredentialsProvider{},
		))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, err
	}

	client := bedrockruntime.NewFromConfig(cfg)
	return client, nil
}

// fakeCredentialsProvider provides fake AWS credentials for replay mode
type fakeCredentialsProvider struct{}

func (f *fakeCredentialsProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Source:          "FakeCredentialsProvider",
	}, nil
}

func TestEmbedQuery(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	require.NoError(t, err)

	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelCohereEn), bedrock.WithClient(client))
	require.NoError(t, err)
	_, err = model.EmbedQuery(ctx, "hello world")

	require.NoError(t, err)
}

func TestEmbedDocuments(t *testing.T) {
	ctx := t.Context()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Configure AWS client to use httprr transport
	client, err := setUpTestWithTransport(rr)
	require.NoError(t, err)

	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelCohereEn), bedrock.WithClient(client))
	require.NoError(t, err)

	embeddings, err := model.EmbedDocuments(ctx, []string{"hello world", "goodbye world"})

	require.NoError(t, err)
	require.Len(t, embeddings, 2)
}
