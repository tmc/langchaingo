// Package outcrawl contains implementations of the tool interface for the
// Outcrawl web API (https://outcrawl.ai): a web search tool that returns
// ranked hits, and a scrape tool that returns the main content of a page as
// markdown.
//
// Both tools read the API key from the OUTCRAWL_API_KEY environment variable
// unless WithAPIKey is used.
//
// Example usage:
//
//	search, err := outcrawl.NewSearch()
//	if err != nil {
//		return err
//	}
//	scrape, err := outcrawl.NewScrape()
//	if err != nil {
//		return err
//	}
//
//	agentTools := []tools.Tool{search, scrape}
//	executor := agents.NewExecutor(agents.NewOneShotAgent(llm, agentTools))
package outcrawl
