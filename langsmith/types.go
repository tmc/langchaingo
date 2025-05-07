package langsmith

import "time"

type KVMap map[string]any

func valueIfSetOtherwiseNil[T comparable](v T) *T {
	var empty T
	if v == empty {
		return nil
	}

	return &v
}

func timeToMillisecondsPtr(t time.Time) *int64 {
	if t.IsZero() {
		return nil
	}

	return ptr(t.UnixMilli())
}

func ptr[T any](v T) *T {
	return &v
}
