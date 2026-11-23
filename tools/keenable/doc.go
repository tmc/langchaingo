// Package keenable contains an implementation of the tool interface for the
// Keenable web search API (https://keenable.ai).
//
// Keenable works without an API key: by default the tool calls the keyless,
// per-IP rate-limited public endpoint. Set an API key with WithAPIKey or the
// KEENABLE_API_KEY environment variable to use the keyed endpoint instead.
// Each result carries a long excerpt of the page, so agents get page text
// rather than a short SERP blurb.
package keenable
