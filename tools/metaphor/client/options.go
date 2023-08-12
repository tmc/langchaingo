package client

type ClientOptions func(*MetaphorClient)

// WithNumResults sets the number of expected search results.
//
// Parameters:
//   - numResults: The desired number of results.
//
// Returns: a ClientOptions function that updates the numResults field of the RequestBody struct.
func WithNumResults(numResults int) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.NumResults = numResults
	}
}

// WithIncludeDomains sets the includeDomains field of the RequestBody.
// List of domains to include in the search. If specified, results will
// only come from these domains. Only one of includeDomains and excludeDomains
// should be specified.
//
// Parameters:
// - includeDomains: a slice of strings representing the domains to include.
//
// Returns: a ClientOptions function that updates the includeDomains field of the RequestBody struct.
func WithIncludeDomains(includeDomains []string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.IncludeDomains = includeDomains
	}
}

// WithExcludeDomains sets the ExcludeDomains field of the client's RequestBody.
// List of domains to exclude in the search. If specified, results will only come
// from these domains. Only one of includeDomains and excludeDomains should be specified.
//
// Parameters:
// - excludeDomains: an array of strings representing the domains to be excluded.
//
// Returns: a ClientOptions function that updates the excludeDomains field of the RequestBody struct.
func WithExcludeDomains(excludeDomains []string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.ExcludeDomains = excludeDomains
	}
}

// WithStartCrawlDate sets the start crawl date for the client options.
// If startCrawlDate is specified, results will only include links that
// were crawled after startCrawlDate.
// Must be specified in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
//
// Parameters:
// - startCrawlDate: the start date for the crawl
//
// Returns: a ClientOptions function that updates the startCrawlDate field of the RequestBody struct.
func WithStartCrawlDate(startCrawlDate string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.StartCrawlDate = startCrawlDate
	}
}

// WithEndCrawlDate sets the end crawl date for the client options.
// If endCrawlDate is specified, results will only include links that
// were crawled before endCrawlDate.
// Must be specified in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ)
//
// Parameters:
// - endCrawlDate: the end crawl date to be set.
//
// Returns: a ClientOptions function that updates the endCrawlDate field of the RequestBody struct.
func WithEndCrawlDate(endCrawlDate string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.EndCrawlDate = endCrawlDate
	}
}

// WithStartPublishedDate sets the start published date for the client options.
// If specified, only links with a published date after startPublishedDate will
// be returned.
// Must be specified in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ).
//
// Parameters:
// - startPublishedDate: a string representing the start published date.
//
// Returns: a ClientOptions function that updates the startPublishedDate field of the RequestBody struct.
func WithStartPublishedDate(startPublishedDate string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.StartPublishedDate = startPublishedDate
	}
}

// WithEndPublishedDate sets the end published date for the client options.
// If specified, only links with a published date before endPublishedDate will
// be returned.
// Must be specified in ISO 8601 format (YYYY-MM-DDTHH:MM:SSZ).
//
// Parameters:
// - endPublishedDate: the end published date to be set.
//
// Returns: a ClientOptions function that updates the endPublishedDate field of the RequestBody struct.
func WithEndPublishedDate(endPublishedDate string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.EndPublishedDate = endPublishedDate
	}
}

// WithAutoprompt sets the value of the UseAutoprompt field in the RequestBody.
// If true, your query will be converted to a Metaphor query. Latency will be much higher.
// Default: false
//
// Parameters:
// - useAutoprompt: a boolean value indicating whether to use autoprompt or not.
//
// Returns: a ClientOptions function that updates the useAutoprompt field of the RequestBody struct.
func WithAutoprompt(useAutoprompt bool) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.UseAutoprompt = useAutoprompt
	}
}

// WithType sets the search type for the client.
// Type of search, 'keyword' or 'neural'.
// Default: neural
//
// Parameters:
// - searchType: the type of search to be performed.
//
// Returns: a ClientOptions function that updates the type field of the RequestBody struct.
func WithType(searchType string) ClientOptions {
	return func(client *MetaphorClient) {
		client.RequestBody.Type = searchType
	}
}
