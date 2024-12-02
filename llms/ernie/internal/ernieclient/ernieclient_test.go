package ernieclient

import "testing"

func TestClient_buildURL(t *testing.T) {
	t.Parallel()
	type fields struct {
		apiKey      string
		secretKey   string
		accessToken string
		httpClient  Doer
		Model       string
		ModelPath   ModelPath
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
			name: "one",
			fields: fields{
				accessToken: "token",
			},
			args: args{modelpath: "eb-instant"},
			want: "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/eb-instant?access_token=token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				apiKey:      tt.fields.apiKey,
				secretKey:   tt.fields.secretKey,
				accessToken: tt.fields.accessToken,
				httpClient:  tt.fields.httpClient,
				Model:       tt.fields.Model,
				ModelPath:   tt.fields.ModelPath,
			}
			if got := c.buildURL(tt.args.modelpath); got != tt.want {
				t.Errorf("buildURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
