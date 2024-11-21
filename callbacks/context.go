package callbacks

import "context"

type contextKeyType int

// nolint: gochecknoglobals
var callbackHandlerKey = contextKeyType(0)

func CallbackHandler(ctx context.Context) Handler {
	handler := ctx.Value(callbackHandlerKey)
	if t, ok := handler.(Handler); ok {
		return t
	}
	return nil
}

func WithCallback(ctx context.Context, handler Handler) context.Context {
	return context.WithValue(ctx, callbackHandlerKey, handler)
}
