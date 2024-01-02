package util

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenNewSubCtx(t *testing.T) {
	t.Parallel()
	{
		// empty context
		newCtx := GenNewSubCtx(context.Background())
		runID, parentRunID := GetRunIDParentIDFromCtx(newCtx)
		assert.NotEqualValues(t, len(runID), 0)
		assert.Nil(t, parentRunID)
	}

	{
		// context with run id
		ctx := GenNewSubCtx(context.Background())
		ctxRunID, _ := GetRunIDParentIDFromCtx(ctx)
		subCtx := GenNewSubCtx(ctx)
		subCtxRunID, subCtxParentRunID := GetRunIDParentIDFromCtx(subCtx)
		assert.EqualValues(t, ctxRunID, *subCtxParentRunID)
		assert.NotEqualValues(t, subCtxRunID, ctxRunID)
	}

	{
		// context with run id and parent id
		ctxWithRunID := GenNewSubCtx(context.Background())
		ctxWithRunIDAndParentID := GenNewSubCtx(ctxWithRunID)
		runID, parentRunID := GetRunIDParentIDFromCtx(ctxWithRunIDAndParentID)

		subCtx := GenNewSubCtx(ctxWithRunIDAndParentID)
		subCtxRunID, subCtxParentRunID := GetRunIDParentIDFromCtx(subCtx)
		assert.NotEqualValues(t, len(*parentRunID), 0)
		assert.EqualValues(t, runID, *subCtxParentRunID)
		assert.NotEqualValues(t, len(subCtxRunID), 0)
	}
}

func TestGetRunIDParentIDFromCtx(t *testing.T) {
	t.Parallel()
	{
		// get run id and parent id with empty context
		runID, parentRunID := GetRunIDParentIDFromCtx(context.Background())
		assert.EqualValues(t, runID, "")
		assert.Nil(t, parentRunID)
	}
}
