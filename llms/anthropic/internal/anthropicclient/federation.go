package anthropicclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	jwtBearerGrantType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
	tokenExchangePath  = "/oauth/token"
	tokenRefreshSkew   = 60 * time.Second
)

// ErrFederationIncomplete is returned when a federation config omits a field the
// token exchange requires.
var ErrFederationIncomplete = errors.New("federation config is incomplete")

// AssertionSource yields the identity token to exchange. It is called for every
// exchange, not once per client.
type AssertionSource func(ctx context.Context) (string, error)

// AssertionFromFile reads the identity token from path on every exchange.
func AssertionFromFile(path string) AssertionSource {
	return func(context.Context) (string, error) {
		body, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read identity token: %w", err)
		}
		return strings.TrimSpace(string(body)), nil
	}
}

// AssertionFromString yields a fixed identity token.
func AssertionFromString(assertion string) AssertionSource {
	return func(context.Context) (string, error) { return assertion, nil }
}

// FederationConfig names the federation rule a workload exchanges its identity
// token under, and where that token comes from.
type FederationConfig struct {
	RuleID           string
	OrganizationID   string
	ServiceAccountID string
	// WorkspaceID scopes the minted token; the literal "default" names the
	// organization's default workspace. Omitted from the exchange when empty.
	WorkspaceID string
	Assertion   AssertionSource
}

func (cfg FederationConfig) validate() error {
	missing := []string{}
	if cfg.RuleID == "" {
		missing = append(missing, "rule id")
	}
	if cfg.OrganizationID == "" {
		missing = append(missing, "organization id")
	}
	if cfg.ServiceAccountID == "" {
		missing = append(missing, "service account id")
	}
	if cfg.Assertion == nil {
		missing = append(missing, "assertion source")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing %s", ErrFederationIncomplete, strings.Join(missing, ", "))
	}
	return nil
}

type federatedAuth struct {
	cfg    FederationConfig
	client *Client
	now    func() time.Time

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (f *federatedAuth) authorize(ctx context.Context, req *http.Request) error {
	token, err := f.accessToken(ctx)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (f *federatedAuth) accessToken(ctx context.Context) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.token != "" && f.now().Add(tokenRefreshSkew).Before(f.expiresAt) {
		return f.token, nil
	}

	assertion, err := f.cfg.Assertion(ctx)
	if err != nil {
		return "", err
	}
	if assertion == "" {
		return "", fmt.Errorf("%w: assertion source returned an empty token", ErrFederationIncomplete)
	}

	token, lifetime, err := f.exchange(ctx, assertion)
	if err != nil {
		return "", err
	}

	f.token = token
	f.expiresAt = f.now().Add(lifetime)
	return token, nil
}

type tokenExchangePayload struct {
	GrantType        string `json:"grant_type"`
	Assertion        string `json:"assertion"`
	FederationRuleID string `json:"federation_rule_id"`
	OrganizationID   string `json:"organization_id"`
	ServiceAccountID string `json:"service_account_id"`
	WorkspaceID      string `json:"workspace_id,omitempty"`
}

type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func (f *federatedAuth) exchange(ctx context.Context, assertion string) (string, time.Duration, error) {
	payload, err := json.Marshal(tokenExchangePayload{
		GrantType:        jwtBearerGrantType,
		Assertion:        assertion,
		FederationRuleID: f.cfg.RuleID,
		OrganizationID:   f.cfg.OrganizationID,
		ServiceAccountID: f.cfg.ServiceAccountID,
		WorkspaceID:      f.cfg.WorkspaceID,
	})
	if err != nil {
		return "", 0, fmt.Errorf("marshal token exchange: %w", err)
	}

	base := f.client.baseURL
	if base == "" {
		base = DefaultBaseURL
	}

	url := base + tokenExchangePath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", 0, fmt.Errorf("create token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("send token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("token exchange failed: %w", f.client.decodeError(resp))
	}

	var exchanged tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&exchanged); err != nil {
		return "", 0, fmt.Errorf("decode token exchange: %w", err)
	}
	if exchanged.AccessToken == "" {
		return "", 0, errors.New("token exchange returned an empty access token")
	}

	return exchanged.AccessToken, time.Duration(exchanged.ExpiresIn) * time.Second, nil
}
