package main

import (
	"context"
	"net/http"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

// To run this example you need to run a Tika server and then set the address
// on the TikaURL constant. The easiest way of running a Tika server is by
// using Docker:
//
// $ docker run -d -p 9998:9998 apache/tika
//
// Tika will be listening on http://localhost:9998, you then just need to ajust
// the TikaURL constant.

const TikaURL = "http://localhost:9998"

func main() {
	resp, err := http.Get("https://www.golang-book.com/public/pdf/gobook.pdf")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	splitter := textsplitter.NewRecursiveCharacter()
	tika := documentloaders.NewTika(TikaURL, resp.Body)
	docs, err := tika.LoadAndSplit(context.Background(), splitter)
	if err != nil {
		panic(err)
	}

	_ = docs
}
