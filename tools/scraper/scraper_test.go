package scraper_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/tools/scraper"
)

func TestScraperConcurrentPageContent(t *testing.T) {
	const pages = 8

	for _, tc := range []struct {
		name        string
		async       bool
		parallelism int
	}{
		{name: "default_parallelism", async: true, parallelism: scraper.DefualtParallels},
		{name: "four_parallel_requests", async: true, parallelism: 4},
		{name: "synchronous", async: false, parallelism: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()

			content := make([]string, pages)
			for i := range content {
				content[i] = strings.Repeat(fmt.Sprintf("page-%d-content ", i), 2048)
			}

			ready := make(chan struct{})
			var arrivals atomic.Int32
			var release sync.Once
			t.Cleanup(func() { release.Do(func() { close(ready) }) })

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				if r.URL.Path == "/" {
					fmt.Fprint(w, "<html><body>")
					for i := range pages {
						fmt.Fprintf(w, `<a href="/page/%d">page %d</a>`, i, i)
						fmt.Fprintf(w, `<a href="/page/%d">duplicate %d</a>`, i, i)
					}
					fmt.Fprint(w, "</body></html>")
					return
				}

				i, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/page/"))
				if err != nil || i < 0 || i >= pages {
					http.NotFound(w, r)
					return
				}
				if tc.async {
					if arrivals.Add(1) == int32(tc.parallelism) {
						release.Do(func() { close(ready) })
					}
					select {
					case <-ready:
					case <-ctx.Done():
						return
					}
				}
				fmt.Fprintf(w, "<html><head><title>title-%d</title></head><body><h1>heading-%d</h1><p>%s</p></body></html>", i, i, content[i])
			}))
			defer server.Close()

			opts := []scraper.Options{
				scraper.WithDelay(0),
				scraper.WithAsync(tc.async),
				scraper.WithMaxDepth(2),
			}
			if tc.parallelism != scraper.DefualtParallels {
				opts = append(opts, scraper.WithParallelsNum(tc.parallelism))
			}
			tool, err := scraper.New(opts...)
			require.NoError(t, err)

			result, err := tool.Call(ctx, server.URL+"/")
			require.NoError(t, err)
			for i, text := range content {
				page := fmt.Sprintf("\n\nPage URL: %s/page/%d\nPage Title: title-%d\nHeaders:\nheading-%d\nContent:\n%s", server.URL, i, i, i, text)
				require.Equal(t, 1, strings.Count(result, page), "page %d content was lost or interleaved", i)
			}
			if tc.async {
				require.Contains(t, result, "Error following link /page/0:")
			}
		})
	}
}
