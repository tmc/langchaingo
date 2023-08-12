package scraper

type Options func(*Scraper)

// WithMaxDepth sets the maximum depth for the Scraper.
//
// maxDepth: the maximum depth to set.
// Returns: an Options function.
func WithMaxDepth(maxDepth int) Options {
	return func(o *Scraper) {
		o.MaxDepth = maxDepth
	}
}

// WithParallelsNum sets the number of maximum allowed concurrent
// requests of the matching domains
//
// parallels: the number of parallels to set.
// Returns: the updated Scraper options.
func WithParallelsNum(parallels int) Options {
	return func(o *Scraper) {
		o.Parallels = parallels
	}
}

// WithDelay creates an Options function that sets the delay of a Scraper.
//
// The delay parameter specifies the amount of time in milliseconds that
// the Scraper should wait between requests.
//
// delay: the delay to set.
// Returns: an Options function.
func WithDelay(delay int64) Options {
	return func(o *Scraper) {
		o.Delay = delay
	}
}

// WithAsync sets the async option for the Scraper.
//
// async: The boolean value indicating if the scraper should run asynchronously.
// Returns a function that sets the async option for the Scraper.
func WithAsync(async bool) Options {
	return func(o *Scraper) {
		o.Async = async
	}
}

// WithNewBlacklist creates an Options function that replaces
// the list of url endpoints to be excluded from the scraping,
// with a new list.
//
// Default value:
//
//	[]string{
//		"login",
//		"signup",
//		"signin",
//		"register",
//		"logout",
//		"download",
//		"redirect",
//	},
//
// blacklist: slice of strings with url endpoints to be excluded from the scraping.
// Returns: an Options function.
func WithNewBlacklist(blacklist []string) Options {
	return func(o *Scraper) {
		o.Blacklist = blacklist
	}
}

// WithBlacklist creates an Options function that appends
// the url endpoints to be excluded from the scraping,
// to the current list
//
// Default value:
//
//	[]string{
//		"login",
//		"signup",
//		"signin",
//		"register",
//		"logout",
//		"download",
//		"redirect",
//	},
//
// blacklist: slice of strings with url endpoints to be excluded from the scraping.
// Returns: an Options function.
func WithBlacklist(blacklist []string) Options {
	return func(o *Scraper) {
		o.Blacklist = append(o.Blacklist, blacklist...)
	}
}
