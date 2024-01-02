package util

import (
	"context"

	"github.com/google/uuid"
)

type contextKey int

const (
	runIDKey contextKey = iota
	parentRunIDKey
	// ...
)

// GetRunIDParentIDFromCtx get Run ID and its Parent Run ID by context.
func GetRunIDParentIDFromCtx(ctx context.Context) (string, *string) {
	runIDValue := ctx.Value(runIDKey)
	if runIDValue == nil {
		return "", nil
	}
	var ok bool
	runID, ok := runIDValue.(string)
	if !ok {
		return "", nil
	}

	parentRunIDValue := ctx.Value(parentRunIDKey)
	if parentRunIDValue == nil {
		return runID, nil
	}
	parentRunIDStr, ok := parentRunIDValue.(string)
	if !ok {
		return runID, nil
	}
	return runID, &parentRunIDStr
}

// nolint: lll
// GenNewSubCtx creates a new sub-context from a parent context. It retrieves the current run ID from the parent context,
// sets it as the parent run ID in the new sub-context, and then generates a new unique run ID for the sub-context itself.
func GenNewSubCtx(ctx context.Context) context.Context {
	currentRunIDValue := ctx.Value(runIDKey)
	if currentRunID, ok := currentRunIDValue.(string); currentRunIDValue != nil && ok {
		ctx = context.WithValue(ctx, parentRunIDKey, currentRunID)
	}
	ctx = context.WithValue(ctx, runIDKey, uuid.NewString())
	return ctx
}
