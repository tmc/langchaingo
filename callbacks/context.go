package callbacks

import "context"

type contextKeyType int

var _callbackHandlerKey = contextKeyType(0)

func GetHandlerFromContext(ctx context.Context) Handler {
	handler := ctx.Value(_callbackHandlerKey)
	if t, ok := handler.(Handler); ok {
		return t
	}
	return nil
}

func SetHandlerInContext(ctx context.Context, handler Handler) context.Context {
	return context.WithValue(ctx, _callbackHandlerKey, handler)
}