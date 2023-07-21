package tools

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"golang.org/x/net/context"
)

var ErrInvalidLink = errors.New("error scraping website url")

type Scraper struct {
	MaxDepth  int
	Parallels int
	Delay     int64
}

var _ Tool = Scraper{}

func New(maxDepth ...int) (*Scraper, error) {
	depth := 1
	if len(maxDepth) > 0 {
		depth = maxDepth[0]
	}

	return &Scraper{
		MaxDepth:  depth,
		Parallels: 2,
		Delay:     5,
	}, nil
}

func (scraper Scraper) Name() string {
	return "Web Scraper"
}

func (scraper Scraper) Description() string {
	return `
		Web Scraper will scan a url and return the content of the web page.
		Input should be a working url.
	`
}

func (scraper Scraper) Call(ctx context.Context, input string) (string, error) {
	// Check that input is a valid URL
	_, err := url.ParseRequestURI(input)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidLink, err)
	}

	c := colly.NewCollector(
		colly.MaxDepth(scraper.MaxDepth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		Parallelism: scraper.Parallels,
		Delay:       time.Duration(scraper.Delay) * time.Second,
	})

	var siteData strings.Builder

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
		if ctx.Err() != nil {
			r.Abort()
		}
	})

	c.OnHTML("html", func(e *colly.HTMLElement) {
		siteData.WriteString("\n\nPage URL: " + e.Request.URL.String())

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

		siteData.WriteString("\nLinks:")
		e.ForEach("a", func(_ int, el *colly.HTMLElement) {
			siteData.WriteString("\n" + el.Attr("href"))
		})

		siteData.WriteString("\nContent:")
		e.ForEach("p", func(_ int, el *colly.HTMLElement) {
			siteData.WriteString("\n" + el.Text)
		})
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		link = e.Request.AbsoluteURL(link)
		if err := c.Visit(link); err != nil {
			siteData.WriteString(fmt.Sprintf("\nError following link %s: %v", link, err))
		}
	})

	err = c.Visit(input)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidLink, err)
	}

	// Check for context cancellation before waiting
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		c.Wait()
	}

	return siteData.String(), nil
}
