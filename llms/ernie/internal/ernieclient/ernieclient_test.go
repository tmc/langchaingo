package ernieclient

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockHttpClient struct{}

// implement ernieclient.Doer interface
func (m *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	authResponse := &authResponse{
		AccessToken: "test",
	}
	b, err := json.Marshal(authResponse)
	if err != nil {
		return nil, err
	}
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}
	return response, nil
}

func TestClient_buildURL(t *testing.T) {
	t.Parallel()
	type fields struct {
		apiKey      string
		secretKey   string
		accessToken string
		httpClient  Doer
	}
	type args struct {
		modelpath ModelPath
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "with access token",
			fields: fields{
				accessToken: "token",
			},
			args: args{modelpath: "eb-instant"},
			want: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant?access_token=token",
		},
		{
			name: "with client, aksk",
			fields: fields{
				apiKey:     "test",
				secretKey:  "test",
				httpClient: &mockHttpClient{},
			},
			args: args{modelpath: "eb-instant"},
			want: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant?access_token=test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c, err := New(
				WithAKSK(tt.fields.apiKey, tt.fields.secretKey),
				WithAccessToken(tt.fields.accessToken),
				WithHTTPClient(tt.fields.httpClient),
			)
			if err != nil {
				t.Errorf("New got error. %v", err)
			}
			if got := c.buildURL(tt.args.modelpath); got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
