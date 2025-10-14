package bedrock_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vendasta/langchaingo/embeddings/bedrock"
	"github.com/vendasta/langchaingo/httputil"
	"github.com/vendasta/langchaingo/internal/httprr"
)

func TestEmbedQuery(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelTitanEmbedG1))
	require.NoError(t, err)
	_, err = model.EmbedQuery(ctx, "hello world")

	require.NoError(t, err)
}

func TestEmbedDocuments(t *testing.T) {
	ctx := context.Background()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "AWS_ACCESS_KEY_ID")

	rr := httprr.OpenForTest(t, http.DefaultTransport)
	defer rr.Close()

	// Only run tests in parallel when not recording (to avoid rate limits)
	if !rr.Recording() {
		t.Parallel()
	}

	// Replace httputil.DefaultClient with httprr client
	oldClient := httputil.DefaultClient
	httputil.DefaultClient = rr.Client()
	defer func() { httputil.DefaultClient = oldClient }()

	model, err := bedrock.NewBedrock(bedrock.WithModel(bedrock.ModelCohereEn))
	require.NoError(t, err)

	embeddings, err := model.EmbedDocuments(ctx, []string{"hello world", "goodbye world"})

	require.NoError(t, err)
	require.Len(t, embeddings, 2)
}
