package cloudflareclient

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestClient_GenerateContent(t *testing.T) { // nolint:funlen
	t.Parallel()

	type fields struct {
		httpClient         httpClient
		accountID          string
		token              string
		baseURL            string
		modelName          string
		embeddingModelName string
		endpointURL        string
		bearerToken        string
	}
	type args struct {
		ctx     context.Context // nolint:containedctx
		request *GenerateContentRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GenerateContentResponse
		wantErr bool
	}{
		{
			name: "test generate content success",
			fields: fields{
				httpClient: &mockHTTPClient{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"result": {"response": "response"}}`)),
					},
				},
				accountID:          "accountID",
				token:              "token",
				baseURL:            "baseURL",
				modelName:          "modelName",
				embeddingModelName: "embeddingModelName",
			},
			args: args{
				ctx: t.Context(),
				request: &GenerateContentRequest{
					Messages: []Message{
						{Role: "system", Content: "systemPrompt"},
						{Role: "user", Content: "userPrompt"},
					},
				},
			},
			want: &GenerateContentResponse{
				Result: struct {
					Response string `json:"response"`
				}{
					Response: "response",
				},
			},
		},
		{
			name: "test generate content stream success",
			fields: fields{
				httpClient: &mockHTTPClient{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"result": {"response": "response"}}`)),
					},
				},
				accountID:          "accountID",
				token:              "token",
				baseURL:            "baseURL",
				modelName:          "modelName",
				embeddingModelName: "embeddingModelName",
			},
			args: args{
				ctx: t.Context(),
				request: &GenerateContentRequest{
					Messages: []Message{
						{Role: "system", Content: "systemPrompt"},
						{Role: "user", Content: "userPrompt"},
					},
					Stream: true,
					StreamingFunc: func(_ context.Context, chunk []byte) error {
						if string(chunk) != `{"result": {"response": "response"}}` {
							return io.EOF
						}
						return nil
					},
				},
			},
			want: &GenerateContentResponse{
				Result: struct {
					Response string `json:"response"`
				}{
					Response: "",
				},
			},
		},
	}
	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := Client{
				httpClient:         tt.fields.httpClient,
				accountID:          tt.fields.accountID,
				token:              tt.fields.token,
				baseURL:            tt.fields.baseURL,
				modelName:          tt.fields.modelName,
				embeddingModelName: tt.fields.embeddingModelName,
				endpointURL:        tt.fields.endpointURL,
				bearerToken:        tt.fields.bearerToken,
			}

			got, err := c.GenerateContent(t.Context(), tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateContent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
