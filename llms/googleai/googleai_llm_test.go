package googleai

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"google.golang.org/api/option"
)

func TestMain(m *testing.M) {
	// This loads the API key if it is a .env file
	_ = godotenv.Load()
	os.Exit(m.Run())
}

func TestGenerateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		parts   []llms.ContentPart
		opts    []llms.CallOption
		wantErr bool
		substr  string
	}{
		{
			name:    "EmptyParts",
			parts:   []llms.ContentPart{},
			opts:    []llms.CallOption{},
			wantErr: true,
			substr:  "",
		},
		{
			name:    "SinglePart",
			parts:   []llms.ContentPart{llms.TextContent{Text: "Say the word Panama and nothing else."}},
			opts:    []llms.CallOption{},
			wantErr: false,
			substr:  "Panama",
		},
		// Following is disabled because the API throws this error:
		// googleapi: Error 400: Image input modality is not enabled for models/gemini-pro
		// {
		//	name: "ImagePart",
		//	parts: []llms.ContentPart{
		//		llms.TextContent{Text: "Does this image contain a cat? Answer with a single word: yes or no."},
		//		llms.ImageURLContent{URL: "https://cdn2.thecatapi.com/images/cej.jpg"}},
		//	opts:    []llms.CallOption{},
		//	wantErr: false,
		//	substr:  "yes",
		// },
	}

	ctx := context.Background()

	opts := []option.ClientOption{}
	apiKey := os.Getenv("GENAI_API_KEY")
	if len(apiKey) > 0 {
		opts = append(opts, option.WithAPIKey(apiKey))
	}

	g, err := New(ctx, opts...)
	if err != nil {
		t.Fatalf("Failed to create GoogleAI: %v", err)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out, err := g.GenerateContent(ctx, tt.parts, tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if len(tt.substr) > 0 {
				require.Contains(t, out.Choices[0].Content, tt.substr)
			}
		})
	}
}
