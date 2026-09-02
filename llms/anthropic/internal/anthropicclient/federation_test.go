package anthropicclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type federationStub struct {
	mu sync.Mutex

	exchanges     int
	assertions    []string
	authHeaders   []string
	apiKeyHeaders []string
	payloads      []tokenExchangePayload

	lifetime       int
	exchangeStatus int
	server         *httptest.Server
}

func newFederationStub(t *testing.T) *federationStub {
	t.Helper()

	s := &federationStub{lifetime: 3600, exchangeStatus: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()

		var payload tokenExchangePayload
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		s.exchanges++
		s.assertions = append(s.assertions, payload.Assertion)
		s.payloads = append(s.payloads, payload)

		if s.exchangeStatus != http.StatusOK {
			w.WriteHeader(s.exchangeStatus)
			_, _ = w.Write([]byte(`{"error":{"type":"authentication_error","message":"Authentication failed"}}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenExchangeResponse{
			AccessToken: "sk-ant-oat01-" + payload.Assertion,
			TokenType:   "Bearer",
			ExpiresIn:   s.lifetime,
			Scope:       "workspace:inference",
		})
	})
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.authHeaders = append(s.authHeaders, r.Header.Get("Authorization"))
		s.apiKeyHeaders = append(s.apiKeyHeaders, r.Header.Get("x-api-key"))
		s.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	})

	s.server = httptest.NewServer(mux)
	t.Cleanup(s.server.Close)
	return s
}

func (s *federationStub) counts() (exchanges int, assertions, auth, apiKey []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exchanges, append([]string(nil), s.assertions...),
		append([]string(nil), s.authHeaders...), append([]string(nil), s.apiKeyHeaders...)
}

func (s *federationStub) client(t *testing.T, cfg FederationConfig) (*Client, *federatedAuth) {
	t.Helper()

	c, err := New("", "claude-sonnet-5", s.server.URL+"/v1", WithFederation(cfg))
	require.NoError(t, err)

	auth, ok := c.auth.(*federatedAuth)
	require.True(t, ok, "federation must install the federated authorizer")
	return c, auth
}

func fixedAssertion(value string) AssertionSource {
	return AssertionFromString(value)
}

func TestFederationSendsBearerInsteadOfAPIKey(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	c, _ := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "11111111-1111-1111-1111-111111111111",
		ServiceAccountID: "svac_account",
		WorkspaceID:      "wrkspc_one",
		Assertion:        fixedAssertion("jwt-one"),
	})

	resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	exchanges, _, auth, apiKey := stub.counts()
	require.Equal(t, 1, exchanges)
	require.Equal(t, []string{"Bearer sk-ant-oat01-jwt-one"}, auth)
	require.Equal(t, []string{""}, apiKey, "a federated request must not carry the static key header")

	stub.mu.Lock()
	sent := stub.payloads[0]
	stub.mu.Unlock()
	require.Equal(t, jwtBearerGrantType, sent.GrantType)
	require.Equal(t, "fdrl_rule", sent.FederationRuleID)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", sent.OrganizationID)
	require.Equal(t, "svac_account", sent.ServiceAccountID)
	require.Equal(t, "wrkspc_one", sent.WorkspaceID)
}

func TestFederationOmitsWorkspaceWhenUnset(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	c, _ := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "org",
		ServiceAccountID: "svac_account",
		Assertion:        fixedAssertion("jwt"),
	})

	resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	stub.mu.Lock()
	raw, err := json.Marshal(stub.payloads[0])
	stub.mu.Unlock()
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(raw, &body))
	_, present := body["workspace_id"]
	require.False(t, present, "an unset workspace must stay off the exchange body")
}

func TestFederationReusesTokenUntilItNearsExpiry(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	c, auth := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "org",
		ServiceAccountID: "svac_account",
		Assertion:        fixedAssertion("jwt"),
	})

	now := time.Now()
	auth.now = func() time.Time { return now }

	for range 3 {
		resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}

	exchanges, _, _, _ := stub.counts()
	require.Equal(t, 1, exchanges, "a live token must be reused rather than exchanged again")
}

func TestFederationRefreshesWithAFreshAssertion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	require.NoError(t, os.WriteFile(path, []byte("jwt-first\n"), 0o600))

	stub := newFederationStub(t)
	c, auth := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "org",
		ServiceAccountID: "svac_account",
		Assertion:        AssertionFromFile(path),
	})

	now := time.Now()
	auth.now = func() time.Time { return now }

	resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	require.NoError(t, os.WriteFile(path, []byte("jwt-second\n"), 0o600))
	now = now.Add(time.Duration(stub.lifetime) * time.Second)

	resp, err = c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	exchanges, assertions, auths, _ := stub.counts()
	require.Equal(t, 2, exchanges, "an expired token must be exchanged again")
	require.Equal(t, []string{"jwt-first", "jwt-second"}, assertions,
		"each exchange must re-read the assertion source")
	require.Equal(t, []string{"Bearer sk-ant-oat01-jwt-first", "Bearer sk-ant-oat01-jwt-second"}, auths)
}

func TestFederationExchangesOnceUnderConcurrency(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	c, _ := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "org",
		ServiceAccountID: "svac_account",
		Assertion:        fixedAssertion("jwt"),
	})

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
	}
	wg.Wait()

	exchanges, _, auths, _ := stub.counts()
	require.Equal(t, 1, exchanges, "concurrent callers must share one exchange, not race two")
	require.Len(t, auths, 8)
}

func TestFederationReportsExchangeFailure(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	stub.exchangeStatus = http.StatusUnauthorized

	c, _ := stub.client(t, FederationConfig{
		RuleID:           "fdrl_rule",
		OrganizationID:   "org",
		ServiceAccountID: "svac_account",
		Assertion:        fixedAssertion("jwt"),
	})

	_, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token exchange failed")
	require.Contains(t, err.Error(), "Authentication failed")

	_, _, auths, _ := stub.counts()
	require.Empty(t, auths, "a failed exchange must not send the downstream request")
}

func TestFederationRejectsIncompleteConfig(t *testing.T) {
	t.Parallel()

	_, err := New("", "claude-sonnet-5", "https://example.invalid/v1", WithFederation(FederationConfig{
		RuleID: "fdrl_rule",
	}))
	require.ErrorIs(t, err, ErrFederationIncomplete)
	require.Contains(t, err.Error(), "organization id")
	require.Contains(t, err.Error(), "service account id")
	require.Contains(t, err.Error(), "assertion source")
}

func TestStaticKeyStillSendsAPIKeyHeader(t *testing.T) {
	t.Parallel()

	stub := newFederationStub(t)
	c, err := New("sk-ant-static", "claude-sonnet-5", stub.server.URL+"/v1")
	require.NoError(t, err)

	resp, err := c.request(context.Background(), http.MethodGet, "/models", http.NoBody, nil)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	exchanges, _, auths, apiKey := stub.counts()
	require.Zero(t, exchanges, "a static key must not trigger a token exchange")
	require.Equal(t, []string{"sk-ant-static"}, apiKey)
	require.Equal(t, []string{""}, auths)
}
