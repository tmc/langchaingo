package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vxcontrol/langchaingo/llms"
)

func federatedTestServer(t *testing.T) (*httptest.Server, func() (int, []string)) {
	t.Helper()

	var (
		mu        sync.Mutex
		exchanges int
		auth      []string
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		exchanges++
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"sk-ant-oat01-minted","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/v1/messages", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		auth = append(auth, r.Header.Get("Authorization")+"|"+r.Header.Get("x-api-key"))
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_1", "type": "message", "role": "assistant", "model": "claude-sonnet-5",
			"content":     []map[string]any{{"type": "text", "text": "ok"}},
			"stop_reason": "end_turn",
			"usage":       map[string]any{"input_tokens": 1, "output_tokens": 1},
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	return srv, func() (int, []string) {
		mu.Lock()
		defer mu.Unlock()
		return exchanges, append([]string(nil), auth...)
	}
}

func TestFederationReachesTheWireWithoutAnAPIKey(t *testing.T) {
	t.Setenv(tokenEnvVarName, "")

	srv, observed := federatedTestServer(t)

	llm, err := New(
		WithBaseURL(srv.URL+"/v1"),
		WithModel("claude-sonnet-5"),
		WithFederation(FederationConfig{
			RuleID:           "fdrl_rule",
			OrganizationID:   "org",
			ServiceAccountID: "svac_account",
			Assertion:        AssertionFromString("jwt"),
		}),
	)
	require.NoError(t, err, "federation must stand in for the API key")

	_, err = llm.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, "hi"),
	}, llms.WithMaxTokens(16))
	require.NoError(t, err)

	exchanges, auth := observed()
	require.Equal(t, 1, exchanges)
	require.Equal(t, []string{"Bearer sk-ant-oat01-minted|"}, auth)
}

func TestMissingTokenStillFailsWithoutFederation(t *testing.T) {
	t.Setenv(tokenEnvVarName, "")

	_, err := New(WithBaseURL("https://example.invalid/v1"))
	require.ErrorIs(t, err, ErrMissingToken)
}
