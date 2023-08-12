package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/tmc/langchaingo/tools"
)

const (
	DefualtMaxDept   = 1
	DefualtParallels = 2
	DefualtDelay     = 3
	DefualtAsync     = true
)

var ErrScrapingFailed = errors.New("scraper could not read URL, or scraping is not allowed for provided URL")

type Scraper struct {
	MaxDepth  int
	Parallels int
	Delay     int64
	Blacklist []string
	Async     bool
}

var _ tools.Tool = Scraper{}

// NewScraper initializes a new Scraper object with the given maximum depth.
//
// It takes an optional parameter `maxDepth` which specifies the maximum depth to scrape.
// The default value of `maxDepth` is 1 if not provided.
//
// It returns a pointer to a Scraper object and an error.
func New() (*Scraper, error) {
	scraper := &Scraper{
		MaxDepth:  DefualtMaxDept,
		Parallels: DefualtParallels,
		Delay:     int64(DefualtDelay),
		Async:     DefualtAsync,
		Blacklist: []string{
			"login",
			"signup",
			"signin",
			"register",
			"logout",
			"download",
			"redirect",
		},
	}

	return scraper, nil
}

// Name returns the name of the Scraper.
//
// No parameters.
// Returns a string.
func (s Scraper) Name() string {
	return "Web Scraper"
}

// Description returns the description of the Go function.
//
// There are no parameters.
// It returns a string.
func (s Scraper) Description() string {
	return `
		Web Scraper will scan a url and return the content of the web page.
		Input should be a working url.
	`
}

// Call scrapes a website and returns the site data.
//
// The function takes a context.Context object for managing the execution
// context and a string input representing the URL of the website to be scraped.
// It returns a string containing the scraped data and an error if any.
//
//nolint:all
func (s Scraper) Call(ctx context.Context, input string) (string, error) {
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	c := colly.NewCollector(
		colly.MaxDepth(s.MaxDepth),
		colly.Async(s.Async),
	)

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.Parallels,
		Delay:       time.Duration(s.Delay) * time.Second,
	})

	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	var siteData strings.Builder
	homePageLinks := make(map[string]bool)
	scrapedLinks := make(map[string]bool)

	c.OnRequest(func(r *colly.Request) {
		if ctx.Err() != nil {
			r.Abort()
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		currentURL := e.Request.URL.String()

		// Only process the page if it hasn't been visited yet
		if !scrapedLinks[currentURL] {
			scrapedLinks[currentURL] = true

			siteData.WriteString("\n\nPage URL: " + currentURL)

			title := e.ChildText("title")
			if title != "" {
				siteData.WriteString("\nPage Title: " + title)
			}

			description := e.ChildAttr("meta[name=description]", "content")
			if description != "" {
				siteData.WriteString("\nPage Description: " + description)
			}

			siteData.WriteString("\nHeaders:")
			e.ForEach("h1, h2, h3, h4, h5, h6", func(_ int, el *colly.HTMLElement) {
				siteData.WriteString("\n" + el.Text)
			})

			siteData.WriteString("\nContent:")
			e.ForEach("p", func(_ int, el *colly.HTMLElement) {
				siteData.WriteString("\n" + el.Text)
			})

			if currentURL == input {
				e.ForEach("a", func(_ int, el *colly.HTMLElement) {
					link := el.Attr("href")
					if link != "" && !homePageLinks[link] {
						homePageLinks[link] = true
						siteData.WriteString("\nLink: " + link)
					}
				})
			}
		}
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		absoluteLink := e.Request.AbsoluteURL(link)

		// Parse the link to get the hostname
		u, err := url.Parse(absoluteLink)
		if err != nil {
			// Handle the error appropriately
			return
		}

		// Check if the link's hostname matches the current request's hostname
		if u.Hostname() != e.Request.URL.Hostname() {
			return
		}

		// Check for redundant pages
		for _, item := range s.Blacklist {
			if strings.Contains(u.Path, item) {
				return
			}
		}

		// Normalize the path to treat '/' and '/index.html' as the same path
		if u.Path == "/index.html" || u.Path == "" {
			u.Path = "/"
		}

		// Only visit the page if it hasn't been visited yet
		if !scrapedLinks[u.String()] {
			err := c.Visit(u.String())
			if err != nil {
				siteData.WriteString(fmt.Sprintf("\nError following link %s: %v", link, err))
			}
		}
	})

	err = c.Visit(input)
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrScrapingFailed, err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		c.Wait()
	}

	// Append all scraped links
	siteData.WriteString("\n\nScraped Links:")
	for link := range scrapedLinks {
		siteData.WriteString("\n" + link)
	}

	return siteData.String(), nil
}
