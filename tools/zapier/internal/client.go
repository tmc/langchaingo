package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type listResponse struct {
	Results           []ListResult `json:"results"`
	ConfigurationLink string       `json:"configuration_link"`
}

type ListResult struct {
	ID          string            `json:"id"`
	OperationID string            `json:"operation_id"`
	Description string            `json:"description"`
	Params      map[string]string `json:"params"`
}

type executionResponse struct {
	ActionUsed string      `json:"action_used"`
	Result     interface{} `json:"result"`
	Status     string      `json:"status"`
	Error      string      `json:"error"`
}

const (
	zapierNLABaseURL = "https://nla.zapier.com/api/v1"
)

// Client for interacting with Zapier NLA API.
type Client struct {
	client *http.Client
}

// Transport RoundTripper for Zapier NLA API which adds on Correct Headers.
type Transport struct {
	RoundTripper http.RoundTripper
	apiKey       string
	accessToken  string
	UserAgent    string
}

// ClientOptions for configuring a new Client.
type ClientOptions struct {
	// User OAuth Access Token for Zapier NLA Takes Precedents over APIKey.
	AccessToken string
	// API Key for Zapier NLA.
	APIKey string
	// Customer User-Agent if one isn't passed Defaults to "LangChainGo/X.X.X".
	UserAgent string
	// Base URL for Zapier NLA API.
	ZapierNLABaseURL string
}

func (cOpts *ClientOptions) Validate() error {
	if cOpts.APIKey == "" {
		cOpts.APIKey = os.Getenv("ZAPIER_NLA_API_KEY")
	}

	if cOpts.APIKey == "" && cOpts.AccessToken == "" {
		return NoCredentialsError{}
	}

	if cOpts.UserAgent == "" {
		cOpts.UserAgent = "LangChainGo/0.0.1"
	}

	if cOpts.ZapierNLABaseURL == "" {
		cOpts.ZapierNLABaseURL = zapierNLABaseURL
	}

	return nil
}

/*
Client for Zapier NLA.

Full docs here: https://nla.zapier.com/start/

This Client supports both API Key and OAuth Credential auth methods. API Key
is the fastest way to get started using this wrapper.

Call this Client with either `APIKey` or
`AccessToken` arguments, or set the `ZAPIER_NLA_API_KEY`
environment variable. If both arguments are set, the Access Token will take
precedence.

For use-cases where LangChain + Zapier NLA is powering a user-facing application,
and LangChain needs access to the end-user's connected accounts on Zapier.com,
you'll need to use OAuth. Review the full docs above to learn how to create
your own provider and generate credentials.
*/
func NewClient(opts ClientOptions) (*Client, error) {
	err := opts.Validate()
	if err != nil {
		return nil, err
	}

	return &Client{
		client: &http.Client{
			Transport: &Transport{
				RoundTripper: http.DefaultTransport,
				apiKey:       opts.APIKey,
				accessToken:  opts.AccessToken,
				UserAgent:    opts.UserAgent,
			},
		},
	}, nil
}

/*
List returns a list of all exposed (enabled) actions associated with
current user (associated with the set api_key). Change your exposed
actions here: https://nla.zapier.com/demo/start/

The return list can be empty if no actions exposed. Else will contain
a list of ListResult structs, which look like this:

	[
		ListResult{
			"ID": str,
			"OperationID": str,
			"Description": str,
			"Params": Dict[str, str]
		}
	]

`Params` will always contain an `instructions` key, the only required
param. All others optional and if provided will override any AI guesses
(see "understanding the AI guessing flow" here:
https://nla.zapier.com/api/v1/docs).
*/
func (c *Client) List(ctx context.Context) ([]ListResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, formatListURL(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	lr := listResponse{}

	err = json.Unmarshal(b, &lr)
	if err != nil {
		return nil, err
	}

	return lr.Results, nil
}

/*
Execute an action that is identified by action_id, must be exposed
(enabled) by the current user (associated with the set api_key). Change
your exposed actions here: https://nla.zapier.com/demo/start/

The return JSON is guaranteed to be less than ~500 words (350
tokens) making it safe to inject into the prompt of another LLM
call.
*/
func (c *Client) Execute(
	ctx context.Context,
	actionID string,
	input string,
	params map[string]string,
) (interface{}, error) {
	body, err := createPayload(input, params)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, formatExecuteURL(actionID), body)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	executionResponse := executionResponse{}

	err = json.Unmarshal(b, &executionResponse)
	if err != nil {
		return "", err
	}

	return executionResponse.Result, nil
}

/*
ExecuteAsString is a convenience wrapper around Execute that returns a string response.
*/
func (c *Client) ExecuteAsString(
	ctx context.Context,
	actionID string,
	input string,
	params map[string]string,
) (string, error) {
	r, err := c.Execute(ctx, actionID, input, params)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", r), nil
}

func formatListURL() string {
	return fmt.Sprintf("%s/exposed", zapierNLABaseURL)
}

func formatExecuteURL(actionID string) string {
	return fmt.Sprintf("%s/exposed/%s/execute/", zapierNLABaseURL, actionID)
}

func createPayload(input string, params map[string]string) (*bytes.Buffer, error) {
	params["instructions"] = input

	b, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.createHeaders(req)
	return t.RoundTripper.RoundTrip(req)
}

func (t *Transport) createAuthHeader(req *http.Request) {
	if t.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.accessToken)
	} else {
		req.Header.Set("X-API-Key", t.apiKey)
	}
}

func (t *Transport) createHeaders(req *http.Request) {
	t.createAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", t.UserAgent)
}
