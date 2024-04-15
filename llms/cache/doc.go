// Package cache provides a generic wrapper that adds caching to a `llms.Model`. Responses are
// cached under a key calculated based on the provided messages and options. Different cache
// backends can be used when creating the wrapper.
package cache
