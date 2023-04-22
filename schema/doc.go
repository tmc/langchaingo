// Package schema implements a shared core set of data types for use in
// langchaingo.
//
// The primary interface through which end users interact with LLMs is a chat
// interface. For this reason, some model providers have started providing
// access to the underlying API in a way that expects chat messages. These
// messages have a content field (which is usually text) and are associated
// with a user (or role). Right now the supported users are System, Human, AI,
// and a generic/arbitrary user.
package schema
