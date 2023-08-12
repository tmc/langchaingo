package scraper

type Options func(*Scraper)

func WithMaxDepth(maxDepth int) Options {
	return func(o *Scraper) {
		o.MaxDepth = maxDepth
	}
}

func WithParallelsNum(parallels int) Options {
	return func(o *Scraper) {
		o.Parallels = parallels
	}
}

func WithDelay(delay int64) Options {
	return func(o *Scraper) {
		o.Delay = delay
	}
}

func WithAsync(async bool) Options {
	return func(o *Scraper) {
		o.Async = async
	}
}

func WithNewBlacklist(blacklist []string) Options {
	return func(o *Scraper) {
		o.Blacklist = blacklist
	}
}

func WithBlacklist(blacklist []string) Options {
	return func(o *Scraper) {
		o.Blacklist = append(o.Blacklist, blacklist...)
	}
}
