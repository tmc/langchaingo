package zapier

import (
	"context"

	"github.com/tmc/langchaingo/tools"
	"github.com/tmc/langchaingo/tools/zapier/internal"
)

type ToolkitOpts struct {
	// User OAuth Access Token for Zapier NLA Takes Precedents over APIKey.
	AccessToken string
	// API Key for Zapier NLA.
	APIKey string
	// Customer User-Agent if one isn't passed Defaults to "LangChainGo/X.X.X".
	UserAgent string
	// Base URL for Zapier NLA API.
	ZapierNLABaseURL string
}

/*
Toolkit gets all the Zapier NLA Tools configured for the account.

Full docs here: https://nla.zapier.com/start/

Note: this wrapper currently only implemented the `api_key` auth method for testing
and server-side production use cases (using the developer's connected accounts on
Zapier.com)

For use-cases where LangChain + Zapier NLA is powering a user-facing application, and
LangChain needs access to the end-user's connected accounts on Zapier.com, you'll need
to use oauth. Review the full docs above and reach out to nla@zapier.com for
developer support.
*/
func Toolkit(ctx context.Context, opts ToolkitOpts) ([]tools.Tool, error) {
	c, err := internal.NewClient(internal.ClientOptions{
		APIKey:           opts.APIKey,
		AccessToken:      opts.AccessToken,
		UserAgent:        opts.UserAgent,
		ZapierNLABaseURL: opts.ZapierNLABaseURL,
	})
	if err != nil {
		return nil, err
	}

	listResponse, err := c.List(ctx)
	if err != nil {
		return nil, err
	}

	tools := make([]tools.Tool, len(listResponse))

	for i, result := range listResponse {
		tool, err := New(ToolOptions{
			Name:        result.Description,
			ActionID:    result.ID,
			Params:      result.Params,
			UserAgent:   opts.UserAgent,
			APIKey:      opts.APIKey,
			AccessToken: opts.AccessToken,
			Client:      c,
		})
		if err != nil {
			return nil, err
		}

		tools[i] = tool
	}

	return tools, nil
}
