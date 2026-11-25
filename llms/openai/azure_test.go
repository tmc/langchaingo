package openai

import (
	"context"
	"errors"
	"testing"
)

// clearAzureEnv keeps the ambient environment from filling in what a case
// deliberately leaves out
func clearAzureEnv(t *testing.T) {
	t.Helper()
	t.Setenv(tokenEnvVarName, "")
	t.Setenv(modelEnvVarName, "")
	t.Setenv(baseURLEnvVarName, "")
	t.Setenv(baseAPIBaseEnvVarName, "")
}

// chat completions do not touch the embeddings deployment, requiring one kept
// Azure users from creating a chat client at all
func TestAzureChatWithoutEmbeddingModel(t *testing.T) {
	clearAzureEnv(t)

	llm, err := New(
		WithAPIType(APITypeAzure),
		WithToken("token"),
		WithModel("gpt-4o-deployment"),
		WithBaseURL("https://resource.openai.azure.com"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if llm == nil {
		t.Fatal("expected a client")
	}
}

func TestAzureMissingModel(t *testing.T) {
	clearAzureEnv(t)

	for name, opts := range map[string][]Option{
		"default api version": {
			WithAPIType(APITypeAzure),
			WithToken("token"),
		},
		// the check used to sit behind an empty api version and was skipped
		// whenever one was passed
		"explicit api version": {
			WithAPIType(APITypeAzure),
			WithToken("token"),
			WithAPIVersion("2024-10-21"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := New(opts...); !errors.Is(err, ErrMissingAzureModel) {
				t.Errorf("expected ErrMissingAzureModel, got %v", err)
			}
		})
	}
}

func TestAzureAPIVersion(t *testing.T) {
	clearAzureEnv(t)

	for name, tc := range map[string]struct {
		opts []Option
		want string
	}{
		"default":  {[]Option{}, DefaultAPIVersion},
		"explicit": {[]Option{WithAPIVersion("2024-10-21")}, "2024-10-21"},
	} {
		t.Run(name, func(t *testing.T) {
			opts := append([]Option{
				WithAPIType(APITypeAzure),
				WithToken("token"),
				WithModel("gpt-4o-deployment"),
			}, tc.opts...)

			options, _, err := newClient(opts...)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if options.apiVersion != tc.want {
				t.Errorf("expected api version %q, got %q", tc.want, options.apiVersion)
			}
		})
	}
}

// the embeddings deployment is part of the request path, so it is reported
// where it is needed rather than at construction
func TestAzureEmbeddingWithoutEmbeddingModel(t *testing.T) {
	clearAzureEnv(t)

	llm, err := New(
		WithAPIType(APITypeAzure),
		WithToken("token"),
		WithModel("gpt-4o-deployment"),
		WithBaseURL("https://resource.openai.azure.com"),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := llm.CreateEmbedding(context.Background(), []string{"hello"}); !errors.Is(err, ErrMissingAzureEmbeddingModel) {
		t.Errorf("expected ErrMissingAzureEmbeddingModel, got %v", err)
	}
}
